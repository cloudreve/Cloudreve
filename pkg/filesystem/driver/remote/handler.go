package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Driver 远程存储策略适配器
type Driver struct {
	Client       request.Client
	Policy       *model.Policy
	AuthInstance auth.Auth

	uploadClient Client
}

// NewDriver initializes a new Driver from policy
// TODO: refactor all method into upload client
func NewDriver(policy *model.Policy) (*Driver, error) {
	client, err := NewClient(policy)
	if err != nil {
		return nil, err
	}

	return &Driver{
		Policy:       policy,
		Client:       request.NewClient(),
		AuthInstance: auth.HMACAuth{[]byte(policy.SecretKey)},
		uploadClient: client,
	}, nil
}

// List 列取文件
func (handler *Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	var res []response.Object

	reqBody := serializer.ListRequest{
		Path:      path,
		Recursive: recursive,
	}
	reqBodyEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return res, err
	}

	// 发送列表请求
	bodyReader := strings.NewReader(string(reqBodyEncoded))
	signTTL := model.GetIntSetting("slave_api_timeout", 60)
	resp, err := handler.Client.Request(
		"POST",
		handler.getAPIUrl("list"),
		bodyReader,
		request.WithCredential(handler.AuthInstance, int64(signTTL)),
		request.WithMasterMeta(),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return res, err
	}

	// 处理列取结果
	if resp.Code != 0 {
		return res, errors.New(resp.Error)
	}

	if resStr, ok := resp.Data.(string); ok {
		err = json.Unmarshal([]byte(resStr), &res)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

// getAPIUrl 获取接口请求地址
func (handler *Driver) getAPIUrl(scope string, routes ...string) string {
	serverURL, err := url.Parse(handler.Policy.Server)
	if err != nil {
		return ""
	}
	var controller *url.URL

	switch scope {
	case "delete":
		controller, _ = url.Parse("/api/v3/slave/delete")
	case "thumb":
		controller, _ = url.Parse("/api/v3/slave/thumb")
	case "list":
		controller, _ = url.Parse("/api/v3/slave/list")
	default:
		controller = serverURL
	}

	for _, r := range routes {
		controller.Path = path.Join(controller.Path, r)
	}

	return serverURL.ResolveReference(controller).String()
}

// Get 获取文件内容
func (handler *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 尝试获取速度限制
	speedLimit := 0
	if user, ok := ctx.Value(fsctx.UserCtx).(model.User); ok {
		speedLimit = user.Group.SpeedLimit
	}

	// 获取文件源地址
	downloadURL, err := handler.Source(ctx, path, 0, true, speedLimit)
	if err != nil {
		return nil, err
	}

	// 获取文件数据流
	resp, err := handler.Client.Request(
		"GET",
		downloadURL,
		nil,
		request.WithContext(ctx),
		request.WithTimeout(time.Duration(0)),
		request.WithMasterMeta(),
	).CheckHTTPResponse(200).GetRSCloser()
	if err != nil {
		return nil, err
	}

	resp.SetFirstFakeChunk()

	// 尝试获取文件大小
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		resp.SetContentLength(int64(file.Size))
	}

	return resp, nil
}

// Put 将文件流保存到指定目录
func (handler *Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	defer file.Close()

	return handler.uploadClient.Upload(ctx, file)
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	// 封装接口请求正文
	reqBody := serializer.RemoteDeleteRequest{
		Files: files,
	}
	reqBodyEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return files, err
	}

	// 发送删除请求
	bodyReader := strings.NewReader(string(reqBodyEncoded))
	signTTL := model.GetIntSetting("slave_api_timeout", 60)
	resp, err := handler.Client.Request(
		"POST",
		handler.getAPIUrl("delete"),
		bodyReader,
		request.WithCredential(handler.AuthInstance, int64(signTTL)),
		request.WithMasterMeta(),
		request.WithSlaveMeta(handler.Policy.AccessKey),
	).CheckHTTPResponse(200).GetResponse()
	if err != nil {
		return files, err
	}

	// 处理删除结果
	var reqResp serializer.Response
	err = json.Unmarshal([]byte(resp), &reqResp)
	if err != nil {
		return files, err
	}
	if reqResp.Code != 0 {
		var failedResp serializer.RemoteDeleteRequest
		if failed, ok := reqResp.Data.(string); ok {
			err = json.Unmarshal([]byte(failed), &failedResp)
			if err == nil {
				return failedResp.Files, errors.New(reqResp.Error)
			}
		}
		return files, errors.New("unknown format of returned response")
	}

	return []string{}, nil
}

