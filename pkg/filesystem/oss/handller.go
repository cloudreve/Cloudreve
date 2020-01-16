package oss

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"net/url"
	"path"
	"time"
)

// UploadPolicy 阿里云OSS上传策略
type UploadPolicy struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

// CallbackPolicy 回调策略
type CallbackPolicy struct {
	CallbackURL      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

// Handler 阿里云OSS策略适配器
type Handler struct {
	Policy *model.Policy
	client *oss.Client
	bucket *oss.Bucket
}

// InitOSSClient 初始化OSS鉴权客户端
func (handler *Handler) InitOSSClient() error {
	if handler.Policy == nil {
		return errors.New("存储策略为空")
	}

	// 初始化客户端
	client, err := oss.New(handler.Policy.Server, handler.Policy.AccessKey, handler.Policy.SecretKey)
	if err != nil {
		return err
	}
	handler.client = client

	// 初始化存储桶
	bucket, err := client.Bucket(handler.Policy.BucketName)
	if err != nil {
		return err
	}
	handler.bucket = bucket

	return nil
}

// Get 获取文件
func (handler Handler) Get(ctx context.Context, path string) (response.RSCloser, error) {
	return nil, errors.New("未实现")
}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	return errors.New("未实现")
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Handler) Delete(ctx context.Context, files []string) ([]string, error) {
	return []string{}, errors.New("未实现")
}

// Thumb 获取文件缩略图
func (handler Handler) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	return nil, errors.New("未实现")
}

// Source 获取外链URL
func (handler Handler) Source(
	ctx context.Context,
	path string,
	baseURL url.URL,
	ttl int64,
	isDownload bool,
	speed int,
) (string, error) {
	// 初始化客户端
	if err := handler.InitOSSClient(); err != nil {
		return "", err
	}

	// 尝试从上下文获取文件名
	fileName := ""
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	// 添加各项设置
	var signOptions = make([]oss.Option, 0, 2)
	if isDownload {
		signOptions = append(signOptions, oss.ResponseContentDisposition("attachment; filename=\""+url.PathEscape(fileName)+"\""))
	}
	if speed > 0 {
		// OSS对速度值有范围限制
		if speed < 819200 {
			speed = 819200
		}
		if speed > 838860800 {
			speed = 838860800
		}
		signOptions = append(signOptions, oss.TrafficLimitParam(int64(speed)))
	}

	return handler.signSourceURL(ctx, path, ttl, signOptions)
}

func (handler Handler) signSourceURL(ctx context.Context, path string, ttl int64, options []oss.Option) (string, error) {
	signedURL, err := handler.bucket.SignURL(path, oss.HTTPGet, ttl, options...)
	if err != nil {
		return "", err
	}

	// 将最终生成的签名URL域名换成用户自定义的加速域名（如果有）
	finalURL, err := url.Parse(signedURL)
	if err != nil {
		return "", err
	}
	cdnURL, err := url.Parse(handler.Policy.BaseURL)
	if err != nil {
		return "", err
	}
	finalURL.Host = cdnURL.Host
	finalURL.Scheme = cdnURL.Scheme

	return finalURL.String(), nil
}

// Token 获取上传策略和认证Token
func (handler Handler) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 读取上下文中生成的存储路径
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}

	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/oss/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 回调策略
	callbackPolicy := CallbackPolicy{
		CallbackURL:      apiURL.String(),
		CallbackBody:     `{"name":${x:fname},"source_name":${object},"size":${size},"pic_info":"${imageInfo.width},${imageInfo.height}"}`,
		CallbackBodyType: "application/json",
	}

	// 上传策略
	postPolicy := UploadPolicy{
		Expiration: time.Now().UTC().Add(time.Duration(TTL) * time.Second).Format(time.RFC3339),
		Conditions: []interface{}{
			map[string]string{"bucket": handler.Policy.BucketName},
			[]string{"starts-with", "$key", path.Dir(savePath)},
			[]interface{}{"content-length-range", 0, handler.Policy.MaxSize},
		},
	}

	return handler.getUploadCredential(ctx, postPolicy, callbackPolicy, TTL)
}

func (handler Handler) getUploadCredential(ctx context.Context, policy UploadPolicy, callback CallbackPolicy, TTL int64) (serializer.UploadCredential, error) {
	// 读取上下文中生成的存储路径
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}

	// 处理回调策略
	callbackPolicyEncoded := ""
	if callback.CallbackURL != "" {
		callbackPolicyJSON, err := json.Marshal(callback)
		if err != nil {
			return serializer.UploadCredential{}, err
		}
		callbackPolicyEncoded = base64.StdEncoding.EncodeToString(callbackPolicyJSON)
		policy.Conditions = append(policy.Conditions, map[string]string{"callback": callbackPolicyEncoded})
	}

	// 编码上传策略
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	policyEncoded := base64.StdEncoding.EncodeToString(policyJSON)

	// 签名上传策略
	hmacSign := hmac.New(sha1.New, []byte(handler.Policy.SecretKey))
	_, err = io.WriteString(hmacSign, policyEncoded)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	signature := base64.StdEncoding.EncodeToString(hmacSign.Sum(nil))

	return serializer.UploadCredential{
		Policy:    fmt.Sprintf("%s:%s", callbackPolicyEncoded, policyEncoded),
		Path:      savePath,
		AccessKey: handler.Policy.AccessKey,
		Token:     signature,
	}, nil
}
