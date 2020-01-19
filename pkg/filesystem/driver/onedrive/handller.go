package onedrive

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
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
	return errors.New("未实现")
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return []string{}, errors.New("未实现")
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	return nil, errors.New("未实现")
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

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	err := handler.Client.UpdateCredential(ctx)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	return serializer.UploadCredential{
		Policy: handler.Client.Credential.AccessToken,
	}, nil
	//res,err := handler.Client.ObtainToken(ctx,WithCode("M2e92c4a9-de12-cdda-9cf4-e01f67272831"))
	//if err != nil{
	//	return serializer.UploadCredential{},err
	//}
	//return serializer.UploadCredential{
	//	Policy:res.RefreshToken,
	//}, nil
}
