package onedrive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

const (
	// SmallFileSize 单文件上传接口最大尺寸
	SmallFileSize uint64 = 4 * 1024 * 1024
	// ChunkSize 服务端中转分片上传分片大小
	ChunkSize uint64 = 10 * 1024 * 1024
	// ListRetry 列取请求重试次数
	ListRetry = 1
)

// GetSourcePath 获取文件的绝对路径
func (info *FileInfo) GetSourcePath() string {
	res, err := url.PathUnescape(
		strings.TrimPrefix(
			path.Join(
				strings.TrimPrefix(info.ParentReference.Path, "/drive/root:"),
				info.Name,
			),
			"/",
		),
	)
	if err != nil {
		return ""
	}
	return res
}

// Error 实现error接口
func (err RespError) Error() string {
	return err.APIError.Message
}

func (client *Client) getRequestURL(api string) string {
	base, _ := url.Parse(client.Endpoints.EndpointURL)
	if base == nil {
		return ""
	}
	base.Path = path.Join(base.Path, api)
	return base.String()
}

// ListChildren 根据路径列取子对象
func (client *Client) ListChildren(ctx context.Context, path string) ([]FileInfo, error) {
	var requestURL string
	dst := strings.TrimPrefix(path, "/")
	if dst == "" {
		requestURL = client.getRequestURL("me/drive/root/children")
	} else {
		requestURL = client.getRequestURL("me/drive/root:/" + dst + ":/children")
	}

	res, err := client.requestWithStr(ctx, "GET", requestURL+"?$top=999999999", "", 200)
	if err != nil {
		retried := 0
		if v, ok := ctx.Value(fsctx.RetryCtx).(int); ok {
			retried = v
		}
		if retried < ListRetry {
			retried++
			util.Log().Debug("路径[%s]列取请求失败[%s]，5秒钟后重试", path, err)
			time.Sleep(time.Duration(5) * time.Second)
			return client.ListChildren(context.WithValue(ctx, fsctx.RetryCtx, retried), path)
		}
		return nil, err
	}

	var (
		decodeErr error
		fileInfo  ListResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &fileInfo)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return fileInfo.Value, nil
}

// Meta 根据资源ID或文件路径获取文件元信息
func (client *Client) Meta(ctx context.Context, id string, path string) (*FileInfo, error) {
	var requestURL string
	if id != "" {
		requestURL = client.getRequestURL("/me/drive/items/" + id)
	} else {
		dst := strings.TrimPrefix(path, "/")
		requestURL = client.getRequestURL("me/drive/root:/" + dst)
	}

	res, err := client.requestWithStr(ctx, "GET", requestURL+"?expand=thumbnails", "", 200)
	if err != nil {
		return nil, err
	}

	var (
		decodeErr error
		fileInfo  FileInfo
	)
	decodeErr = json.Unmarshal([]byte(res), &fileInfo)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &fileInfo, nil

}

