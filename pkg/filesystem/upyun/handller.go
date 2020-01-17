package upyun

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"io"
	"net/url"
	"strings"
	"time"
)

// UploadPolicy 又拍云上传策略
type UploadPolicy struct {
	Bucket             string `json:"bucket"`
	SaveKey            string `json:"save-key"`
	Expiration         int64  `json:"expiration"`
	CallbackURL        string `json:"notify-url"`
	ContentLength      uint64 `json:"content-length"`
	ContentLengthRange string `json:"content-length-range"`
	AllowFileType      string `json:"allow-file-type,omitempty"`
}

// Driver 又拍云策略适配器
type Driver struct {
	Policy *model.Policy
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
	// 读取上下文中生成的存储路径和文件大小
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}
	fileSize, ok := ctx.Value(fsctx.FileSizeCtx).(uint64)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取文件大小")
	}

	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/upyun/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 上传策略
	putPolicy := UploadPolicy{
		Bucket: handler.Policy.BucketName,
		// TODO escape
		SaveKey:            savePath,
		Expiration:         time.Now().Add(time.Duration(TTL) * time.Second).Unix(),
		CallbackURL:        apiURL.String(),
		ContentLength:      fileSize,
		ContentLengthRange: fmt.Sprintf("0,%d", fileSize),
		AllowFileType:      strings.Join(handler.Policy.OptionsSerialized.FileType, ","),
	}

	// 生成上传凭证
	return handler.getUploadCredential(ctx, putPolicy)
}

func (handler Driver) getUploadCredential(ctx context.Context, policy UploadPolicy) (serializer.UploadCredential, error) {
	// 生成上传策略
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	policyEncoded := base64.StdEncoding.EncodeToString(policyJSON)

	// 生成签名
	password := fmt.Sprintf("%x", md5.Sum([]byte(handler.Policy.SecretKey)))
	mac := hmac.New(sha1.New, []byte(password))
	elements := []string{"POST", "/" + handler.Policy.BucketName, policyEncoded}
	value := strings.Join(elements, "&")
	mac.Write([]byte(value))
	signStr := base64.StdEncoding.EncodeToString((mac.Sum(nil)))

	return serializer.UploadCredential{
		Policy: policyEncoded,
		Token:  fmt.Sprintf("UPYUN %s:%s", handler.Policy.AccessKey, signStr),
	}, nil
}
