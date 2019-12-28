package remote

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/stretchr/testify/assert"
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
