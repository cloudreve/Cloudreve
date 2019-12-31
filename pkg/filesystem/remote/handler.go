package remote

// TODO 测试
import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Handler 远程存储策略适配器
type Handler struct {
	Policy *model.Policy
}

// Get 获取文件内容
func (handler Handler) Get(ctx context.Context, path string) (response.RSCloser, error) {

	return nil, nil
}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	return errors.New("远程策略不支持此上传方式")
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Handler) Delete(ctx context.Context, files []string) ([]string, error) {
	return []string{}, nil
}

// Thumb 获取文件缩略图
func (handler Handler) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	return nil, errors.New("未实现")
}

// Source 获取外链URL
// TODO 测试
func (handler Handler) Source(
	ctx context.Context,
	path string,
	baseURL url.URL,
	ttl int64,
	isDownload bool,
) (string, error) {
	file, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return "", errors.New("无法获取文件记录上下文")
	}

	serverURL, err := url.Parse(handler.Policy.Server)
	if err != nil {
		return "", errors.New("无法解析远程服务端地址")
	}

	var (
		expires    int64
		signedURI  *url.URL
		controller = "/api/v3/slave/download"
	)
	if !isDownload {
		controller = "/api/v3/slave/source"
	}
	if ttl > 0 {
		expires = time.Now().Unix() + ttl
	}

	// 签名下载地址
	sourcePath := base64.RawURLEncoding.EncodeToString([]byte(file.SourceName))
	authInstance := auth.HMACAuth{SecretKey: []byte(handler.Policy.SecretKey)}
	signedURI, err = auth.SignURI(
		authInstance,
		fmt.Sprintf("%s/%s", controller, sourcePath),
		expires,
	)

	if err != nil {
		return "", serializer.NewError(serializer.CodeEncryptError, "无法对URL进行签名", err)
	}

	finalURL := serverURL.ResolveReference(signedURI).String()
	return finalURL, nil

}

// Token 获取上传策略和认证Token
func (handler Handler) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/remote/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 生成上传策略
	policy := serializer.UploadPolicy{
		SavePath:         handler.Policy.DirNameRule,
		FileName:         handler.Policy.FileNameRule,
		AutoRename:       handler.Policy.AutoRename,
		MaxSize:          handler.Policy.MaxSize,
		AllowedExtension: handler.Policy.OptionsSerialized.FileType,
		CallbackURL:      apiURL.String(),
	}
	policyEncoded, err := policy.EncodeUploadPolicy()
	if err != nil {
		return serializer.UploadCredential{}, err
	}

	// 签名上传策略
	uploadRequest, _ := http.NewRequest("POST", "/api/v3/slave/upload", nil)
	uploadRequest.Header = map[string][]string{
		"X-Policy": {policyEncoded},
	}
	remoteAuth := auth.HMACAuth{SecretKey: []byte(handler.Policy.SecretKey)}
	auth.SignRequest(remoteAuth, uploadRequest, time.Now().Unix()+TTL)

	if credential, ok := uploadRequest.Header["Authorization"]; ok && len(credential) == 1 {
		return serializer.UploadCredential{
			Token:  credential[0],
			Policy: policyEncoded,
		}, nil
	}
	return serializer.UploadCredential{}, errors.New("无法签名上传策略")

}
