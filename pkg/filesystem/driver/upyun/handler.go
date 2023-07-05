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
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/upyun/go-sdk/upyun"
)

// UploadPolicy 又拍云上传策略
type UploadPolicy struct {
	Bucket             string `json:"bucket"`
	SaveKey            string `json:"save-key"`
	Expiration         int64  `json:"expiration"`
	CallbackURL        string `json:"notify-url"`
	ContentLength      uint64 `json:"content-length"`
	ContentLengthRange string `json:"content-length-range,omitempty"`
	AllowFileType      string `json:"allow-file-type,omitempty"`
}

// Driver 又拍云策略适配器
type Driver struct {
	Policy *model.Policy
}

func (handler Driver) List(ctx context.Context, base string, recursive bool) ([]response.Object, error) {
	base = strings.TrimPrefix(base, "/")

	// 用于接受SDK返回对象的chan
	objChan := make(chan *upyun.FileInfo)
	objects := []*upyun.FileInfo{}

	// 列取配置
	listConf := &upyun.GetObjectsConfig{
		Path:         "/" + base,
		ObjectsChan:  objChan,
		MaxListTries: 1,
	}
	// 递归列取时不限制递归次数
	if recursive {
		listConf.MaxListLevel = -1
	}

	// 启动一个goroutine收集列取的对象信
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(input chan *upyun.FileInfo, output *[]*upyun.FileInfo, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			file, ok := <-input
			if !ok {
				return
			}
			*output = append(*output, file)
		}
	}(objChan, &objects, wg)

	up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   handler.Policy.BucketName,
		Operator: handler.Policy.AccessKey,
		Password: handler.Policy.SecretKey,
	})

	err := up.List(listConf)
	if err != nil {
		return nil, err
	}

	wg.Wait()

	// 汇总处理列取结果
	res := make([]response.Object, 0, len(objects))
	for _, object := range objects {
		res = append(res, response.Object{
			Name:         path.Base(object.Name),
			RelativePath: object.Name,
			Source:       path.Join(base, object.Name),
			Size:         uint64(object.Size),
			IsDir:        object.IsDir,
			LastModify:   object.Time,
		})
	}

	return res, nil
}

