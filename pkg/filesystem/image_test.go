package filesystem

import (
	"context"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestFileSystem_GetThumb(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{}}

	// 非图像文件
	{
		fs.SetTargetFile(&[]model.File{{}})
		_, err := fs.GetThumb(context.Background(), 1)
		asserts.Equal(err, ErrObjectNotExist)
	}

	// 成功
	{
		cache.Set("setting_thumb_width", "10", 0)
		cache.Set("setting_thumb_height", "10", 0)
		cache.Set("setting_preview_timeout", "50", 0)
		testHandller2 := new(FileHeaderMock)
		testHandller2.On("Thumb", testMock.Anything, "").Return(&response.ContentResponse{}, nil)
		fs.CleanTargets()
		fs.SetTargetFile(&[]model.File{{PicInfo: "1,1", Policy: model.Policy{Type: "mock"}}})
		fs.FileTarget[0].Policy.ID = 1
		fs.Handler = testHandller2
		res, err := fs.GetThumb(context.Background(), 1)
		asserts.NoError(err)
		asserts.EqualValues(50, res.MaxAge)
	}
}
