package onedrive

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Driver OneDrive 适配器
type Driver struct {
	Policy     *model.Policy
	Client     *Client
	HTTPClient request.Client
}

// NewDriver 从存储策略初始化新的Driver实例
func NewDriver(policy *model.Policy) (driver.Handler, error) {
	client, err := NewClient(policy)
	if policy.OptionsSerialized.ChunkSize == 0 {
		policy.OptionsSerialized.ChunkSize = 50 << 20 // 50MB
	}

	return Driver{
		Policy:     policy,
		Client:     client,
		HTTPClient: request.NewClient(),
	}, err
}

// List 列取项目
func (handler Driver) List(ctx context.Context, base string, recursive bool) ([]response.Object, error) {
	base = strings.TrimPrefix(base, "/")
	// 列取子项目
	objects, _ := handler.Client.ListChildren(ctx, base)

	// 获取真实的列取起始根目录
	rootPath := base
	if realBase, ok := ctx.Value(fsctx.PathCtx).(string); ok {
		rootPath = realBase
	} else {
		ctx = context.WithValue(ctx, fsctx.PathCtx, base)
	}

	// 整理结果
	res := make([]response.Object, 0, len(objects))
	for _, object := range objects {
		source := path.Join(base, object.Name)
		rel, err := filepath.Rel(rootPath, source)
		if err != nil {
			continue
		}
		res = append(res, response.Object{
			Name:         object.Name,
			RelativePath: filepath.ToSlash(rel),
			Source:       source,
			Size:         object.Size,
			IsDir:        object.Folder != nil,
			LastModify:   time.Now(),
		})
	}

	// 递归列取子目录
	if recursive {
		for _, object := range objects {
			if object.Folder != nil {
				sub, _ := handler.List(ctx, path.Join(base, object.Name), recursive)
				res = append(res, sub...)
			}
		}
	}

	return res, nil
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 获取文件源地址
	downloadURL, err := handler.Source(
		ctx,
		path,
		60,
		false,
		0,
	)
	if err != nil {
		return nil, err
	}

	// 获取文件数据流
	resp, err := handler.HTTPClient.Request(
		"GET",
		downloadURL,
		nil,
		request.WithContext(ctx),
		request.WithTimeout(time.Duration(0)),
	).CheckHTTPResponse(200).GetRSCloser()
	if err != nil {
		return nil, err
	}

	resp.SetFirstFakeChunk()

	// 尝试自主获取文件大小
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		resp.SetContentLength(int64(file.Size))
	}

	return resp, nil
}

// Put 将文件流保存到指定目录
func (handler Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	defer file.Close()

	return handler.Client.Upload(ctx, file)
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return handler.Client.BatchDelete(ctx, files)
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	var (
		thumbSize = [2]uint{400, 300}
		ok        = false
	)
	if thumbSize, ok = ctx.Value(fsctx.ThumbSizeCtx).([2]uint); !ok {
		return nil, errors.New("failed to get thumbnail size")
	}

	res, err := handler.Client.GetThumbURL(ctx, file.SourceName, thumbSize[0], thumbSize[1])
	if err != nil {
		var apiErr *RespError
		if errors.As(err, &apiErr); err == ErrThumbSizeNotFound || (apiErr != nil && apiErr.APIError.Code == notFoundError) {
			// OneDrive cannot generate thumbnail for this file
			return nil, driver.ErrorThumbNotSupported
		}
	}

	return &response.ContentResponse{
		Redirect: true,
		URL:      res,
	}, err
}

// Source 获取外链URL
func (handler Driver) Source(
	ctx context.Context,
	path string,
	ttl int64,
	isDownload bool,
	speed int,
) (string, error) {
	cacheKey := fmt.Sprintf("onedrive_source_%d_%s", handler.Policy.ID, path)
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		cacheKey = fmt.Sprintf("onedrive_source_file_%d_%d", file.UpdatedAt.Unix(), file.ID)
	}

	// 尝试从缓存中查找
	if cachedURL, ok := cache.Get(cacheKey); ok {
		return handler.replaceSourceHost(cachedURL.(string))
	}

	// 缓存不存在，重新获取
	res, err := handler.Client.Meta(ctx, "", path)
	if err == nil {
		// 写入新的缓存
		cache.Set(
			cacheKey,
			res.DownloadURL,
			model.GetIntSetting("onedrive_source_timeout", 1800),
		)
		return handler.replaceSourceHost(res.DownloadURL)
	}
	return "", err
}

func (handler Driver) replaceSourceHost(origin string) (string, error) {
	if handler.Policy.OptionsSerialized.OdProxy != "" {
		source, err := url.Parse(origin)
		if err != nil {
			return "", err
		}

		cdn, err := url.Parse(handler.Policy.OptionsSerialized.OdProxy)
		if err != nil {
			return "", err
		}

		// 替换反代地址
		source.Scheme = cdn.Scheme
		source.Host = cdn.Host
		return source.String(), nil
	}

	return origin, nil
}

// Token 获取上传会话URL
func (handler Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	fileInfo := file.Info()

	uploadURL, err := handler.Client.CreateUploadSession(ctx, fileInfo.SavePath, WithConflictBehavior("fail"))
	if err != nil {
		return nil, err
	}

	// 监控回调及上传
	go handler.Client.MonitorUpload(uploadURL, uploadSession.Key, fileInfo.SavePath, fileInfo.Size, ttl)

	uploadSession.UploadURL = uploadURL
	return &serializer.UploadCredential{
		SessionID:  uploadSession.Key,
		ChunkSize:  handler.Policy.OptionsSerialized.ChunkSize,
		UploadURLs: []string{uploadURL},
	}, nil
}

// 取消上传凭证
func (handler Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	return handler.Client.DeleteUploadSession(ctx, uploadSession.UploadURL)
}
