package qiniu

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
)

// Driver 本地策略适配器
type Driver struct {
	Policy *model.Policy
}

// List 列出给定路径下的文件
func (handler Driver) List(ctx context.Context, base string, recursive bool) ([]response.Object, error) {
	base = strings.TrimPrefix(base, "/")
	if base != "" {
		base += "/"
	}

	var (
		delimiter string
		marker    string
		objects   []storage.ListItem
		commons   []string
	)
	if !recursive {
		delimiter = "/"
	}

	mac := qbox.NewMac(handler.Policy.AccessKey, handler.Policy.SecretKey)
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	for {
		entries, folders, nextMarker, hashNext, err := bucketManager.ListFiles(
			handler.Policy.BucketName,
			base, delimiter, marker, 1000)
		if err != nil {
			return nil, err
		}
		objects = append(objects, entries...)
		commons = append(commons, folders...)
		if !hashNext {
			break
		}
		marker = nextMarker
	}

	// 处理列取结果
	res := make([]response.Object, 0, len(objects)+len(commons))
	// 处理目录
	for _, object := range commons {
		rel, err := filepath.Rel(base, object)
		if err != nil {
			continue
		}
		res = append(res, response.Object{
			Name:         path.Base(object),
			RelativePath: filepath.ToSlash(rel),
			Size:         0,
			IsDir:        true,
			LastModify:   time.Now(),
		})
	}
	// 处理文件
	for _, object := range objects {
		rel, err := filepath.Rel(base, object.Key)
		if err != nil {
			continue
		}
		res = append(res, response.Object{
			Name:         path.Base(object.Key),
			Source:       object.Key,
			RelativePath: filepath.ToSlash(rel),
			Size:         uint64(object.Fsize),
			IsDir:        false,
			LastModify:   time.Unix(object.PutTime/10000000, 0),
		})
	}

	return res, nil
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 给文件名加上随机参数以强制拉取
	path = fmt.Sprintf("%s?v=%d", path, time.Now().UnixNano())

	// 获取文件源地址
	downloadURL, err := handler.Source(
		ctx,
		path,
		url.URL{},
		int64(model.GetIntSetting("preview_timeout", 60)),
		false,
		0,
	)
	if err != nil {
		return nil, err
	}

	// 获取文件数据流
	client := request.HTTPClient{}
	resp, err := client.Request(
		"GET",
		downloadURL,
		nil,
		request.WithContext(ctx),
		request.WithHeader(
			http.Header{"Cache-Control": {"no-cache", "no-store", "must-revalidate"}},
		),
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

	// 凭证有效期
	credentialTTL := model.GetIntSetting("upload_credential_timeout", 3600)

	// 生成上传策略
	putPolicy := storage.PutPolicy{
		// 指定为覆盖策略
		Scope:        fmt.Sprintf("%s:%s", handler.Policy.BucketName, dst),
		SaveKey:      dst,
		ForceSaveKey: true,
		FsizeLimit:   int64(size),
	}
	// 是否开启了MIMEType限制
	if handler.Policy.OptionsSerialized.MimeType != "" {
		putPolicy.MimeLimit = handler.Policy.OptionsSerialized.MimeType
	}

	// 生成上传凭证
	token, err := handler.getUploadCredential(ctx, putPolicy, int64(credentialTTL))
	if err != nil {
		return err
	}

	// 创建上传表单
	cfg := storage.Config{}
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{},
	}

	// 开始上传
	err = formUploader.Put(ctx, &ret, token.Token, dst, file, int64(size), &putExtra)
	if err != nil {
		return err
	}

	return nil
}

// Delete 删除一个或多个文件，
// 返回未删除的文件
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	// TODO 大于一千个文件需要分批发送
	deleteOps := make([]string, 0, len(files))
	for _, key := range files {
		deleteOps = append(deleteOps, storage.URIDelete(handler.Policy.BucketName, key))
	}

	mac := qbox.NewMac(handler.Policy.AccessKey, handler.Policy.SecretKey)
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	rets, err := bucketManager.Batch(deleteOps)

	// 处理删除结果
	if err != nil {
		failed := make([]string, 0, len(rets))
		for k, ret := range rets {
			if ret.Code != 200 {
				failed = append(failed, files[k])
			}
		}
		return failed, errors.New("删除失败")
	}

	return []string{}, nil
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

	path = fmt.Sprintf("%s?imageView2/1/w/%d/h/%d", path, thumbSize[0], thumbSize[1])
	return &response.ContentResponse{
		Redirect: true,
		URL: handler.signSourceURL(
			ctx,
			path,
			int64(model.GetIntSetting("preview_timeout", 60)),
		),
	}, nil
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
	// 尝试从上下文获取文件名
	fileName := ""
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	// 加入下载相关设置
	if isDownload {
		path = path + "?attname=" + url.PathEscape(fileName)
	}

	// 取得原始文件地址
	return handler.signSourceURL(ctx, path, ttl), nil
}

func (handler Driver) signSourceURL(ctx context.Context, path string, ttl int64) string {
	var sourceURL string
	if handler.Policy.IsPrivate {
		mac := qbox.NewMac(handler.Policy.AccessKey, handler.Policy.SecretKey)
		deadline := time.Now().Add(time.Second * time.Duration(ttl)).Unix()
		sourceURL = storage.MakePrivateURL(mac, handler.Policy.BaseURL, path, deadline)
	} else {
		sourceURL = storage.MakePublicURL(handler.Policy.BaseURL, path)
	}
	return sourceURL
}

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/qiniu/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 读取上下文中生成的存储路径
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}

	// 创建上传策略
	putPolicy := storage.PutPolicy{
		Scope:            handler.Policy.BucketName,
		CallbackURL:      apiURL.String(),
		CallbackBody:     `{"name":"$(fname)","source_name":"$(key)","size":$(fsize),"pic_info":"$(imageInfo.width),$(imageInfo.height)"}`,
		CallbackBodyType: "application/json",
		SaveKey:          savePath,
		ForceSaveKey:     true,
		FsizeLimit:       int64(handler.Policy.MaxSize),
	}
	// 是否开启了MIMEType限制
	if handler.Policy.OptionsSerialized.MimeType != "" {
		putPolicy.MimeLimit = handler.Policy.OptionsSerialized.MimeType
	}

	return handler.getUploadCredential(ctx, putPolicy, TTL)
}

// getUploadCredential 签名上传策略
func (handler Driver) getUploadCredential(ctx context.Context, policy storage.PutPolicy, TTL int64) (serializer.UploadCredential, error) {
	policy.Expires = uint64(TTL)
	mac := qbox.NewMac(handler.Policy.AccessKey, handler.Policy.SecretKey)
	upToken := policy.UploadToken(mac)

	return serializer.UploadCredential{
		Token: upToken,
	}, nil
}
