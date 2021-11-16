package onedrive

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
)

var (
	// ErrAuthEndpoint 无法解析授权端点地址
	ErrAuthEndpoint = errors.New("无法解析授权端点地址")
	// ErrInvalidRefreshToken 上传策略无有效的RefreshToken
	ErrInvalidRefreshToken = errors.New("上传策略无有效的RefreshToken")
	// ErrDeleteFile 无法删除文件
	ErrDeleteFile = errors.New("无法删除文件")
	// ErrClientCanceled 客户端取消操作
	ErrClientCanceled = errors.New("客户端取消操作")
)

// Client OneDrive客户端
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
	OAuthURL       string // OAuth认证的基URL
	OAuthEndpoints *oauthEndpoint
	EndpointURL    string // 接口请求的基URL
	isInChina      bool   // 是否为世纪互联
	DriverResource string // 要使用的驱动器
}

// NewClient 根据存储策略获取新的client
func NewClient(policy *model.Policy) (*Client, error) {
	client := &Client{
		Endpoints: &Endpoints{
			OAuthURL:       policy.BaseURL,
			EndpointURL:    policy.Server,
			DriverResource: policy.OptionsSerialized.OdDriver,
		},
		Credential: &Credential{
			RefreshToken: policy.AccessKey,
		},
		Policy:            policy,
		ClientID:          policy.BucketName,
		ClientSecret:      policy.SecretKey,
		Redirect:          policy.OptionsSerialized.OdRedirect,
		Request:           request.NewClient(),
		ClusterController: cluster.DefaultController,
	}

	if client.Endpoints.DriverResource == "" {
		client.Endpoints.DriverResource = "me/drive"
	}

	oauthBase := client.getOAuthEndpoint()
	if oauthBase == nil {
		return nil, ErrAuthEndpoint
	}
	client.Endpoints.OAuthEndpoints = oauthBase

	return client, nil
}
