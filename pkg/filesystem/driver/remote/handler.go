package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Driver 远程存储策略适配器
type Driver struct {
	Client       request.Client
	Policy       *model.Policy
	AuthInstance auth.Auth
}

// List 列取文件
func (handler Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
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
func (handler Driver) getAPIUrl(scope string, routes ...string) string {
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
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 尝试获取速度限制
	speedLimit := 0
	if user, ok := ctx.Value(fsctx.UserCtx).(model.User); ok {
		speedLimit = user.Group.SpeedLimit
	}

	// 获取文件源地址
	downloadURL, err := handler.Source(ctx, path, url.URL{}, 0, true, speedLimit)
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
func (handler Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	defer file.Close()

	// 凭证有效期
	credentialTTL := model.GetIntSetting("upload_credential_timeout", 3600)

	// 生成上传策略
	policy := serializer.UploadPolicy{
		SavePath:   path.Dir(dst),
		FileName:   path.Base(dst),
		AutoRename: false,
		MaxSize:    size,
	}
	credential, err := handler.getUploadCredential(ctx, policy, int64(credentialTTL))
	if err != nil {
		return err
	}

	// 对文件名进行URLEncode
	fileName := url.QueryEscape(path.Base(dst))

	// 决定是否要禁用文件覆盖
	overwrite := "true"
	if ctx.Value(fsctx.DisableOverwrite) != nil {
		overwrite = "false"
	}

	// 上传文件
	resp, err := handler.Client.Request(
		"POST",
		handler.Policy.GetUploadURL(),
		file,
		request.WithHeader(map[string][]string{
			"X-Cr-Policy":    {credential.Policy},
			"X-Cr-FileName":  {fileName},
			"X-Cr-Overwrite": {overwrite},
		}),
		request.WithContentLength(int64(size)),
		request.WithTimeout(time.Duration(0)),
		request.WithMasterMeta(),
		request.WithSlaveMeta(handler.Policy.AccessKey),
		request.WithCredential(handler.AuthInstance, int64(credentialTTL)),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}

// Delete 删除一个或多个文件，
// 返回未删除的文件，及遇到的最后一个错误
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
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
		return files, errors.New("未知的返回结果格式")
	}

	return []string{}, nil
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	sourcePath := base64.RawURLEncoding.EncodeToString([]byte(path))
	thumbURL := handler.getAPIUrl("thumb") + "/" + sourcePath
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
func (handler Driver) Source(
	ctx context.Context,
	path string,
	baseURL url.URL,
	ttl int64,
	isDownload bool,
	speed int,
) (string, error) {
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
		return "", serializer.NewError(serializer.CodeEncryptError, "无法对URL进行签名", err)
	}

	finalURL := serverURL.ResolveReference(signedURI).String()
	return finalURL, nil

}

// Token 获取上传策略和认证Token
func (handler Driver) Token(ctx context.Context, TTL int64, key string) (serializer.UploadCredential, error) {
	// 生成回调地址
	siteURL := model.GetSiteURL()
	apiBaseURI, _ := url.Parse("/api/v3/callback/remote/" + key)
	apiURL := siteURL.ResolveReference(apiBaseURI)

	// 生成上传策略
	policy := serializer.UploadPolicy{
		SavePath:         handler.Policy.DirNameRule,
		FileName:         handler.Policy.FileNameRule,
		AutoRename:       handler.Policy.AutoRename,
		MaxSize:          handler.Policy.MaxSize,
		AllowedExtension: handler.Policy.OptionsSerialized.FileType,
		CallbackURL:      apiURL.String(),
	}
	return handler.getUploadCredential(ctx, policy, TTL)
}

func (handler Driver) getUploadCredential(ctx context.Context, policy serializer.UploadPolicy, TTL int64) (serializer.UploadCredential, error) {
	policyEncoded, err := policy.EncodeUploadPolicy()
	if err != nil {
		return serializer.UploadCredential{}, err
	}

	// 签名上传策略
	uploadRequest, _ := http.NewRequest("POST", "/api/v3/slave/upload", nil)
	uploadRequest.Header = map[string][]string{
		"X-Cr-Policy":    {policyEncoded},
		"X-Cr-Overwrite": {"false"},
	}
	auth.SignRequest(handler.AuthInstance, uploadRequest, TTL)

	if credential, ok := uploadRequest.Header["Authorization"]; ok && len(credential) == 1 {
		return serializer.UploadCredential{
			Token:  credential[0],
			Policy: policyEncoded,
		}, nil
	}
	return serializer.UploadCredential{}, errors.New("无法签名上传策略")
}
