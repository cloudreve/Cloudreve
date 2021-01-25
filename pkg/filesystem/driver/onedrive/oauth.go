package onedrive

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Error 实现error接口
func (err OAuthError) Error() string {
	return err.ErrorDescription
}

// OAuthURL 获取OAuth认证页面URL
func (client *Client) OAuthURL(ctx context.Context, scope []string) string {
	query := url.Values{
		"client_id":     {client.ClientID},
		"scope":         {strings.Join(scope, " ")},
		"response_type": {"code"},
		"redirect_uri":  {client.Redirect},
	}
	client.Endpoints.OAuthEndpoints.authorize.RawQuery = query.Encode()
	return client.Endpoints.OAuthEndpoints.authorize.String()
}

// getOAuthEndpoint 根据指定的AuthURL获取详细的认证接口地址
func (client *Client) getOAuthEndpoint() *oauthEndpoint {
	base, err := url.Parse(client.Endpoints.OAuthURL)
	if err != nil {
		return nil
	}
	var (
		token     *url.URL
		authorize *url.URL
	)
	switch base.Host {
	case "login.live.com":
		token, _ = url.Parse("https://login.live.com/oauth20_token.srf")
		authorize, _ = url.Parse("https://login.live.com/oauth20_authorize.srf")
	case "login.chinacloudapi.cn":
		client.Endpoints.isInChina = true
		token, _ = url.Parse("https://login.chinacloudapi.cn/common/oauth2/v2.0/token")
		authorize, _ = url.Parse("https://login.chinacloudapi.cn/common/oauth2/v2.0/authorize")
	default:
		token, _ = url.Parse("https://login.microsoftonline.com/common/oauth2/v2.0/token")
		authorize, _ = url.Parse("https://login.microsoftonline.com/common/oauth2/v2.0/authorize")
	}

	return &oauthEndpoint{
		token:     *token,
		authorize: *authorize,
	}
}

// ObtainToken 通过code或refresh_token兑换token
func (client *Client) ObtainToken(ctx context.Context, opts ...Option) (*Credential, error) {
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	body := url.Values{
		"client_id":     {client.ClientID},
		"redirect_uri":  {client.Redirect},
		"client_secret": {client.ClientSecret},
	}
	if options.code != "" {
		body.Add("grant_type", "authorization_code")
		body.Add("code", options.code)
	} else {
		body.Add("grant_type", "refresh_token")
		body.Add("refresh_token", options.refreshToken)
	}
	strBody := body.Encode()

	res := client.Request.Request(
		"POST",
		client.Endpoints.OAuthEndpoints.token.String(),
		ioutil.NopCloser(strings.NewReader(strBody)),
		request.WithHeader(http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"}},
		),
		request.WithContentLength(int64(len(strBody))),
	)
	if res.Err != nil {
		return nil, res.Err
	}

	respBody, err := res.GetResponse()
	if err != nil {
		return nil, err
	}

	var (
		errResp    OAuthError
		credential Credential
		decodeErr  error
	)

	if res.Response.StatusCode != 200 {
		decodeErr = json.Unmarshal([]byte(respBody), &errResp)
	} else {
		decodeErr = json.Unmarshal([]byte(respBody), &credential)
	}
	if decodeErr != nil {
		return nil, decodeErr
	}

	if errResp.ErrorType != "" {
		return nil, errResp
	}

	return &credential, nil

}

// UpdateCredential 更新凭证，并检查有效期
func (client *Client) UpdateCredential(ctx context.Context) error {
	// 如果已存在凭证
	if client.Credential != nil && client.Credential.AccessToken != "" {
		// 检查已有凭证是否过期
		if client.Credential.ExpiresIn > time.Now().Unix() {
			// 未过期，不要更新
			return nil
		}
	}

	// 尝试从缓存中获取凭证
	if cacheCredential, ok := cache.Get("onedrive_" + client.ClientID); ok {
		credential := cacheCredential.(Credential)
		if credential.ExpiresIn > time.Now().Unix() {
			client.Credential = &credential
			return nil
		}
	}

	// 获取新的凭证
	if client.Credential == nil || client.Credential.RefreshToken == "" {
		// 无有效的RefreshToken
		util.Log().Error("上传策略[%s]凭证刷新失败，请重新授权OneDrive账号", client.Policy.Name)
		return ErrInvalidRefreshToken
	}

	credential, err := client.ObtainToken(ctx, WithRefreshToken(client.Credential.RefreshToken))
	if err != nil {
		return err
	}

	// 更新有效期为绝对时间戳
	expires := credential.ExpiresIn - 60
	credential.ExpiresIn = time.Now().Add(time.Duration(expires) * time.Second).Unix()
	client.Credential = credential

	// 更新存储策略的 RefreshToken
	client.Policy.UpdateAccessKey(credential.RefreshToken)

	// 更新缓存
	cache.Set("onedrive_"+client.ClientID, *credential, int(expires))

	return nil
}
