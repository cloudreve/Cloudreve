package googledrive

import (
	"context"
	"encoding/json"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/oauth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthURL 获取OAuth认证页面URL
func (client *Client) OAuthURL(ctx context.Context, scope []string) string {
	query := url.Values{
		"client_id":     {client.ClientID},
		"scope":         {strings.Join(scope, " ")},
		"response_type": {"code"},
		"redirect_uri":  {client.Redirect},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}

	u, _ := url.Parse(client.Endpoints.UserConsentEndpoint)
	u.RawQuery = query.Encode()
	return u.String()
}

// ObtainToken 通过code或refresh_token兑换token
func (client *Client) ObtainToken(ctx context.Context, code, refreshToken string) (*Credential, error) {
	body := url.Values{
		"client_id":     {client.ClientID},
		"redirect_uri":  {client.Redirect},
		"client_secret": {client.ClientSecret},
	}
	if code != "" {
		body.Add("grant_type", "authorization_code")
		body.Add("code", code)
	} else {
		body.Add("grant_type", "refresh_token")
		body.Add("refresh_token", refreshToken)
	}
	strBody := body.Encode()

	res := client.Request.Request(
		"POST",
		client.Endpoints.TokenEndpoint,
		io.NopCloser(strings.NewReader(strBody)),
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
func (client *Client) UpdateCredential(ctx context.Context, isSlave bool) error {
	if isSlave {
		return client.fetchCredentialFromMaster(ctx)
	}

	oauth.GlobalMutex.Lock(client.Policy.ID)
	defer oauth.GlobalMutex.Unlock(client.Policy.ID)

	// 如果已存在凭证
	if client.Credential != nil && client.Credential.AccessToken != "" {
		// 检查已有凭证是否过期
		if client.Credential.ExpiresIn > time.Now().Unix() {
			// 未过期，不要更新
			return nil
		}
	}

	// 尝试从缓存中获取凭证
	if cacheCredential, ok := cache.Get(TokenCachePrefix + client.ClientID); ok {
		credential := cacheCredential.(Credential)
		if credential.ExpiresIn > time.Now().Unix() {
			client.Credential = &credential
			return nil
		}
	}

	// 获取新的凭证
	if client.Credential == nil || client.Credential.RefreshToken == "" {
		// 无有效的RefreshToken
		util.Log().Error("Failed to refresh credential for policy %q, please login your Google account again.", client.Policy.Name)
		return ErrInvalidRefreshToken
	}

	credential, err := client.ObtainToken(ctx, "", client.Credential.RefreshToken)
	if err != nil {
		return err
	}

	// 更新有效期为绝对时间戳
	expires := credential.ExpiresIn - 60
	credential.ExpiresIn = time.Now().Add(time.Duration(expires) * time.Second).Unix()
	// refresh token for Google Drive does not expire in production
	credential.RefreshToken = client.Credential.RefreshToken
	client.Credential = credential

	// 更新缓存
	cache.Set(TokenCachePrefix+client.ClientID, *credential, int(expires))

	return nil
}

func (client *Client) AccessToken() string {
	return client.Credential.AccessToken
}

// UpdateCredential 更新凭证，并检查有效期
func (client *Client) fetchCredentialFromMaster(ctx context.Context) error {
	res, err := client.ClusterController.GetPolicyOauthToken(client.Policy.MasterID, client.Policy.ID)
	if err != nil {
		return err
	}

	client.Credential = &Credential{AccessToken: res}
	return nil
}