// Thumb 获取文件缩略图
func (handler *Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	// quick check by extension name
	supported := []string{"png", "jpg", "jpeg", "gif"}
	if len(handler.Policy.OptionsSerialized.ThumbExts) > 0 {
		supported = handler.Policy.OptionsSerialized.ThumbExts
	}

	if !util.IsInExtensionList(supported, file.Name) {
		return nil, driver.ErrorThumbNotSupported
	}

	sourcePath := base64.RawURLEncoding.EncodeToString([]byte(file.SourceName))
	thumbURL := fmt.Sprintf("%s/%s/%s", handler.getAPIUrl("thumb"), sourcePath, filepath.Ext(file.Name))
	ttl := model.GetIntSetting("preview_timeout", 60)
	signedThumbURL, err := auth.SignURI(handler.AuthInstance, thumbURL, int64(ttl))
	if err != nil {
		return nil, err
	}

	return &response.ContentResponse{
		Redirect: true,
		URL:      signedThumbURL.String(),
	}, nil
}

// Source 获取外链URL
func (handler *Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {
	// 尝试从上下文获取文件名
	fileName := "file"
	if file, ok := ctx.Value(fsctx.FileModelCtx).(model.File); ok {
		fileName = file.Name
	}

	serverURL, err := url.Parse(handler.Policy.Server)
	if err != nil {
		return "", errors.New("无法解析远程服务端地址")
	}

	// 是否启用了CDN
	if handler.Policy.BaseURL != "" {
		cdnURL, err := url.Parse(handler.Policy.BaseURL)
		if err != nil {
			return "", err
		}
		serverURL = cdnURL
	}

	var (
		signedURI  *url.URL
		controller = "/api/v3/slave/download"
	)
	if !isDownload {
		controller = "/api/v3/slave/source"
	}

	// 签名下载地址
	sourcePath := base64.RawURLEncoding.EncodeToString([]byte(path))
	signedURI, err = auth.SignURI(
		handler.AuthInstance,
		fmt.Sprintf("%s/%d/%s/%s", controller, speed, sourcePath, url.PathEscape(fileName)),
		ttl,
	)

	if err != nil {
		return "", serializer.NewError(serializer.CodeEncryptError, "Failed to sign URL", err)
	}

	finalURL := serverURL.ResolveReference(signedURI).String()
	return finalURL, nil

}

// Token 获取上传策略和认证Token
func (handler *Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse(path.Join("/api/v3/callback/remote", uploadSession.Key, uploadSession.CallbackSecret))
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 在从机端创建上传会话
	uploadSession.Callback = apiURL.String()
	if err := handler.uploadClient.CreateUploadSession(ctx, uploadSession, ttl, false); err != nil {
		return nil, err
	}

	// 获取上传地址
	uploadURL, sign, err := handler.uploadClient.GetUploadURL(ttl, uploadSession.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign upload url: %w", err)
	}

	return &serializer.UploadCredential{
		SessionID:  uploadSession.Key,
		ChunkSize:  handler.Policy.OptionsSerialized.ChunkSize,
		UploadURLs: []string{uploadURL},
		Credential: sign,
	}, nil
}

// 取消上传凭证
func (handler *Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	return handler.uploadClient.DeleteUploadSession(ctx, uploadSession.Key)
}
