package onedrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
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
		url.URL{},
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
func (handler Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	defer file.Close()
	return handler.Client.Upload(ctx, dst, int(size), file)
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return handler.Client.BatchDelete(ctx, files)
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	var (
		thumbSize = [2]uint{400, 300}
		ok        = false
	)
	if thumbSize, ok = ctx.Value(fsctx.ThumbSizeCtx).([2]uint); !ok {
		return nil, errors.New("无法获取缩略图尺寸设置")
	}

	res, err := handler.Client.GetThumbURL(ctx, path, thumbSize[0], thumbSize[1])
	if err != nil {
		// 如果出现异常，就清空文件的pic_info
		if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
			file.UpdatePicInfo("")
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
	baseURL url.URL,
	ttl int64,
	isDownload bool,
	speed int,
) (string, error) {
	cacheKey := fmt.Sprintf("onedrive_source_%d_%s", handler.Policy.ID, path)
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		cacheKey = fmt.Sprintf("onedrive_source_file_%d_%d", file.UpdatedAt.Unix(), file.ID)
		// 如果是永久链接，则返回签名后的中转外链
		if ttl == 0 {
			signedURI, err := auth.SignURI(
				auth.General,
				fmt.Sprintf("/api/v3/file/source/%d/%s", file.ID, file.Name),
				ttl,
			)
			if err != nil {
				return "", err
			}
			return baseURL.ResolveReference(signedURI).String(), nil
		}

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
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {

	// 读取上下文中生成的存储路径和文件大小
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}
	fileSize, ok := ctx.Value(fsctx.FileSizeCtx).(uint64)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取文件大小")
	}

	// 如果小于4MB，则由服务端中转
	if fileSize <= SmallFileSize {
		return serializer.UploadCredential{}, nil
	}

	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/onedrive/finish/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	uploadURL, err := handler.Client.CreateUploadSession(ctx, savePath, WithConflictBehavior("fail"))
	if err != nil {
		return serializer.UploadCredential{}, err
	}

	// 监控回调及上传
	go handler.Client.MonitorUpload(uploadURL, key, savePath, fileSize, TTL)

	return serializer.UploadCredential{
		Policy: uploadURL,
		Token:  apiURL.String(),
	}, nil
}
