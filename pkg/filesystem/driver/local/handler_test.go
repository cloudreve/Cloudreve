package local

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
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
	handler := Driver{}
	ctx := context.WithValue(context.Background(), fsctx.DisableOverwrite, true)
	os.Remove(util.RelativePath("test/test/txt"))

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
			dst:  "test/test/txt",
			err:  true,
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
			asserts.True(util.Exists(util.RelativePath(testCase.dst)))
		}
	}
}

func TestHandler_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{}
	ctx := context.Background()
	filePath := util.RelativePath("test.file")

	file, err := os.Create(filePath)
	asserts.NoError(err)
	_ = file.Close()
	list, err := handler.Delete(ctx, []string{"test.file"})
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	file, err = os.Create(filePath)
	_ = file.Close()
	file, _ = os.OpenFile(filePath, os.O_RDWR, os.FileMode(0))
	asserts.NoError(err)
	list, err = handler.Delete(ctx, []string{"test.file", "test.notexist"})
	file.Close()
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	list, err = handler.Delete(ctx, []string{"test.notexist"})
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	file, err = os.Create(filePath)
	asserts.NoError(err)
	list, err = handler.Delete(ctx, []string{"test.file"})
	_ = file.Close()
	asserts.Equal([]string{}, list)
	asserts.NoError(err)
}

func TestHandler_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 成功
	file, err := os.Create(util.RelativePath("TestHandler_Get.txt"))
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
	handler := Driver{}
	ctx := context.Background()
	file, err := os.Create(util.RelativePath("TestHandler_Thumb" + conf.ThumbConfig.FileSuffix))
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
	handler := Driver{
		Policy: &model.Policy{},
	}
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

	// 设定了CDN
	{
		handler.Policy.BaseURL = "https://cqu.edu.cn"
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
		asserts.Contains(sourceURL, "https://cqu.edu.cn")
	}

	// 设定了CDN，解析失败
	{
		handler.Policy.BaseURL = string([]byte{0x7f})
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
		asserts.Error(err)
		asserts.Empty(sourceURL)
	}
}

func TestHandler_GetDownloadURL(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{Policy: &model.Policy{}}
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
	handler := Driver{}
	ctx := context.Background()
	_, err := handler.Token(ctx, 10, "123")
	asserts.NoError(err)
}

func TestDriver_List(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{}
	ctx := context.Background()

	// 创建测试目录结构
	for _, path := range []string{
		"test/TestDriver_List/parent.txt",
		"test/TestDriver_List/parent_folder2/sub2.txt",
		"test/TestDriver_List/parent_folder1/sub_folder/sub1.txt",
		"test/TestDriver_List/parent_folder1/sub_folder/sub2.txt",
	} {
		f, _ := util.CreatNestedFile(util.RelativePath(path))
		f.Close()
	}

	// 非递归列出
	{
		res, err := handler.List(ctx, "test/TestDriver_List", false)
		asserts.NoError(err)
		asserts.Len(res, 3)
	}

	// 递归列出
	{
		res, err := handler.List(ctx, "test/TestDriver_List", true)
		asserts.NoError(err)
		asserts.Len(res, 7)
	}
}
