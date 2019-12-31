package local

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestHandler_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()

	testCases := []struct {
		file io.ReadCloser
		dst  string
		err  bool
	}{
		{
			file: ioutil.NopCloser(strings.NewReader("test input file")),
			dst:  "test/test/txt",
			err:  false,
		},
		{
			file: ioutil.NopCloser(strings.NewReader("test input file")),
			dst:  "/notexist:/S.TXT",
			err:  true,
		},
	}

	for _, testCase := range testCases {
		err := handler.Put(ctx, testCase.file, testCase.dst, 15)
		if testCase.err {
			asserts.Error(err)
		} else {
			asserts.NoError(err)
			asserts.True(util.Exists(testCase.dst))
		}
	}
}

func TestHandler_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()

	file, err := os.Create("test.file")
	asserts.NoError(err)
	_ = file.Close()
	list, err := handler.Delete(ctx, []string{"test.file"})
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	file, err = os.Create("test.file")
	asserts.NoError(err)
	_ = file.Close()
	list, err = handler.Delete(ctx, []string{"test.file", "test.notexist"})
	asserts.Equal([]string{"test.notexist"}, list)
	asserts.Error(err)

	list, err = handler.Delete(ctx, []string{"test.notexist"})
	asserts.Equal([]string{"test.notexist"}, list)
	asserts.Error(err)
}

func TestHandler_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 成功
	file, err := os.Create("TestHandler_Get.txt")
	asserts.NoError(err)
	_ = file.Close()

	rs, err := handler.Get(ctx, "TestHandler_Get.txt")
	asserts.NoError(err)
	asserts.NotNil(rs)

	// 文件不存在

	rs, err = handler.Get(ctx, "TestHandler_Get_notExist.txt")
	asserts.Error(err)
	asserts.Nil(rs)
}

func TestHandler_Thumb(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()
	file, err := os.Create("TestHandler_Thumb" + conf.ThumbConfig.FileSuffix)
	asserts.NoError(err)
	file.Close()

	// 正常
	{
		thumb, err := handler.Thumb(ctx, "TestHandler_Thumb")
		asserts.NoError(err)
		asserts.NotNil(thumb.Content)
	}

	// 不存在
	{
		_, err := handler.Thumb(ctx, "not_exist")
		asserts.Error(err)
	}
}

func TestHandler_Source(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 成功
	{
		file := model.File{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test.jpg",
		}
		ctx := context.WithValue(ctx, fsctx.FileModelCtx, file)
		baseURL, err := url.Parse("https://cloudreve.org")
		asserts.NoError(err)
		sourceURL, err := handler.Source(ctx, "", *baseURL, 0, false, 0)
		asserts.NoError(err)
		asserts.NotEmpty(sourceURL)
		asserts.Contains(sourceURL, "sign=")
		asserts.Contains(sourceURL, "https://cloudreve.org")
	}

	// 无法获取上下文
	{
		baseURL, err := url.Parse("https://cloudreve.org")
		asserts.NoError(err)
		sourceURL, err := handler.Source(ctx, "", *baseURL, 0, false, 0)
		asserts.Error(err)
		asserts.Empty(sourceURL)
	}
}

func TestHandler_GetDownloadURL(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()
	auth.General = auth.HMACAuth{SecretKey: []byte("test")}

	// 成功
	{
		file := model.File{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test.jpg",
		}
		ctx := context.WithValue(ctx, fsctx.FileModelCtx, file)
		baseURL, err := url.Parse("https://cloudreve.org")
		asserts.NoError(err)
		downloadURL, err := handler.Source(ctx, "", *baseURL, 10, true, 0)
		asserts.NoError(err)
		asserts.Contains(downloadURL, "sign=")
		asserts.Contains(downloadURL, "https://cloudreve.org")
	}

	// 无法获取上下文
	{
		baseURL, err := url.Parse("https://cloudreve.org")
		asserts.NoError(err)
		downloadURL, err := handler.Source(ctx, "", *baseURL, 10, true, 0)
		asserts.Error(err)
		asserts.Empty(downloadURL)
	}
}

func TestHandler_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Handler{}
	ctx := context.Background()
	_, err := handler.Token(ctx, 10, "123")
	asserts.NoError(err)
}
