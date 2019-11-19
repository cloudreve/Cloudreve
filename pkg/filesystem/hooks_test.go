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
	file := local.FileData{
		Size: 5,
		Name: "1.txt",
	}
	ctx := context.WithValue(context.Background(), FileHeaderCtx, file)
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

	asserts.Error(GenericBeforeUpload(ctx, &fs))

	file.Size = 1
	file.Name = "1"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.Error(GenericBeforeUpload(ctx, &fs))

	file.Name = "1.txt"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.NoError(GenericBeforeUpload(ctx, &fs))

	file.Name = "1.t/xt"
	ctx = context.WithValue(context.Background(), FileHeaderCtx, file)
	asserts.Error(GenericBeforeUpload(ctx, &fs))
}
