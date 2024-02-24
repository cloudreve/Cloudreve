package oss

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/HFO4/aliyun-oss-go-sdk/oss"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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

// Driver 阿里云OSS策略适配器
type Driver struct {
	Policy     *model.Policy
	client     *oss.Client
	bucket     *oss.Bucket
	HTTPClient request.Client
}

type key int

const (
	chunkRetrySleep = time.Duration(5) * time.Second

	// MultiPartUploadThreshold 服务端使用分片上传的阈值
	MultiPartUploadThreshold uint64 = 5 * (1 << 30) // 5GB
	// VersionID 文件版本标识
	VersionID key = iota
)

func NewDriver(policy *model.Policy) (*Driver, error) {
	if policy.OptionsSerialized.ChunkSize == 0 {
		policy.OptionsSerialized.ChunkSize = 25 << 20 // 25 MB
	}

	driver := &Driver{
		Policy:     policy,
		HTTPClient: request.NewClient(),
	}

	return driver, driver.InitOSSClient(false)
}

// CORS 创建跨域策略
func (handler *Driver) CORS() error {
	return handler.client.SetBucketCORS(handler.Policy.BucketName, []oss.CORSRule{
		{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
				"HEAD",
			},
			ExposeHeader:  []string{},
			AllowedHeader: []string{"*"},
			MaxAgeSeconds: 3600,
		},
	})
}

