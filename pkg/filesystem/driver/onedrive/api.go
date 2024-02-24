package onedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/chunk/backoff"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

const (
	// SmallFileSize 单文件上传接口最大尺寸
	SmallFileSize uint64 = 4 * 1024 * 1024
	// ChunkSize 服务端中转分片上传分片大小
	ChunkSize uint64 = 10 * 1024 * 1024
	// ListRetry 列取请求重试次数
	ListRetry       = 1
	chunkRetrySleep = time.Second * 5

	notFoundError = "itemNotFound"
)

// GetSourcePath 获取文件的绝对路径
func (info *FileInfo) GetSourcePath() string {
	res, err := url.PathUnescape(info.ParentReference.Path)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(
		path.Join(
			strings.TrimPrefix(res, "/drive/root:"),
			info.Name,
		),
		"/",
	)
}

func (client *Client) getRequestURL(api string, opts ...Option) string {
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	base, _ := url.Parse(client.Endpoints.EndpointURL)
	if base == nil {
		return ""
	}

	if options.useDriverResource {
		base.Path = path.Join(base.Path, client.Endpoints.DriverResource, api)
	} else {
		base.Path = path.Join(base.Path, api)
	}

	return base.String()
}

// ListChildren 根据路径列取子对象
func (client *Client) ListChildren(ctx context.Context, path string) ([]FileInfo, error) {
	var requestURL string
	dst := strings.TrimPrefix(path, "/")
	if dst == "" {
		requestURL = client.getRequestURL("root/children")
	} else {
		requestURL = client.getRequestURL("root:/" + dst + ":/children")
	}

	res, err := client.requestWithStr(ctx, "GET", requestURL+"?$top=999999999", "", 200)
	if err != nil {
		retried := 0
		if v, ok := ctx.Value(fsctx.RetryCtx).(int); ok {
			retried = v
		}
		if retried < ListRetry {
			retried++
			util.Log().Debug("Failed to list path %q: %s, will retry in 5 seconds.", path, err)
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
		requestURL = client.getRequestURL("items/" + id)
	} else {
		dst := strings.TrimPrefix(path, "/")
		requestURL = client.getRequestURL("root:/" + dst)
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
	requestURL := client.getRequestURL("root:/" + dst + ":/createUploadSession")
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

// GetSiteIDByURL 通过 SharePoint 站点 URL 获取站点ID
func (client *Client) GetSiteIDByURL(ctx context.Context, siteUrl string) (string, error) {
	siteUrlParsed, err := url.Parse(siteUrl)
	if err != nil {
		return "", err
	}

	hostName := siteUrlParsed.Hostname()
	relativePath := strings.Trim(siteUrlParsed.Path, "/")
	requestURL := client.getRequestURL(fmt.Sprintf("sites/%s:/%s", hostName, relativePath), WithDriverResource(false))
	res, reqErr := client.requestWithStr(ctx, "GET", requestURL, "", 200)
	if reqErr != nil {
		return "", reqErr
	}

	var (
		decodeErr error
		siteInfo  Site
	)
	decodeErr = json.Unmarshal([]byte(res), &siteInfo)
	if decodeErr != nil {
		return "", decodeErr
	}

	return siteInfo.ID, nil
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
func (client *Client) UploadChunk(ctx context.Context, uploadURL string, content io.Reader, current *chunk.ChunkGroup) (*UploadSessionResponse, error) {
	res, err := client.request(
		ctx, "PUT", uploadURL, content,
		request.WithContentLength(current.Length()),
		request.WithHeader(http.Header{
			"Content-Range": {current.RangeHeader()},
		}),
		request.WithoutHeader([]string{"Authorization", "Content-Type"}),
		request.WithTimeout(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload OneDrive chunk #%d: %w", current.Index(), err)
	}

	if current.IsLast() {
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
func (client *Client) Upload(ctx context.Context, file fsctx.FileHeader) error {
	fileInfo := file.Info()
	// 决定是否覆盖文件
	overwrite := "fail"
	if fileInfo.Mode&fsctx.Overwrite == fsctx.Overwrite {
		overwrite = "replace"
	}

	size := int(fileInfo.Size)
	dst := fileInfo.SavePath

	// 小文件，使用简单上传接口上传
	if size <= int(SmallFileSize) {
		_, err := client.SimpleUpload(ctx, dst, file, int64(size), WithConflictBehavior(overwrite))
		return err
	}

	// 大文件，进行分片
	// 创建上传会话
	uploadURL, err := client.CreateUploadSession(ctx, dst, WithConflictBehavior(overwrite))
	if err != nil {
		return err
	}

	// Initial chunk groups
	chunks := chunk.NewChunkGroup(file, client.Policy.OptionsSerialized.ChunkSize, &backoff.ConstantBackoff{
		Max:   model.GetIntSetting("chunk_retries", 5),
		Sleep: chunkRetrySleep,
	}, model.IsTrueVal(model.GetSettingByName("use_temp_chunk_buffer")))

	uploadFunc := func(current *chunk.ChunkGroup, content io.Reader) error {
		_, err := client.UploadChunk(ctx, uploadURL, content, current)
		return err
	}

	// upload chunks
	for chunks.Next() {
		if err := chunks.Process(uploadFunc); err != nil {
			return fmt.Errorf("failed to upload chunk #%d: %w", chunks.Index(), err)
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
func (client *Client) SimpleUpload(ctx context.Context, dst string, body io.Reader, size int64, opts ...Option) (*UploadResult, error) {
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	dst = strings.TrimPrefix(dst, "/")
	requestURL := client.getRequestURL("root:/" + dst + ":/content")
	requestURL += ("?@microsoft.graph.conflictBehavior=" + options.conflictBehavior)

	res, err := client.request(ctx, "PUT", requestURL, body, request.WithContentLength(int64(size)),
		request.WithTimeout(0),
	)
	if err != nil {
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
	res, err := client.requestWithStr(ctx, "POST", client.getRequestURL("$batch",
		WithDriverResource(false)), body, 200)
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
		if v.Status != 204 && v.Status != 404 {
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
		filePath, _ := url.Parse("/" + client.Endpoints.DriverResource + "/root:/")
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
	requestURL := client.getRequestURL("root:/"+dst+":/thumbnails/0") + "/large"

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
		if res, ok := thumbRes.Value[0]["large"]; ok {
			return res.(map[string]interface{})["url"].(string), nil
		}
	}

	return "", ErrThumbSizeNotFound
}

// MonitorUpload 监控客户端分片上传进度
func (client *Client) MonitorUpload(uploadURL, callbackKey, path string, size uint64, ttl int64) {
	// 回调完成通知chan
	callbackChan := mq.GlobalMQ.Subscribe(callbackKey, 1)
	defer mq.GlobalMQ.Unsubscribe(callbackKey, callbackChan)

	timeout := model.GetIntSetting("onedrive_monitor_timeout", 600)
	interval := model.GetIntSetting("onedrive_callback_check", 20)

	for {
		select {
		case <-callbackChan:
			util.Log().Debug("Client finished OneDrive callback.")
			return
		case <-time.After(time.Duration(ttl) * time.Second):
			// 上传会话到期，仍未完成上传，创建占位符
			client.DeleteUploadSession(context.Background(), uploadURL)
			_, err := client.SimpleUpload(context.Background(), path, strings.NewReader(""), 0, WithConflictBehavior("replace"))
			if err != nil {
				util.Log().Debug("Failed to create placeholder file: %s", err)
			}
			return
		case <-time.After(time.Duration(timeout) * time.Second):
			util.Log().Debug("Checking OneDrive upload status.")
			status, err := client.GetUploadSessionStatus(context.Background(), uploadURL)

			if err != nil {
				if resErr, ok := err.(*RespError); ok {
					if resErr.APIError.Code == notFoundError {
						util.Log().Debug("Upload completed, will check upload callback later.")
						select {
						case <-time.After(time.Duration(interval) * time.Second):
							util.Log().Warning("No callback is made, file will be deleted.")
							cache.Deletes([]string{callbackKey}, "callback_")
							_, err = client.Delete(context.Background(), []string{path})
							if err != nil {
								util.Log().Warning("Failed to delete file without callback: %s", err)
							}
						case <-callbackChan:
							util.Log().Debug("Client finished callback.")
						}
						return
					}
				}
				util.Log().Debug("Failed to get upload session status: %s, continue next iteration.", err.Error())
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
				util.Log().Debug("Upload has not started, or uploaded file size not match, canceling upload session...")
				// 取消上传会话，实测OneDrive取消上传会话后，客户端还是可以上传，
				// 所以上传一个空文件占位，阻止客户端上传
				client.DeleteUploadSession(context.Background(), uploadURL)
				_, err := client.SimpleUpload(context.Background(), path, strings.NewReader(""), 0, WithConflictBehavior("replace"))
				if err != nil {
					util.Log().Debug("无法创建占位文件，%s", err)
				}
				return
			}

		}
	}
}

func sysError(err error) *RespError {
	return &RespError{APIError: APIError{
		Code:    "system",
		Message: err.Error(),
	}}
}

func (client *Client) request(ctx context.Context, method string, url string, body io.Reader, option ...request.Option) (string, error) {
	// 获取凭证
	err := client.UpdateCredential(ctx, conf.SystemConfig.Mode == "slave")
	if err != nil {
		return "", sysError(err)
	}

	option = append(option,
		request.WithHeader(http.Header{
			"Authorization": {"Bearer " + client.Credential.AccessToken},
			"Content-Type":  {"application/json"},
		}),
		request.WithContext(ctx),
		request.WithTPSLimit(
			fmt.Sprintf("policy_%d", client.Policy.ID),
			client.Policy.OptionsSerialized.TPSLimit,
			client.Policy.OptionsSerialized.TPSLimitBurst,
		),
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
			util.Log().Debug("Onedrive returns unknown response: %s", respBody)
			return "", sysError(decodeErr)
		}

		if res.Response.StatusCode == 429 {
			util.Log().Warning("OneDrive request is throttled.")
			return "", backoff.NewRetryableErrorFromHeader(&errResp, res.Response.Header)
		}

		return "", &errResp
	}

	return respBody, nil
}

func (client *Client) requestWithStr(ctx context.Context, method string, url string, body string, expectedCode int) (string, error) {
	// 发送请求
	bodyReader := io.NopCloser(strings.NewReader(body))
	return client.request(ctx, method, url, bodyReader,
		request.WithContentLength(int64(len(body))),
	)
}
