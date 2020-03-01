package cos

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
	"github.com/HFO4/cloudreve/pkg/request"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/google/go-querystring/query"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
	"time"
)

// UploadPolicy 腾讯云COS上传策略
type UploadPolicy struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

// MetaData 文件元信息
type MetaData struct {
	Size        uint64
	CallbackKey string
	CallbackURL string
}

type urlOption struct {
	Speed              int    `url:"x-cos-traffic-limit,omitempty"`
	ContentDescription string `url:"response-content-disposition,omitempty"`
}

// Driver 腾讯云COS适配器模板
type Driver struct {
	Policy     *model.Policy
	Client     *cossdk.Client
	HTTPClient request.Client
}

// CORS 创建跨域策略
func (handler Driver) CORS() error {
	_, err := handler.Client.Bucket.PutCORS(context.Background(), &cossdk.BucketPutCORSOptions{
		Rules: []cossdk.BucketCORSRule{{
			AllowedMethods: []string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
				"HEAD",
			},
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"*"},
			MaxAgeSeconds:  3600,
			ExposeHeaders:  []string{},
		}},
	})

	return err
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
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
	opt := &cossdk.ObjectPutOptions{}
	_, err := handler.Client.Object.Put(ctx, dst, file, opt)
	return err
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	obs := []cossdk.Object{}
	for _, v := range files {
		obs = append(obs, cossdk.Object{Key: v})
	}
	opt := &cossdk.ObjectDeleteMultiOptions{
		Objects: obs,
		Quiet:   true,
	}

	res, _, err := handler.Client.Object.DeleteMulti(context.Background(), opt)
	if err != nil {
		return files, err
	}

	// 整理删除结果
	failed := make([]string, 0, len(files))
	for _, v := range res.Errors {
		failed = append(failed, v.Key)
	}

	if len(failed) == 0 {
		return failed, nil
	}

	return failed, errors.New("删除失败")
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
	// 尝试从上下文获取文件名
	fileName := ""
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	// 添加各项设置
	options := urlOption{}
	if speed > 0 {
		if speed < 819200 {
			speed = 819200
		}
		if speed > 838860800 {
			speed = 838860800
		}
		options.Speed = speed
	}
	if isDownload {
		options.ContentDescription = "attachment; filename=\"" + url.PathEscape(fileName) + "\""
	}

	return handler.signSourceURL(ctx, path, ttl, &options)
}

func (handler Driver) signSourceURL(ctx context.Context, path string, ttl int64, options *urlOption) (string, error) {
	cdnURL, err := url.Parse(handler.Policy.BaseURL)
	if err != nil {
		return "", err
	}

	// 公有空间不需要签名
	if !handler.Policy.IsPrivate {
		file, err := url.Parse(path)
		if err != nil {
			return "", err
		}

		// 非签名URL不支持设置响应header
		options.ContentDescription = ""

		optionQuery, err := query.Values(*options)
		if err != nil {
			return "", err
		}
		file.RawQuery = optionQuery.Encode()
		sourceURL := cdnURL.ResolveReference(file)

		return sourceURL.String(), nil
	}

	presignedURL, err := handler.Client.Object.GetPresignedURL(ctx, http.MethodGet, path,
		handler.Policy.AccessKey, handler.Policy.SecretKey, time.Duration(ttl)*time.Second, options)
	if err != nil {
		return "", err
	}

	// 将最终生成的签名URL域名换成用户自定义的加速域名（如果有）
	presignedURL.Host = cdnURL.Host
	presignedURL.Scheme = cdnURL.Scheme

	return presignedURL.String(), nil
}

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 读取上下文中生成的存储路径
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}

	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/cos/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI).String()

	// 上传策略
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(TTL) * time.Second)
	keyTime := fmt.Sprintf("%d;%d", startTime.Unix(), endTime.Unix())
	postPolicy := UploadPolicy{
		Expiration: endTime.UTC().Format(time.RFC3339),
		Conditions: []interface{}{
			map[string]string{"bucket": handler.Policy.BucketName},
			map[string]string{"$key": savePath},
			map[string]string{"x-cos-meta-callback": apiURL},
			map[string]string{"x-cos-meta-key": key},
			map[string]string{"q-sign-algorithm": "sha1"},
			map[string]string{"q-ak": handler.Policy.AccessKey},
			map[string]string{"q-sign-time": keyTime},
		},
	}

	if handler.Policy.MaxSize > 0 {
		postPolicy.Conditions = append(postPolicy.Conditions,
			[]interface{}{"content-length-range", 0, handler.Policy.MaxSize})
	}

	res, err := handler.getUploadCredential(ctx, postPolicy, keyTime)
	if err == nil {
		res.Callback = apiURL
		res.Key = key
	}

	return res, err

}

// Meta 获取文件信息
func (handler Driver) Meta(ctx context.Context, path string) (*MetaData, error) {
	res, err := handler.Client.Object.Head(ctx, path, &cossdk.ObjectHeadOptions{})
	if err != nil {
		return nil, err
	}
	return &MetaData{
		Size:        uint64(res.ContentLength),
		CallbackKey: res.Header.Get("x-cos-meta-key"),
		CallbackURL: res.Header.Get("x-cos-meta-callback"),
	}, nil
}

func (handler Driver) getUploadCredential(ctx context.Context, policy UploadPolicy, keyTime string) (serializer.UploadCredential, error) {
	// 读取上下文中生成的存储路径
	savePath, ok := ctx.Value(fsctx.SavePathCtx).(string)
	if !ok {
		return serializer.UploadCredential{}, errors.New("无法获取存储路径")
	}

	// 编码上传策略
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	policyEncoded := base64.StdEncoding.EncodeToString(policyJSON)

	// 签名上传策略
	hmacSign := hmac.New(sha1.New, []byte(handler.Policy.SecretKey))
	_, err = io.WriteString(hmacSign, keyTime)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	signKey := fmt.Sprintf("%x", hmacSign.Sum(nil))

	sha1Sign := sha1.New()
	_, err = sha1Sign.Write(policyJSON)
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	stringToSign := fmt.Sprintf("%x", sha1Sign.Sum(nil))

	// 最终签名
	hmacFinalSign := hmac.New(sha1.New, []byte(signKey))
	_, err = hmacFinalSign.Write([]byte(stringToSign))
	if err != nil {
		return serializer.UploadCredential{}, err
	}
	signature := hmacFinalSign.Sum(nil)

	return serializer.UploadCredential{
		Policy:    policyEncoded,
		Path:      savePath,
		AccessKey: handler.Policy.AccessKey,
		Token:     fmt.Sprintf("%x", signature),
		KeyTime:   keyTime,
	}, nil
}