// CreateUploadSession 创建分片上传会话
func (client *Client) CreateUploadSession(ctx context.Context, dst string, opts ...Option) (string, error) {

	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	dst = strings.TrimPrefix(dst, "/")
	requestURL := client.getRequestURL("me/drive/root:/" + dst + ":/createUploadSession")
	body := map[string]map[string]interface{}{
		"item": {
			"@microsoft.graph.conflictBehavior": options.conflictBehavior,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	res, err := client.requestWithStr(ctx, "POST", requestURL, string(bodyBytes), 200)
	if err != nil {
		return "", err
	}

	var (
		decodeErr     error
		uploadSession UploadSessionResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadSession)
	if decodeErr != nil {
		return "", decodeErr
	}

	return uploadSession.UploadURL, nil
}

// GetUploadSessionStatus 查询上传会话状态
func (client *Client) GetUploadSessionStatus(ctx context.Context, uploadURL string) (*UploadSessionResponse, error) {
	res, err := client.requestWithStr(ctx, "GET", uploadURL, "", 200)
	if err != nil {
		return nil, err
	}

	var (
		decodeErr     error
		uploadSession UploadSessionResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadSession)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &uploadSession, nil
}

// UploadChunk 上传分片
func (client *Client) UploadChunk(ctx context.Context, uploadURL string, chunk *Chunk) (*UploadSessionResponse, error) {
	res, err := client.request(
		ctx, "PUT", uploadURL, bytes.NewReader(chunk.Data[0:chunk.ChunkSize]),
		request.WithContentLength(int64(chunk.ChunkSize)),
		request.WithHeader(http.Header{
			"Content-Range": {fmt.Sprintf("bytes %d-%d/%d", chunk.Offset, chunk.Offset+chunk.ChunkSize-1, chunk.Total)},
		}),
		request.WithoutHeader([]string{"Authorization", "Content-Type"}),
		request.WithTimeout(time.Duration(300)*time.Second),
	)
	if err != nil {
		// 如果重试次数小于限制，5秒后重试
		if chunk.Retried < model.GetIntSetting("onedrive_chunk_retries", 1) {
			chunk.Retried++
			util.Log().Debug("分片偏移%d上传失败[%s]，5秒钟后重试", chunk.Offset, err)
			time.Sleep(time.Duration(5) * time.Second)
			return client.UploadChunk(ctx, uploadURL, chunk)
		}
		return nil, err
	}

	if chunk.IsLast() {
		return nil, nil
	}

	var (
		decodeErr error
		uploadRes UploadSessionResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadRes)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &uploadRes, nil
}

// Upload 上传文件
func (client *Client) Upload(ctx context.Context, dst string, size int, file io.Reader) error {
	// 小文件，使用简单上传接口上传
	if size <= int(SmallFileSize) {
		_, err := client.SimpleUpload(ctx, dst, file, int64(size))
		return err
	}

	// 大文件，进行分片
	// 创建上传会话
	uploadURL, err := client.CreateUploadSession(ctx, dst, WithConflictBehavior("replace"))
	if err != nil {
		return err
	}

	offset := 0
	chunkNum := size / int(ChunkSize)
	if size%int(ChunkSize) != 0 {
		chunkNum++
	}

	chunkData := make([]byte, ChunkSize)

	for i := 0; i < chunkNum; i++ {
		select {
		case <-ctx.Done():
			util.Log().Debug("OneDrive 客户端取消")
			return ErrClientCanceled
		default:
			// 分块
			chunkSize := int(ChunkSize)
			if size-offset < chunkSize {
				chunkSize = size - offset
			}

			// 因为后面需要错误重试，这里要把分片内容读到内存中
			chunkContent := chunkData[:chunkSize]
			_, err := io.ReadFull(file, chunkContent)

			chunk := Chunk{
				Offset:    offset,
				ChunkSize: chunkSize,
				Total:     size,
				Data:      chunkContent,
			}

			// 上传
			_, err = client.UploadChunk(ctx, uploadURL, &chunk)
			if err != nil {
				return err
			}
			offset += chunkSize
		}

	}
	return nil
}

// DeleteUploadSession 删除上传会话
func (client *Client) DeleteUploadSession(ctx context.Context, uploadURL string) error {
	_, err := client.requestWithStr(ctx, "DELETE", uploadURL, "", 204)
	if err != nil {
		return err
	}

	return nil
}

// SimpleUpload 上传小文件到dst
func (client *Client) SimpleUpload(ctx context.Context, dst string, body io.Reader, size int64) (*UploadResult, error) {
	dst = strings.TrimPrefix(dst, "/")
	requestURL := client.getRequestURL("me/drive/root:/" + dst + ":/content")

	res, err := client.request(ctx, "PUT", requestURL, body, request.WithContentLength(int64(size)),
		request.WithTimeout(time.Duration(150)*time.Second),
	)
	if err != nil {
		retried := 0
		if v, ok := ctx.Value(fsctx.RetryCtx).(int); ok {
			retried = v
		}
		if retried < model.GetIntSetting("onedrive_chunk_retries", 1) {
			retried++
			util.Log().Debug("文件[%s]上传失败[%s]，5秒钟后重试", dst, err)
			time.Sleep(time.Duration(5) * time.Second)
			return client.SimpleUpload(context.WithValue(ctx, fsctx.RetryCtx, retried), dst, body, size)
		}
		return nil, err
	}

	var (
		decodeErr error
		uploadRes UploadResult
	)
	decodeErr = json.Unmarshal([]byte(res), &uploadRes)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &uploadRes, nil
}

// BatchDelete 并行删除给出的文件，返回删除失败的文件，及第一个遇到的错误。此方法将文件分为
// 20个一组，调用Delete并行删除
// TODO 测试
func (client *Client) BatchDelete(ctx context.Context, dst []string) ([]string, error) {
	groupNum := len(dst)/20 + 1
	finalRes := make([]string, 0, len(dst))
	res := make([]string, 0, 20)
	var err error

	for i := 0; i < groupNum; i++ {
		end := 20*i + 20
		if i == groupNum-1 {
			end = len(dst)
		}
		res, err = client.Delete(ctx, dst[20*i:end])
		finalRes = append(finalRes, res...)
	}

	return finalRes, err
}

// Delete 并行删除文件，返回删除失败的文件，及第一个遇到的错误，
// 由于API限制，最多删除20个
func (client *Client) Delete(ctx context.Context, dst []string) ([]string, error) {
	body := client.makeBatchDeleteRequestsBody(dst)
	res, err := client.requestWithStr(ctx, "POST", client.getRequestURL("$batch"), body, 200)
	if err != nil {
		return dst, err
	}

	var (
		decodeErr error
		deleteRes BatchResponses
	)
	decodeErr = json.Unmarshal([]byte(res), &deleteRes)
	if decodeErr != nil {
		return dst, decodeErr
	}

	// 取得删除失败的文件
	failed := getDeleteFailed(&deleteRes)
	if len(failed) != 0 {
		return failed, ErrDeleteFile
	}
	return failed, nil
}

func getDeleteFailed(res *BatchResponses) []string {
	var failed = make([]string, 0, len(res.Responses))
	for _, v := range res.Responses {
		if v.Status != 204 {
			failed = append(failed, v.ID)
		}
	}
	return failed
}

// makeBatchDeleteRequestsBody 生成批量删除请求正文
func (client *Client) makeBatchDeleteRequestsBody(files []string) string {
	req := BatchRequests{
		Requests: make([]BatchRequest, len(files)),
	}
	for i, v := range files {
		v = strings.TrimPrefix(v, "/")
		filePath, _ := url.Parse("/me/drive/root:/")
		filePath.Path = path.Join(filePath.Path, v)
		req.Requests[i] = BatchRequest{
			ID:     v,
			Method: "DELETE",
			URL:    filePath.EscapedPath(),
		}
	}

	res, _ := json.Marshal(req)
	return string(res)
}

// GetThumbURL 获取给定尺寸的缩略图URL
func (client *Client) GetThumbURL(ctx context.Context, dst string, w, h uint) (string, error) {
	dst = strings.TrimPrefix(dst, "/")
	var (
		cropOption string
		requestURL string
	)
	if client.Endpoints.isInChina {
		cropOption = "large"
		requestURL = client.getRequestURL("me/drive/root:/"+dst+":/thumbnails/0") + "/" + cropOption
	} else {
		cropOption = fmt.Sprintf("c%dx%d_Crop", w, h)
		requestURL = client.getRequestURL("me/drive/root:/"+dst+":/thumbnails") + "?select=" + cropOption
	}

	res, err := client.requestWithStr(ctx, "GET", requestURL, "", 200)
	if err != nil {
		return "", err
	}

	var (
		decodeErr error
		thumbRes  ThumbResponse
	)
	decodeErr = json.Unmarshal([]byte(res), &thumbRes)
	if decodeErr != nil {
		return "", decodeErr
	}

	if thumbRes.URL != "" {
		return thumbRes.URL, nil
	}

	if len(thumbRes.Value) == 1 {
		if res, ok := thumbRes.Value[0][cropOption]; ok {
			return res.(map[string]interface{})["url"].(string), nil
		}
	}

	return "", errors.New("无法生成缩略图")
}

// MonitorUpload 监控客户端分片上传进度
func (client *Client) MonitorUpload(uploadURL, callbackKey, path string, size uint64, ttl int64) {
	// 回调完成通知chan
	callbackChan := make(chan bool)
	callbackSignal.Store(callbackKey, callbackChan)
	defer callbackSignal.Delete(callbackKey)
	timeout := model.GetIntSetting("onedrive_monitor_timeout", 600)
	interval := model.GetIntSetting("onedrive_callback_check", 20)

	for {
		select {
		case <-callbackChan:
			util.Log().Debug("客户端完成回调")
			return
		case <-time.After(time.Duration(ttl) * time.Second):
			// 上传会话到期，仍未完成上传，创建占位符
			client.DeleteUploadSession(context.Background(), uploadURL)
			_, err := client.SimpleUpload(context.Background(), path, strings.NewReader(""), 0)
			if err != nil {
				util.Log().Debug("无法创建占位文件，%s", err)
			}
			return
		case <-time.After(time.Duration(timeout) * time.Second):
			util.Log().Debug("检查上传情况")
			status, err := client.GetUploadSessionStatus(context.Background(), uploadURL)

			if err != nil {
				if resErr, ok := err.(*RespError); ok {
					if resErr.APIError.Code == "itemNotFound" {
						util.Log().Debug("上传会话已完成，稍后检查回调")
						time.Sleep(time.Duration(interval) * time.Second)
						util.Log().Debug("开始检查回调")
						_, ok := cache.Get("callback_" + callbackKey)
						if ok {
							util.Log().Warning("未发送回调，删除文件")
							cache.Deletes([]string{callbackKey}, "callback_")
							_, err = client.Delete(context.Background(), []string{path})
							if err != nil {
								util.Log().Warning("无法删除未回调的文件，%s", err)
							}
						}
						return
					}
				}
				util.Log().Debug("无法获取上传会话状态，继续下一轮，%s", err.Error())
				continue
			}

			// 成功获取分片上传状态，检查文件大小
			if len(status.NextExpectedRanges) == 0 {
				continue
			}
			sizeRange := strings.Split(
				status.NextExpectedRanges[len(status.NextExpectedRanges)-1],
				"-",
			)
			if len(sizeRange) != 2 {
				continue
			}
			uploadFullSize, _ := strconv.ParseUint(sizeRange[1], 10, 64)
			if (sizeRange[0] == "0" && sizeRange[1] == "") || uploadFullSize+1 != size {
				util.Log().Debug("未开始上传或文件大小不一致，取消上传会话")
				// 取消上传会话，实测OneDrive取消上传会话后，客户端还是可以上传，
				// 所以上传一个空文件占位，阻止客户端上传
				client.DeleteUploadSession(context.Background(), uploadURL)
				_, err := client.SimpleUpload(context.Background(), path, strings.NewReader(""), 0)
				if err != nil {
					util.Log().Debug("无法创建占位文件，%s", err)
				}
				return
			}

		}
	}
}

// FinishCallback 向Monitor发送回调结束信号
func FinishCallback(key string) {
	if signal, ok := callbackSignal.Load(key); ok {
		if signalChan, ok := signal.(chan bool); ok {
			close(signalChan)
		}
	}
}

func sysError(err error) *RespError {
	return &RespError{APIError: APIError{
		Code:    "system",
		Message: err.Error(),
	}}
}

func (client *Client) request(ctx context.Context, method string, url string, body io.Reader, option ...request.Option) (string, *RespError) {
	// 获取凭证
	err := client.UpdateCredential(ctx)
	if err != nil {
		return "", sysError(err)
	}

	option = append(option,
		request.WithHeader(http.Header{
			"Authorization": {"Bearer " + client.Credential.AccessToken},
			"Content-Type":  {"application/json"},
		}),
		request.WithContext(ctx),
	)

	// 发送请求
	res := client.Request.Request(
		method,
		url,
		body,
		option...,
	)

	if res.Err != nil {
		return "", sysError(res.Err)
	}

	respBody, err := res.GetResponse()
	if err != nil {
		return "", sysError(err)
	}

	// 解析请求响应
	var (
		errResp   RespError
		decodeErr error
	)
	// 如果有错误
	if res.Response.StatusCode < 200 || res.Response.StatusCode >= 300 {
		decodeErr = json.Unmarshal([]byte(respBody), &errResp)
		if decodeErr != nil {
			util.Log().Debug("Onedrive返回未知响应[%s]", respBody)
			return "", sysError(decodeErr)
		}
		return "", &errResp
	}

	return respBody, nil
}

func (client *Client) requestWithStr(ctx context.Context, method string, url string, body string, expectedCode int) (string, *RespError) {
	// 发送请求
	bodyReader := ioutil.NopCloser(strings.NewReader(body))
	return client.request(ctx, method, url, bodyReader,
		request.WithContentLength(int64(len(body))),
	)
}
