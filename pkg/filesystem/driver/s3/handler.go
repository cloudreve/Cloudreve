package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Driver 适配器模板
type Driver struct {
	Policy *model.Policy
	sess   *session.Session
	svc    *s3.S3
}

// UploadPolicy S3上传策略
type UploadPolicy struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

// MetaData 文件信息
type MetaData struct {
	Size uint64
	Etag string
}

func NewDriver(policy *model.Policy) (*Driver, error) {
	if policy.OptionsSerialized.ChunkSize == 0 {
		policy.OptionsSerialized.ChunkSize = 25 << 20 // 25 MB
	}

	driver := &Driver{
		Policy: policy,
	}

	return driver, driver.InitS3Client()
}

// InitS3Client 初始化S3会话
func (handler *Driver) InitS3Client() error {
	if handler.Policy == nil {
		return errors.New("empty policy")
	}

	if handler.svc == nil {
		// 初始化会话
		sess, err := session.NewSession(&aws.Config{
			Credentials:      credentials.NewStaticCredentials(handler.Policy.AccessKey, handler.Policy.SecretKey, ""),
			Endpoint:         &handler.Policy.Server,
			Region:           &handler.Policy.OptionsSerialized.Region,
			S3ForcePathStyle: &handler.Policy.OptionsSerialized.S3ForcePathStyle,
		})

		if err != nil {
			return err
		}
		handler.sess = sess
		handler.svc = s3.New(sess)
	}
	return nil
}

// List 列出给定路径下的文件
func (handler *Driver) List(ctx context.Context, base string, recursive bool) ([]response.Object, error) {
	// 初始化列目录参数
	base = strings.TrimPrefix(base, "/")
	if base != "" {
		base += "/"
	}

	opt := &s3.ListObjectsInput{
		Bucket:  &handler.Policy.BucketName,
		Prefix:  &base,
		MaxKeys: aws.Int64(1000),
	}

	// 是否为递归列出
	if !recursive {
		opt.Delimiter = aws.String("/")
	}

	var (
		objects []*s3.Object
		commons []*s3.CommonPrefix
	)

	for {
		res, err := handler.svc.ListObjectsWithContext(ctx, opt)
		if err != nil {
			return nil, err
		}
		objects = append(objects, res.Contents...)
		commons = append(commons, res.CommonPrefixes...)

		// 如果本次未列取完，则继续使用marker获取结果
		if *res.IsTruncated {
			opt.Marker = res.NextMarker
		} else {
			break
		}
	}

	// 处理列取结果
	res := make([]response.Object, 0, len(objects)+len(commons))

	// 处理目录
	for _, object := range commons {
		rel, err := filepath.Rel(*opt.Prefix, *object.Prefix)
		if err != nil {
			continue
		}
		res = append(res, response.Object{
			Name:         path.Base(*object.Prefix),
			RelativePath: filepath.ToSlash(rel),
			Size:         0,
			IsDir:        true,
			LastModify:   time.Now(),
		})
	}
	// 处理文件
	for _, object := range objects {
		rel, err := filepath.Rel(*opt.Prefix, *object.Key)
		if err != nil {
			continue
		}

		if strings.HasSuffix(*object.Key, "/") {
			res = append(res, response.Object{
				Name:         path.Base(*object.Key),
				RelativePath: filepath.ToSlash(rel),
				Size:         0,
				IsDir:        true,
				LastModify:   time.Now(),
			})
		} else {
			res = append(res, response.Object{
				Name:         path.Base(*object.Key),
				Source:       *object.Key,
				RelativePath: filepath.ToSlash(rel),
				Size:         uint64(*object.Size),
				IsDir:        false,
				LastModify:   *object.LastModified,
			})
		}
	}

	return res, nil

}

// Get 获取文件
func (handler *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 获取文件源地址
	downloadURL, err := handler.Source(ctx, path, int64(model.GetIntSetting("preview_timeout", 60)), false, 0)
	if err != nil {
		return nil, err
	}

	// 获取文件数据流
	client := request.NewClient()
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
func (handler *Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	defer file.Close()

	// 初始化客户端
	if err := handler.InitS3Client(); err != nil {
		return err
	}

	uploader := s3manager.NewUploader(handler.sess, func(u *s3manager.Uploader) {
		u.PartSize = int64(handler.Policy.OptionsSerialized.ChunkSize)
	})

	dst := file.Info().SavePath
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &handler.Policy.BucketName,
		Key:    &dst,
		Body:   io.LimitReader(file, int64(file.Info().Size)),
	})

	if err != nil {
		return err
	}

	return nil
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	failed := make([]string, 0, len(files))
	deleted := make([]string, 0, len(files))

	keys := make([]*s3.ObjectIdentifier, 0, len(files))
	for _, file := range files {
		filePath := file
		keys = append(keys, &s3.ObjectIdentifier{Key: &filePath})
	}

	// 发送异步删除请求
	res, err := handler.svc.DeleteObjects(
		&s3.DeleteObjectsInput{
			Bucket: &handler.Policy.BucketName,
			Delete: &s3.Delete{
				Objects: keys,
			},
		})

	if err != nil {
		return files, err
	}

	// 统计未删除的文件
	for _, deleteRes := range res.Deleted {
		deleted = append(deleted, *deleteRes.Key)
	}
	failed = util.SliceDifference(files, deleted)

	return failed, nil

}

