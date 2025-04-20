package onedrive

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/credmanager"
	"github.com/cloudreve/Cloudreve/v4/pkg/request"
	"github.com/samber/lo"
)

const (
	AccessTokenExpiryMargin = 600 // 10 minutes
)

// Error 实现error接口
func (err OAuthError) Error() string {
	return err.ErrorDescription
}

// OAuthURL 获取OAuth认证页面URL
func (client *client) OAuthURL(ctx context.Context, scope []string) string {
	query := url.Values{
		"client_id":     {client.policy.BucketName},
		"scope":         {strings.Join(scope, " ")},
		"response_type": {"code"},
		"redirect_uri":  {client.policy.Settings.OauthRedirect},
		"state":         {strconv.Itoa(client.policy.ID)},
	}
	client.endpoints.oAuthEndpoints.authorize.RawQuery = query.Encode()
	return client.endpoints.oAuthEndpoints.authorize.String()
}

// getOAuthEndpoint gets OAuth endpoints from API endpoint
func getOAuthEndpoint(apiEndpoint string) *oauthEndpoint {
	base, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil
	}
	var (
		token     *url.URL
		authorize *url.URL
	)
	switch base.Host {
	//case "login.live.com":
	//	token, _ = url.Parse("https://login.live.com/oauth20_token.srf")
	//	authorize, _ = url.Parse("https://login.live.com/oauth20_authorize.srf")
	case "microsoftgraph.chinacloudapi.cn":
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

// Credential 获取token时返回的凭证
type Credential struct {
	ExpiresIn       int64  `json:"expires_in"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	RefreshedAtUnix int64  `json:"refreshed_at"`

	PolicyID int `json:"policy_id"`
}

func init() {
	gob.Register(Credential{})
}

func (c Credential) Refresh(ctx context.Context) (credmanager.Credential, error) {
	if c.RefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	dep := dependency.FromContext(ctx)
	storagePolicyClient := dep.StoragePolicyClient()
	policy, err := storagePolicyClient.GetPolicyByID(ctx, c.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage policy: %w", err)
	}

	oauthBase := getOAuthEndpoint(policy.Server)

	newCredential, err := obtainToken(ctx, &obtainTokenArgs{
		clientId:      policy.BucketName,
		redirect:      policy.Settings.OauthRedirect,
		secret:        policy.SecretKey,
		refreshToken:  c.RefreshToken,
		client:        dep.RequestClient(request.WithLogger(dep.Logger())),
		tokenEndpoint: oauthBase.token.String(),
		policyID:      c.PolicyID,
	})

	if err != nil {
		return nil, err
	}

	c.RefreshToken = newCredential.RefreshToken
	c.AccessToken = newCredential.AccessToken
	c.ExpiresIn = newCredential.ExpiresIn
	c.RefreshedAtUnix = time.Now().Unix()

	// Write refresh token to db
	if err := storagePolicyClient.UpdateAccessKey(ctx, policy, newCredential.RefreshToken); err != nil {
		return nil, err
	}

	return c, nil
}

func (c Credential) Key() string {
	return CredentialKey(c.PolicyID)
}

func (c Credential) Expiry() time.Time {
	return time.Unix(c.ExpiresIn-AccessTokenExpiryMargin, 0)
}

func (c Credential) String() string {
	return c.AccessToken
}

func (c Credential) RefreshedAt() *time.Time {
	if c.RefreshedAtUnix == 0 {
		return nil
	}
	refreshedAt := time.Unix(c.RefreshedAtUnix, 0)
	return &refreshedAt
}

// ObtainToken 通过code或refresh_token兑换token
func (client *client) ObtainToken(ctx context.Context, opts ...Option) (*Credential, error) {
	options := newDefaultOption()
	for _, o := range opts {
		o.apply(options)
	}

	return obtainToken(ctx, &obtainTokenArgs{
		clientId:      client.policy.BucketName,
		redirect:      client.policy.Settings.OauthRedirect,
		secret:        client.policy.SecretKey,
		code:          options.code,
		refreshToken:  options.refreshToken,
		client:        client.httpClient,
		tokenEndpoint: client.endpoints.oAuthEndpoints.token.String(),
		policyID:      client.policy.ID,
	})

}

type obtainTokenArgs struct {
	clientId      string
	redirect      string
	secret        string
	code          string
	refreshToken  string
	client        request.Client
	tokenEndpoint string
	policyID      int
}

// obtainToken fetch new access token from Microsoft Graph API
func obtainToken(ctx context.Context, args *obtainTokenArgs) (*Credential, error) {
	body := url.Values{
		"client_id":     {args.clientId},
		"redirect_uri":  {args.redirect},
		"client_secret": {args.secret},
	}
	if args.code != "" {
		body.Add("grant_type", "authorization_code")
		body.Add("code", args.code)
	} else {
		body.Add("grant_type", "refresh_token")
		body.Add("refresh_token", args.refreshToken)
	}
	strBody := body.Encode()

	res := args.client.Request(
		"POST",
		args.tokenEndpoint,
		io.NopCloser(strings.NewReader(strBody)),
		request.WithHeader(http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"}},
		),
		request.WithContentLength(int64(len(strBody))),
		request.WithContext(ctx),
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

	credential.PolicyID = args.policyID
	credential.ExpiresIn = time.Now().Unix() + credential.ExpiresIn
	if args.code != "" {
		credential.ExpiresIn = time.Now().Unix() - 10
	}
	return &credential, nil
}

// UpdateCredential 更新凭证，并检查有效期
func (client *client) UpdateCredential(ctx context.Context) error {
	newCred, err := client.cred.Obtain(ctx, CredentialKey(client.policy.ID))
	if err != nil {
		return fmt.Errorf("failed to obtain token from CredManager: %w", err)
	}

	client.credential = newCred
	return nil
}

// RetrieveOneDriveCredentials retrieves OneDrive credentials from DB inventory
func RetrieveOneDriveCredentials(ctx context.Context, storagePolicyClient inventory.StoragePolicyClient) ([]credmanager.Credential, error) {
	odPolicies, err := storagePolicyClient.ListPolicyByType(ctx, types.PolicyTypeOd)
	if err != nil {
		return nil, fmt.Errorf("failed to list OneDrive policies: %w", err)
	}

	return lo.Map(odPolicies, func(item *ent.StoragePolicy, index int) credmanager.Credential {
		return &Credential{
			PolicyID:     item.ID,
			ExpiresIn:    0,
			RefreshToken: item.AccessKey,
		}
	}), nil
}

func CredentialKey(policyId int) string {
	return fmt.Sprintf("cred_od_%d", policyId)
}
