package wopi

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestDiscovery(t *testing.T) {
	a := assert.New(t)
	endpoint, _ := url.Parse("http://localhost:8001/hosting/discovery")
	client := &client{
		cache: cache.Store,
		http:  request.NewClient(),
		config: config{
			discoveryEndpoint: endpoint,
		},
	}

	a.NoError(client.refreshDiscovery())
	client.NewSession(nil, &model.File{Name: "123.pptx"}, ActionPreview)
}
