package filesystem

import (
	model "github.com/HFO4/cloudreve/models"
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
