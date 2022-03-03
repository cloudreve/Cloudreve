package remote

import (
	"context"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	basePath        = "/api/v3/slave"
	OverwriteHeader = auth.CrHeaderPrefix + "Overwrite"
)

// Client to operate remote slave server
type Client interface {
	// CreateUploadSession creates remote upload session
	CreateUploadSession(ctx context.Context, session *serializer.UploadSession, ttl int64) error
	// GetUploadURL signs an url for uploading file
	GetUploadURL(ttl int64, sessionID string) (string, string, error)
}

// NewClient creates new Client from given policy
func NewClient(policy *model.Policy) (Client, error) {
	authInstance := auth.HMACAuth{[]byte(policy.SecretKey)}
	serverURL, err := url.Parse(policy.Server)
	if err != nil {
		return nil, err
	}

	base, _ := url.Parse(basePath)
	signTTL := model.GetIntSetting("slave_api_timeout", 60)

	return &remoteClient{
		policy:       policy,
		authInstance: authInstance,
		httpClient: request.NewClient(
			request.WithEndpoint(serverURL.ResolveReference(base).String()),
			request.WithCredential(authInstance, int64(signTTL)),
			request.WithMasterMeta(),
		),
	}, nil
}

type remoteClient struct {
	policy       *model.Policy
	authInstance auth.Auth
	httpClient   request.Client
}

func (c *remoteClient) CreateUploadSession(ctx context.Context, session *serializer.UploadSession, ttl int64) error {
	reqBodyEncoded, err := json.Marshal(map[string]interface{}{
		"session": session,
		"ttl":     ttl,
	})
	if err != nil {
		return err
	}

	bodyReader := strings.NewReader(string(reqBodyEncoded))
	resp, err := c.httpClient.Request(
		"PUT",
		"upload",
		bodyReader,
		request.WithContext(ctx),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return serializer.NewErrorFromResponse(resp)
	}

	return nil
}

func (c *remoteClient) GetUploadURL(ttl int64, sessionID string) (string, string, error) {
	base, err := url.Parse(c.policy.Server)
	if err != nil {
		return "", "", err
	}

	base.Path = path.Join(base.Path, basePath, "upload", sessionID)
	req, err := http.NewRequest("POST", base.String(), nil)
	if err != nil {
		return "", "", err
	}

	req = auth.SignRequest(c.authInstance, req, ttl)
	return req.URL.String(), req.Header["Authorization"][0], nil
}
