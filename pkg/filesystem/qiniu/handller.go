package qiniu

import (
	"context"
	"errors"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"io"
	"net/url"
)

// Handler 本地策略适配器
type Handler struct {
	Policy *model.Policy
}

// Get 获取文件
func (handler Handler) Get(ctx context.Context, path string) (response.RSCloser, error) {
	return nil, errors.New("未实现")
}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	return errors.New("未实现")
	//// 凭证生成
	//putPolicy := storage.PutPolicy{
	//	Scope: "cloudrevetest",
	//}
	//mac := auth.New("YNzTBBpDUq4EEiFV0-vyJCZCJ0LvUEI0_WvxtEXE", "Clm9d9M2CH7pZ8vm049ZlGZStQxrRQVRTjU_T5_0")
	//upToken := putPolicy.UploadToken(mac)
	//
	//cfg := storage.Config{}
	//// 空间对应的机房
	//cfg.Zone = &storage.ZoneHuadong
	//formUploader := storage.NewFormUploader(&cfg)
	//ret := storage.PutRet{}
	//putExtra := storage.PutExtra{
	//	Params: map[string]string{},
	//}
	//
	//defer file.Close()
	//
	//err := formUploader.Put(ctx, &ret, upToken, dst, file, int64(size), &putExtra)
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//fmt.Println(ret.Key, ret.Hash)
	//return nil
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
	return "", errors.New("未实现")
}

// Token 获取上传策略和认证Token
func (handler Handler) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
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
func (handler Handler) getUploadCredential(ctx context.Context, policy storage.PutPolicy, TTL int64) (serializer.UploadCredential, error) {
	policy.Expires = uint64(TTL)
	mac := qbox.NewMac(handler.Policy.AccessKey, handler.Policy.SecretKey)
	upToken := policy.UploadToken(mac)

	return serializer.UploadCredential{
		Token: upToken,
	}, nil
}
