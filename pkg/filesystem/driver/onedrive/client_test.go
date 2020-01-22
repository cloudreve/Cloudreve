package onedrive

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	asserts := assert.New(t)
	// getOAuthEndpoint失败
	{
		policy := model.Policy{
			BaseURL: string([]byte{0x7f}),
		}
		res, err := NewClient(&policy)
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 成功
	{
		policy := model.Policy{}
		res, err := NewClient(&policy)
		asserts.NoError(err)
		asserts.NotNil(res)
		asserts.NotNil(res.Credential)
		asserts.NotNil(res.Endpoints)
		asserts.NotNil(res.Endpoints.OAuthEndpoints)
	}
}
