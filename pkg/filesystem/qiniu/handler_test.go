package qiniu

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{
		Policy: &model.Policy{
			MaxSize: 10,
			OptionsSerialized: model.PolicyOption{
				MimeType: "ss",
			},
			AccessKey: "ak",
			SecretKey: "sk",
			Server:    "http://test.com",
		},
	}
	ctx := context.Background()

	// 成功
	{
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		ctx = context.WithValue(ctx, fsctx.SavePathCtx, "/123")
		_, err := handler.Token(ctx, 10, "123")
		asserts.NoError(err)
	}

	// 上下文无存储路径
	{
		ctx = context.Background()
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		_, err := handler.Token(ctx, 10, "123")
		asserts.Error(err)
	}
}