// Get 获取文件
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
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
func (handler Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	defer file.Close()

	up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   handler.Policy.BucketName,
		Operator: handler.Policy.AccessKey,
		Password: handler.Policy.SecretKey,
	})
	err := up.Put(&upyun.PutObjectConfig{
		Path:   file.Info().SavePath,
		Reader: file,
	})

	return err
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   handler.Policy.BucketName,
		Operator: handler.Policy.AccessKey,
		Password: handler.Policy.SecretKey,
	})

	var (
		failed       = make([]string, 0, len(files))
		lastErr      error
		currentIndex = 0
		indexLock    sync.Mutex
		failedLock   sync.Mutex
		wg           sync.WaitGroup
		routineNum   = 4
	)
	wg.Add(routineNum)

	// upyun不支持批量操作，这里开四个协程并行操作
	for i := 0; i < routineNum; i++ {
		go func() {
			for {
				// 取得待删除文件
				indexLock.Lock()
				if currentIndex >= len(files) {
					// 所有文件处理完成
					wg.Done()
					indexLock.Unlock()
					return
				}
				path := files[currentIndex]
				currentIndex++
				indexLock.Unlock()

				// 发送异步删除请求
				err := up.Delete(&upyun.DeleteObjectConfig{
					Path:  path,
					Async: true,
				})

				// 处理错误
				if err != nil {
					failedLock.Lock()
					lastErr = err
					failed = append(failed, path)
					failedLock.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	return failed, lastErr
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	// quick check by extension name
	// https://help.upyun.com/knowledge-base/image/
	supported := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "svg"}
	if len(handler.Policy.OptionsSerialized.ThumbExts) > 0 {
		supported = handler.Policy.OptionsSerialized.ThumbExts
	}

	if !util.IsInExtensionList(supported, file.Name) {
		return nil, driver.ErrorThumbNotSupported
	}

	var (
		thumbSize = [2]uint{400, 300}
		ok        = false
	)
	if thumbSize, ok = ctx.Value(fsctx.ThumbSizeCtx).([2]uint); !ok {
		return nil, errors.New("failed to get thumbnail size")
	}

	thumbEncodeQuality := model.GetIntSetting("thumb_encode_quality", 85)

	thumbParam := fmt.Sprintf("!/fwfh/%dx%d/quality/%d", thumbSize[0], thumbSize[1], thumbEncodeQuality)
	thumbURL, err := handler.Source(ctx, file.SourceName+thumbParam, int64(model.GetIntSetting("preview_timeout", 60)), false, 0)
	if err != nil {
		return nil, err
	}

	return &response.ContentResponse{
		Redirect: true,
		URL:      thumbURL,
	}, nil
}

// Source 获取外链URL
func (handler Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {
	// 尝试从上下文获取文件名
	fileName := ""
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	sourceURL, err := url.Parse(handler.Policy.BaseURL)
	if err != nil {
		return "", err
	}

	fileKey, err := url.Parse(url.PathEscape(path))
	if err != nil {
		return "", err
	}

	sourceURL = sourceURL.ResolveReference(fileKey)

	// 如果是下载文件URL
	if isDownload {
		query := sourceURL.Query()
		query.Add("_upd", fileName)
		sourceURL.RawQuery = query.Encode()
	}

	return handler.signURL(ctx, sourceURL, ttl)
}

func (handler Driver) signURL(ctx context.Context, path *url.URL, TTL int64) (string, error) {
	if !handler.Policy.IsPrivate {
		// 未开启Token防盗链时，直接返回
		return path.String(), nil
	}

	etime := time.Now().Add(time.Duration(TTL) * time.Second).Unix()
	signStr := fmt.Sprintf(
		"%s&%d&%s",
		handler.Policy.OptionsSerialized.Token,
		etime,
		path.Path,
	)
	signMd5 := fmt.Sprintf("%x", md5.Sum([]byte(signStr)))
	finalSign := signMd5[12:20] + strconv.FormatInt(etime, 10)

	// 将签名添加到URL中
	query := path.Query()
	query.Add("_upt", finalSign)
	path.RawQuery = query.Encode()

	return path.String(), nil
}

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/upyun/" + uploadSession.Key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 上传策略
	fileInfo := file.Info()
	putPolicy := UploadPolicy{
		Bucket: handler.Policy.BucketName,
		// TODO escape
		SaveKey:            fileInfo.SavePath,
		Expiration:         time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
		CallbackURL:        apiURL.String(),
		ContentLength:      fileInfo.Size,
		ContentLengthRange: fmt.Sprintf("0,%d", fileInfo.Size),
		AllowFileType:      strings.Join(handler.Policy.OptionsSerialized.FileType, ","),
	}

	// 生成上传凭证
	policyJSON, err := json.Marshal(putPolicy)
	if err != nil {
		return nil, err
	}
	policyEncoded := base64.StdEncoding.EncodeToString(policyJSON)

	// 生成签名
	elements := []string{"POST", "/" + handler.Policy.BucketName, policyEncoded}
	signStr := handler.Sign(ctx, elements)

	return &serializer.UploadCredential{
		SessionID:  uploadSession.Key,
		Policy:     policyEncoded,
		Credential: signStr,
		UploadURLs: []string{"https://v0.api.upyun.com/" + handler.Policy.BucketName},
	}, nil
}

// 取消上传凭证
func (handler Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	return nil
}

// Sign 计算又拍云的签名头
func (handler Driver) Sign(ctx context.Context, elements []string) string {
	password := fmt.Sprintf("%x", md5.Sum([]byte(handler.Policy.SecretKey)))
	mac := hmac.New(sha1.New, []byte(password))
	value := strings.Join(elements, "&")
	mac.Write([]byte(value))
	signStr := base64.StdEncoding.EncodeToString((mac.Sum(nil)))
	return fmt.Sprintf("UPYUN %s:%s", handler.Policy.AccessKey, signStr)
}