// Thumb 获取文件缩略图
func (handler *Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	return nil, driver.ErrorThumbNotSupported
}

// Source 获取外链URL
func (handler *Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {

	// 尝试从上下文获取文件名
	fileName := ""
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	// 初始化客户端
	if err := handler.InitS3Client(); err != nil {
		return "", err
	}

	contentDescription := aws.String("attachment; filename=\"" + url.PathEscape(fileName) + "\"")
	if !isDownload {
		contentDescription = nil
	}
	req, _ := handler.svc.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket:                     &handler.Policy.BucketName,
			Key:                        &path,
			ResponseContentDisposition: contentDescription,
		})

	signedURL, err := req.Presign(time.Duration(ttl) * time.Second)
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
		finalURL.RawQuery = ""
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
	// 检查文件是否存在
	fileInfo := file.Info()
	if _, err := handler.Meta(ctx, fileInfo.SavePath); err == nil {
		return nil, fmt.Errorf("file already exist")
	}

	// 创建分片上传
	expires := time.Now().Add(time.Duration(ttl) * time.Second)
	res, err := handler.svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      &handler.Policy.BucketName,
		Key:         &fileInfo.SavePath,
		Expires:     &expires,
		ContentType: aws.String(fileInfo.DetectMimeType()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart upload: %w", err)
	}

	uploadSession.UploadID = *res.UploadId

	// 为每个分片签名上传 URL
	chunks := chunk.NewChunkGroup(file, handler.Policy.OptionsSerialized.ChunkSize, &backoff.ConstantBackoff{}, false)
	urls := make([]string, chunks.Num())
	for chunks.Next() {
		err := chunks.Process(func(c *chunk.ChunkGroup, chunk io.Reader) error {
			signedReq, _ := handler.svc.UploadPartRequest(&s3.UploadPartInput{
				Bucket:     &handler.Policy.BucketName,
				Key:        &fileInfo.SavePath,
				PartNumber: aws.Int64(int64(c.Index() + 1)),
				UploadId:   res.UploadId,
			})

			signedURL, err := signedReq.Presign(time.Duration(ttl) * time.Second)
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

	// 签名完成分片上传的请求URL
	signedReq, _ := handler.svc.CompleteMultipartUploadRequest(&s3.CompleteMultipartUploadInput{
		Bucket:   &handler.Policy.BucketName,
		Key:      &fileInfo.SavePath,
		UploadId: res.UploadId,
	})

	signedURL, err := signedReq.Presign(time.Duration(ttl) * time.Second)
	if err != nil {
		return nil, err
	}

	// 生成上传凭证
	return &serializer.UploadCredential{
		SessionID:   uploadSession.Key,
		ChunkSize:   handler.Policy.OptionsSerialized.ChunkSize,
		UploadID:    *res.UploadId,
		UploadURLs:  urls,
		CompleteURL: signedURL,
	}, nil
}

// Meta 获取文件信息
func (handler *Driver) Meta(ctx context.Context, path string) (*MetaData, error) {
	res, err := handler.svc.HeadObject(
		&s3.HeadObjectInput{
			Bucket: &handler.Policy.BucketName,
			Key:    &path,
		})

	if err != nil {
		return nil, err
	}

	return &MetaData{
		Size: uint64(*res.ContentLength),
		Etag: *res.ETag,
	}, nil

}

// CORS 创建跨域策略
func (handler *Driver) CORS() error {
	rule := s3.CORSRule{
		AllowedMethods: aws.StringSlice([]string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"HEAD",
		}),
		AllowedOrigins: aws.StringSlice([]string{"*"}),
		AllowedHeaders: aws.StringSlice([]string{"*"}),
		ExposeHeaders:  aws.StringSlice([]string{"ETag"}),
		MaxAgeSeconds:  aws.Int64(3600),
	}

	_, err := handler.svc.PutBucketCors(&s3.PutBucketCorsInput{
		Bucket: &handler.Policy.BucketName,
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: []*s3.CORSRule{&rule},
		},
	})

	return err
}

// 取消上传凭证
func (handler *Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	_, err := handler.svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		UploadId: &uploadSession.UploadID,
		Bucket:   &handler.Policy.BucketName,
		Key:      &uploadSession.SavePath,
	})
	return err
}