// InitOSSClient 初始化OSS鉴权客户端
func (handler *Driver) InitOSSClient(forceUsePublicEndpoint bool) error {
	if handler.Policy == nil {
		return errors.New("empty policy")
	}

	// 决定是否使用内网 Endpoint
	endpoint := handler.Policy.Server
	if handler.Policy.OptionsSerialized.ServerSideEndpoint != "" && !forceUsePublicEndpoint {
		endpoint = handler.Policy.OptionsSerialized.ServerSideEndpoint
	}

	// 初始化客户端
	client, err := oss.New(endpoint, handler.Policy.AccessKey, handler.Policy.SecretKey)
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

// List 列出OSS上的文件
func (handler *Driver) List(ctx context.Context, base string, recursive bool) ([]response.Object, error) {
	// 列取文件
	base = strings.TrimPrefix(base, "/")
	if base != "" {
		base += "/"
	}

	var (
		delimiter string
		marker    string
		objects   []oss.ObjectProperties
		commons   []string
	)
	if !recursive {
		delimiter = "/"
	}

	for {
		subRes, err := handler.bucket.ListObjects(oss.Marker(marker), oss.Prefix(base),
			oss.MaxKeys(1000), oss.Delimiter(delimiter))
		if err != nil {
			return nil, err
		}
		objects = append(objects, subRes.Objects...)
		commons = append(commons, subRes.CommonPrefixes...)
		marker = subRes.NextMarker
		if marker == "" {
			break
		}
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

		if strings.HasSuffix(object.Key, "/") {
			res = append(res, response.Object{
				Name:         path.Base(object.Key),
				RelativePath: filepath.ToSlash(rel),
				Size:         0,
				IsDir:        true,
				LastModify:   time.Now(),
			})
		} else {
			res = append(res, response.Object{
				Name:         path.Base(object.Key),
				Source:       object.Key,
				RelativePath: filepath.ToSlash(rel),
				Size:         uint64(object.Size),
				IsDir:        false,
				LastModify:   object.LastModified,
			})
		}
	}

	return res, nil
}

// Get 获取文件
func (handler *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 通过VersionID禁止缓存
	ctx = context.WithValue(ctx, VersionID, time.Now().UnixNano())

	// 尽可能使用私有 Endpoint
	ctx = context.WithValue(ctx, fsctx.ForceUsePublicEndpointCtx, false)

	// 获取文件源地址
	downloadURL, err := handler.Source(ctx, path, int64(model.GetIntSetting("preview_timeout", 60)), false, 0)
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
func (handler *Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	defer file.Close()
	fileInfo := file.Info()

	// 凭证有效期
	credentialTTL := model.GetIntSetting("upload_session_timeout", 3600)

	// 是否允许覆盖
	overwrite := fileInfo.Mode&fsctx.Overwrite == fsctx.Overwrite
	options := []oss.Option{
		oss.Expires(time.Now().Add(time.Duration(credentialTTL) * time.Second)),
		oss.ForbidOverWrite(!overwrite),
	}

	// 小文件直接上传
	if fileInfo.Size < MultiPartUploadThreshold {
		return handler.bucket.PutObject(fileInfo.SavePath, file, options...)
	}

	// 超过阈值时使用分片上传
	imur, err := handler.bucket.InitiateMultipartUpload(fileInfo.SavePath, options...)
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	chunks := chunk.NewChunkGroup(file, handler.Policy.OptionsSerialized.ChunkSize, &backoff.ConstantBackoff{
		Max:   model.GetIntSetting("chunk_retries", 5),
		Sleep: chunkRetrySleep,
	}, model.IsTrueVal(model.GetSettingByName("use_temp_chunk_buffer")))

	uploadFunc := func(current *chunk.ChunkGroup, content io.Reader) error {
		_, err := handler.bucket.UploadPart(imur, content, current.Length(), current.Index()+1)
		return err
	}

	for chunks.Next() {
		if err := chunks.Process(uploadFunc); err != nil {
			return fmt.Errorf("failed to upload chunk #%d: %w", chunks.Index(), err)
		}
	}

	_, err = handler.bucket.CompleteMultipartUpload(imur, oss.CompleteAll("yes"), oss.ForbidOverWrite(!overwrite))
	return err
}

// Delete 删除一个或多个文件，
// 返回未删除的文件
func (handler *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	// 删除文件
	delRes, err := handler.bucket.DeleteObjects(files)

	if err != nil {
		return files, err
	}

	// 统计未删除的文件
	failed := util.SliceDifference(files, delRes.DeletedObjects)
	if len(failed) > 0 {
		return failed, errors.New("failed to delete")
	}

	return []string{}, nil
}

// Thumb 获取文件缩略图
func (handler *Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	// quick check by extension name
	// https://help.aliyun.com/document_detail/183902.html
	supported := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "heic", "tiff", "avif"}
	if len(handler.Policy.OptionsSerialized.ThumbExts) > 0 {
		supported = handler.Policy.OptionsSerialized.ThumbExts
	}

	if !util.IsInExtensionList(supported, file.Name) || file.Size > (20<<(10*2)) {
		return nil, driver.ErrorThumbNotSupported
	}

	// 初始化客户端
	if err := handler.InitOSSClient(true); err != nil {
		return nil, err
	}

	var (
		thumbSize = [2]uint{400, 300}
		ok        = false
	)
	if thumbSize, ok = ctx.Value(fsctx.ThumbSizeCtx).([2]uint); !ok {
		return nil, errors.New("failed to get thumbnail size")
	}

	thumbEncodeQuality := model.GetIntSetting("thumb_encode_quality", 85)

	thumbParam := fmt.Sprintf("image/resize,m_lfit,h_%d,w_%d/quality,q_%d", thumbSize[1], thumbSize[0], thumbEncodeQuality)
	ctx = context.WithValue(ctx, fsctx.ThumbSizeCtx, thumbParam)
	thumbOption := []oss.Option{oss.Process(thumbParam)}
	thumbURL, err := handler.signSourceURL(
		ctx,
		file.SourceName,
		int64(model.GetIntSetting("preview_timeout", 60)),
		thumbOption,
	)
	if err != nil {
		return nil, err
	}

	return &response.ContentResponse{
		Redirect: true,
		URL:      thumbURL,
	}, nil
}

// Source 获取外链URL
func (handler *Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {
	// 初始化客户端
	usePublicEndpoint := true
	if forceUsePublicEndpoint, ok := ctx.Value(fsctx.ForceUsePublicEndpointCtx).(bool); ok {
		usePublicEndpoint = forceUsePublicEndpoint
	}
	if err := handler.InitOSSClient(usePublicEndpoint); err != nil {
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
		// Byte 转换为 bit
		speed *= 8

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

func (handler *Driver) signSourceURL(ctx context.Context, path string, ttl int64, options []oss.Option) (string, error) {
	signedURL, err := handler.bucket.SignURL(path, oss.HTTPGet, ttl, options...)
	if err != nil {
		return "", err
	}

	// 将最终生成的签名URL域名换成用户自定义的加速域名（如果有）
	finalURL, err := url.Parse(signedURL)
	if err != nil {
		return "", err
	}

	// 公有空间替换掉Key及不支持的头
	if !handler.Policy.IsPrivate {
		query := finalURL.Query()
		query.Del("OSSAccessKeyId")
		query.Del("Signature")
		query.Del("response-content-disposition")
		query.Del("x-oss-traffic-limit")
		finalURL.RawQuery = query.Encode()
	}

	if handler.Policy.BaseURL != "" {
		cdnURL, err := url.Parse(handler.Policy.BaseURL)
		if err != nil {
			return "", err
		}
		finalURL.Host = cdnURL.Host
		finalURL.Scheme = cdnURL.Scheme
	}

	return finalURL.String(), nil
}

// Token 获取上传策略和认证Token
func (handler *Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	// 初始化客户端
	if err := handler.InitOSSClient(true); err != nil {
		return nil, err
	}

	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/oss/" + uploadSession.Key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 回调策略
	callbackPolicy := CallbackPolicy{
		CallbackURL:      apiURL.String(),
		CallbackBody:     `{"name":${x:fname},"source_name":${object},"size":${size},"pic_info":"${imageInfo.width},${imageInfo.height}"}`,
		CallbackBodyType: "application/json",
	}
	callbackPolicyJSON, err := json.Marshal(callbackPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to encode callback policy: %w", err)
	}
	callbackPolicyEncoded := base64.StdEncoding.EncodeToString(callbackPolicyJSON)

	// 初始化分片上传
	fileInfo := file.Info()
	options := []oss.Option{
		oss.Expires(time.Now().Add(time.Duration(ttl) * time.Second)),
		oss.ForbidOverWrite(true),
		oss.ContentType(fileInfo.DetectMimeType()),
	}
	imur, err := handler.bucket.InitiateMultipartUpload(fileInfo.SavePath, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize multipart upload: %w", err)
	}
	uploadSession.UploadID = imur.UploadID

	// 为每个分片签名上传 URL
	chunks := chunk.NewChunkGroup(file, handler.Policy.OptionsSerialized.ChunkSize, &backoff.ConstantBackoff{}, false)
	urls := make([]string, chunks.Num())
	for chunks.Next() {
		err := chunks.Process(func(c *chunk.ChunkGroup, chunk io.Reader) error {
			signedURL, err := handler.bucket.SignURL(fileInfo.SavePath, oss.HTTPPut, ttl,
				oss.PartNumber(c.Index()+1),
				oss.UploadID(imur.UploadID),
				oss.ContentType("application/octet-stream"))
			if err != nil {
				return err
			}

			urls[c.Index()] = signedURL
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// 签名完成分片上传的URL
	completeURL, err := handler.bucket.SignURL(fileInfo.SavePath, oss.HTTPPost, ttl,
		oss.ContentType("application/octet-stream"),
		oss.UploadID(imur.UploadID),
		oss.Expires(time.Now().Add(time.Duration(ttl)*time.Second)),
		oss.CompleteAll("yes"),
		oss.ForbidOverWrite(true),
		oss.CallbackParam(callbackPolicyEncoded))
	if err != nil {
		return nil, err
	}

	return &serializer.UploadCredential{
		SessionID:   uploadSession.Key,
		ChunkSize:   handler.Policy.OptionsSerialized.ChunkSize,
		UploadID:    imur.UploadID,
		UploadURLs:  urls,
		CompleteURL: completeURL,
	}, nil
}

// 取消上传凭证
func (handler *Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	return handler.bucket.AbortMultipartUpload(oss.InitiateMultipartUploadResult{UploadID: uploadSession.UploadID, Key: uploadSession.SavePath}, nil)
}
