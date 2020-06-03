package bos

import (
	"context"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

var UserPathPrefix = "user"

// Driver 策略适配器
type Driver struct {
	Policy *model.Policy
	ss     *session.Session
}

// InitSess 初始化会话
func (handler *Driver) connect() {
	if nil == handler.Policy {
		panic(errors.New("存储策略为空"))
	}

	handler.ss = session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(handler.Policy.AccessKey, handler.Policy.SecretKey, ""),
		Endpoint:         aws.String(handler.Policy.Server),
		Region:           aws.String(endpoints.CnNorth1RegionID),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true), // 这个必须设置为TRUE，坑了老半天
	}))
}

// 刷新 S3 Gateway 服务的订单信息
func (handler *Driver) init() error {
	if nil == handler.ss {
		handler.connect()
	}

	svc := s3.New(handler.ss)
	_, err := svc.ListBuckets(nil)

	return err
}

func prefix(path string) string {
	return UserPathPrefix + path
}

// List 列出BOS上存储的文件
func (handler Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	var res []response.Object
	// 刷新Buckets
	if err := handler.init(); nil != err {
		return nil, err
	}

	svc := s3.New(handler.ss)
	resp, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(handler.Policy.BucketName),
		Prefix: aws.String(prefix(path)),
	})
	if nil != err {
		return nil, err
	}

	for _, item := range resp.Contents {
		res = append(res, response.Object{
			Name:         filepath.Base(aws.StringValue(item.Key)),
			Size:         uint64(aws.Int64Value(item.Size)),
			LastModify:   aws.TimeValue(item.LastModified),
			RelativePath: filepath.ToSlash(aws.StringValue(item.Key)),
		})
	}

	return res, nil
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 刷新Buckets
	if err := handler.init(); nil != err {
		return nil, err
	}

	file, err := os.Create(filepath.Base(path))
	if nil != err {
		return nil, err
	}

	downloader := s3manager.NewDownloader(handler.ss)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(handler.Policy.BucketName),
		Key:    aws.String(prefix(path)),
	})

	if nil != err {
		return nil, err
	}

	return file, nil
}

// Put 将文件流保存到指定目录
func (handler Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	// 刷新Buckets
	if err := handler.init(); nil != err {
		return err
	}

	uploader := s3manager.NewUploader(handler.ss)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(handler.Policy.BucketName),
		Key:    aws.String(prefix(dst)),
		Body:   file,
	})

	return err
}

// Delete 删除一个或多个文件，
// 返回未删除的文件
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return files, errors.New("Lambda Chain 暂不支持删除操作")
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	//if nil == handler.ss {
	//	handler.connect()
	//}
	//
	//svc := s3.New(handler.ss)
	//result, err := svc.ListBuckets(nil)
	//if nil != err {
	//	return nil, err
	//}
	//
	//for _, b := range result.Buckets {
	//	log.Println(aws.StringValue(b.Name))
	//
	//	resp, err := svc.ListObjects(&s3.ListObjectsInput{
	//		Bucket: b.Name,
	//	})
	//
	//	if nil != err {
	//		log.Println("Error:", err)
	//		continue
	//	}
	//
	//	for _, item := range resp.Contents {
	//		log.Println("Key:", aws.StringValue(item.Key))
	//		log.Println("Size:", aws.Int64Value(item.Size))
	//		log.Println("Class:", aws.StringValue(item.StorageClass))
	//		log.Println("LastTime:", aws.TimeValue(item.LastModified))
	//	}
	//}
	return &response.ContentResponse{
		Redirect: true,
		URL:      "",
	}, nil
}

// Source 获取外链URL
func (handler Driver) Source(ctx context.Context, path string, baseURL url.URL, ttl int64, isDownload bool, speed int) (string, error) {
	file, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return "", errors.New("无法获取文件记录上下文")
	}

	var (
		signedURI *url.URL
		err       error
	)
	if isDownload {
		// 创建下载会话，将文件信息写入缓存
		downloadSessionID := util.RandStringRunes(16)
		err = cache.Set("download_"+downloadSessionID, file, int(ttl))
		if err != nil {
			return "", serializer.NewError(serializer.CodeCacheOperation, "无法创建下载会话", err)
		}

		// 签名生成文件记录
		signedURI, err = auth.SignURI(
			auth.General,
			fmt.Sprintf("/api/v3/file/download/%s", downloadSessionID),
			ttl,
		)
	}

	if nil != err {
		return "", serializer.NewError(serializer.CodeEncryptError, "无法对URL进行签名", err)

	}

	finalURL := baseURL.ResolveReference(signedURI).String()

	return finalURL, nil
}

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/bos/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI).String()

	return serializer.UploadCredential{
		Callback: apiURL,
		Key:      key,
	}, nil
}

// Meta 获取文件信息用于回调验证
// Cloudreve 本身不允许同名文件存在，所以通过Cloudreve上传到
// Lambda BOS存储的文件原则上也不会有重名，所以用文件名称和大
// 小做验证信息，比较粗劣，后续优化
//func (handler Driver) Meta(path string) (uint64, error) {
//	// 刷新Buckets
//	if err := handler.init(); nil != err {
//		return 0, err
//	}
//
//	dir, _ := filepath.Split(path)
//	svc := s3.New(handler.ss)
//	resp, err := svc.ListObjects(&s3.ListObjectsInput{
//		Bucket: aws.String(handler.Policy.BucketName),
//		Prefix: aws.String(prefix(dir)),
//	})
//	if nil != err {
//		return 0, err
//	}
//
//	for _, item := range resp.Contents {
//		if path == aws.StringValue(item.Key) {
//			return uint64(aws.Int64Value(item.Size)), nil
//		}
//	}
//
//	return 0, errors.New("文件验证异常")
//}
