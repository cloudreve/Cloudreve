package remote

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestHandler_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{
		Policy: &model.Policy{
			MaxSize:      10,
			AutoRename:   true,
			DirNameRule:  "dir",
			FileNameRule: "file",
			OptionsSerialized: model.PolicyOption{
				FileType: []string{"txt"},
			},
		},
	}
	ctx := context.Background()
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 成功
	{
		cache.Set("setting_siteURL", "http://test.cloudreve.org", 0)
		credential, err := handler.Token(ctx, 10, "123")
		asserts.NoError(err)
		policy, err := serializer.DecodeUploadPolicy(credential.Policy)
		asserts.NoError(err)
		asserts.Equal(uint64(10), policy.MaxSize)
		asserts.Equal(true, policy.AutoRename)
		asserts.Equal("dir", policy.SavePath)
		asserts.Equal("file", policy.FileName)
		asserts.Equal([]string{"txt"}, policy.AllowedExtension)
	}

}

func TestHandler_Source(t *testing.T) {
	asserts := assert.New(t)
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 无法获取上下文
	{
		handler := Handler{}
		ctx := context.Background()
		res, err := handler.Source(ctx, "", url.URL{}, 0, true, 0)
		asserts.Error(err)
		asserts.Empty(res)
	}

	// 成功
	{
		handler := Handler{
			Policy: &model.Policy{Server: "/"},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, true, 0)
		asserts.NoError(err)
		asserts.Contains(res, "api/v3/slave/download/0")
	}

	// 成功 预览
	{
		handler := Handler{
			Policy: &model.Policy{Server: "/"},
		}
		file := model.File{
			SourceName: "1.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		res, err := handler.Source(ctx, "", url.URL{}, 10, false, 0)
		asserts.NoError(err)
		asserts.Contains(res, "api/v3/slave/source/0")
	}
}
