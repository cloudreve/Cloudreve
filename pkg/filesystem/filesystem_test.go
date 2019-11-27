package filesystem

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"

	"testing"
)

func TestNewFileSystem(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{
			Type: "local",
		},
	}

	fs, err := NewFileSystem(&user)
	asserts.NoError(err)
	asserts.NotNil(fs.Handler)

	user.Policy.Type = "unknown"
	fs, err = NewFileSystem(&user)
	asserts.Error(err)
}

func TestNewFileSystemFromContext(t *testing.T) {
	asserts := assert.New(t)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user", &model.User{
		Policy: model.Policy{
			Type: "local",
		},
	})
	fs, err := NewFileSystemFromContext(c)
	asserts.NotNil(fs)
	asserts.NoError(err)

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	fs, err = NewFileSystemFromContext(c)
	asserts.Nil(fs)
	asserts.Error(err)
}

func TestDispatchHandler(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{
		User: &model.User{Policy: model.Policy{
			Type: "local",
		}},
	}

	// 未指定，使用用户默认
	err := fs.dispatchHandler()
	asserts.NoError(err)
	asserts.IsType(local.Handler{}, fs.Handler)

	// 已指定，发生错误
	fs.Policy = &model.Policy{Type: "unknown"}
	err = fs.dispatchHandler()
	asserts.Error(err)
}
