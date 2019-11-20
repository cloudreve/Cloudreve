package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/stretchr/testify/assert"
	"os"
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

func TestGenericAfterUploadCanceled(t *testing.T) {
	asserts := assert.New(t)
	f, err := os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()
	file := local.FileStream{
		Size: 5,
		Name: "TestGenericAfterUploadCanceled",
	}
	ctx := context.WithValue(context.Background(), SavePathCtx, "TestGenericAfterUploadCanceled")
	ctx = context.WithValue(ctx, FileHeaderCtx, file)
	fs := FileSystem{
		User:    &model.User{Storage: 5},
		Handler: local.Handler{},
	}

	// 成功
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.NoError(err)
	asserts.Equal(uint64(0), fs.User.Storage)

	f, err = os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()

	// 容量不能再降低
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.Error(err)

	//文件不存在
	fs.User.Storage = 5
	err = GenericAfterUploadCanceled(ctx, &fs)
	asserts.NoError(err)
}

//func TestGenericAfterUpload(t *testing.T) {
//	asserts := assert.New(t)
//	testObj := FileSystem{}
//	ctx := context.WithValue(context.Background(),FileHeaderCtx,local.FileStream{
//		VirtualPath: "/我的文件",
//		Name:        "test.txt",
//	})
//
//
//}
