package googledrive

import (
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"google.golang.org/api/drive/v3"
)

// Client Google Drive client
type Client struct {
	Endpoints  *Endpoints
	Policy     *model.Policy
	Credential *Credential

	ClientID     string
	ClientSecret string
	Redirect     string

	Request           request.Client
	ClusterController cluster.Controller
}

// Endpoints OneDrive客户端相关设置
type Endpoints struct {
	UserConsentEndpoint string // OAuth认证的基URL
	TokenEndpoint       string // OAuth token 基URL
	EndpointURL         string // 接口请求的基URL
}

const (
	TokenCachePrefix = "googledrive_"

	oauthEndpoint   = "https://oauth2.googleapis.com/token"
	userConsentBase = "https://accounts.google.com/o/oauth2/auth"
	v3DriveEndpoint = "https://www.googleapis.com/drive/v3"
)

var (
	// Defualt required scopes
	RequiredScope = []string{
		drive.DriveScope,
		"openid",
		"profile",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	// ErrInvalidRefreshToken 上传策略无有效的RefreshToken
	ErrInvalidRefreshToken = errors.New("no valid refresh token in this policy")
)

// NewClient 根据存储策略获取新的client
func NewClient(policy *model.Policy) (*Client, error) {
	client := &Client{
		Endpoints: &Endpoints{
			TokenEndpoint:       oauthEndpoint,
			UserConsentEndpoint: userConsentBase,
			EndpointURL:         v3DriveEndpoint,
		},
		Credential: &Credential{
			RefreshToken: policy.AccessKey,
		},
		Policy:            policy,
		ClientID:          policy.BucketName,
		ClientSecret:      policy.SecretKey,
		Redirect:          policy.OptionsSerialized.OauthRedirect,
		Request:           request.NewClient(),
		ClusterController: cluster.DefaultController,
	}

	return client, nil
}
