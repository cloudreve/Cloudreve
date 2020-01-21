package onedrive

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"io"
	"net/url"
)

// Driver OneDrive 适配器
type Driver struct {
	Policy *model.Policy
	Client *Client
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	return nil, errors.New("未实现")
}

// Put 将文件流保存到指定目录
func (handler Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	defer file.Close()
	_, err := handler.Client.PutFile(ctx, dst, file)
	return err
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return handler.Client.Delete(ctx, files)
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
	return "", errors.New("未实现")
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
