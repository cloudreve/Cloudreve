package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenericBeforeUpload(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	file := local.FileData{
		Size: 5,
		Name: "1.txt",
	}
	fs := FileSystem{
		User: &model.User{
			Storage: 0,
			Group: model.Group{
				MaxStorage: 11,
			},
			Policy: model.Policy{
				MaxSize: 4,
				OptionsSerialized: model.PolicyOption{
					FileType: []string{"txt"},
				},
			},
		},
	}

	asserts.Error(GenericBeforeUpload(ctx, &fs, file))
	file.Size = 1
	file.Name = "1"
	asserts.Error(GenericBeforeUpload(ctx, &fs, file))
	file.Name = "1.txt"
	asserts.NoError(GenericBeforeUpload(ctx, &fs, file))
}
