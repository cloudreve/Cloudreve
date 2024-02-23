package vol

import (
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"net/http"
)

var ClientSecret = ""

const CRMSite = "https://pro.cloudreve.org/crm/api/vol/"

type Client interface {
	// Sync VOL from CRM, return content (base64 encoded) and signature.
	Sync() (string, string, error)
}

type VolClient struct {
	secret string
	client request.Client
}

func New(secret string) Client {
	return &VolClient{secret: secret, client: request.NewClient()}
}

func (c *VolClient) Sync() (string, string, error) {
	res, err := c.client.Request("GET", CRMSite+c.secret, nil).CheckHTTPResponse(http.StatusOK).DecodeResponse()
	if err != nil {
		return "", "", fmt.Errorf("failed to get VOL from CRM: %w", err)
	}

	if res.Code != 0 {
		return "", "", fmt.Errorf("CRM return error: %s", res.Msg)
	}

	vol := res.Data.(map[string]interface{})
	return vol["content"].(string), vol["signature"].(string), nil
}
