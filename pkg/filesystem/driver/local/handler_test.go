package local

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHandler_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{}

	defer func() {
		os.Remove(util.RelativePath("TestHandler_Put.txt"))
		os.Remove(util.RelativePath("inner/TestHandler_Put.txt"))
	}()

	testCases := []struct {
		file        fsctx.FileHeader
		errContains string
	}{
		{&fsctx.FileStream{
			SavePath: "TestHandler_Put.txt",
			File:     io.NopCloser(strings.NewReader("")),
		}, ""},
		{&fsctx.FileStream{
			SavePath: "TestHandler_Put.txt",
			File:     io.NopCloser(strings.NewReader("")),
		}, "file with the same name existed or unavailable"},
		{&fsctx.FileStream{
			SavePath: "inner/TestHandler_Put.txt",
			File:     io.NopCloser(strings.NewReader("")),
		}, ""},
		{&fsctx.FileStream{
			Mode:     fsctx.Append | fsctx.Overwrite,
			SavePath: "inner/TestHandler_Put.txt",
			File:     io.NopCloser(strings.NewReader("123")),
		}, ""},
		{&fsctx.FileStream{
			AppendStart: 10,
			Mode:        fsctx.Append | fsctx.Overwrite,
			SavePath:    "inner/TestHandler_Put.txt",
			File:        io.NopCloser(strings.NewReader("123")),
		}, "size of unfinished uploaded chunks is not as expected"},
		{&fsctx.FileStream{
			Mode:     fsctx.Append | fsctx.Overwrite,
			SavePath: "inner/TestHandler_Put.txt",
			File:     io.NopCloser(strings.NewReader("123")),
		}, ""},
	}

	for _, testCase := range testCases {
		err := handler.Put(context.Background(), testCase.file)
		if testCase.errContains != "" {
			asserts.Error(err)
			asserts.Contains(err.Error(), testCase.errContains)
		} else {
			asserts.NoError(err)
			asserts.True(util.Exists(util.RelativePath(testCase.file.Info().SavePath)))
		}
	}
}

func TestDriver_TruncateFailed(t *testing.T) {
	a := assert.New(t)
	h := Driver{}
	a.Error(h.Truncate(context.Background(), "TestDriver_TruncateFailed", 0))
}

func TestHandler_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{}
	ctx := context.Background()
	filePath := util.RelativePath("TestHandler_Delete.file")

	file, err := os.Create(filePath)
	asserts.NoError(err)
	_ = file.Close()
	list, err := handler.Delete(ctx, []string{"TestHandler_Delete.file"})
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	file, err = os.Create(filePath)
	_ = file.Close()
	file, _ = os.OpenFile(filePath, os.O_RDWR, os.FileMode(0))
	asserts.NoError(err)
	list, err = handler.Delete(ctx, []string{"TestHandler_Delete.file", "test.notexist"})
	file.Close()
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	list, err = handler.Delete(ctx, []string{"test.notexist"})
	asserts.Equal([]string{}, list)
	asserts.NoError(err)

	file, err = os.Create(filePath)
	asserts.NoError(err)
	list, err = handler.Delete(ctx, []string{"TestHandler_Delete.file"})
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
	file, err := os.Create(util.RelativePath("TestHandler_Thumb._thumb"))
	asserts.NoError(err)
	file.Close()

	f := &model.File{
		SourceName: "TestHandler_Thumb",
		MetadataSerialized: map[string]string{
			model.ThumbStatusMetadataKey: model.ThumbStatusExist,
		},
	}

	// 正常
	{
		thumb, err := handler.Thumb(ctx, f)
		asserts.NoError(err)
		asserts.NotNil(thumb.Content)
	}

	// file 不存在
	{
		f.SourceName = "not_exist"
		_, err := handler.Thumb(ctx, f)
		asserts.Error(err)
		asserts.ErrorIs(err, driver.ErrorThumbNotExist)
	}

	// thumb not exist
	{
		f.MetadataSerialized[model.ThumbStatusMetadataKey] = model.ThumbStatusNotExist
		_, err := handler.Thumb(ctx, f)
		asserts.Error(err)
		asserts.ErrorIs(err, driver.ErrorThumbNotExist)
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
		sourceURL, err := handler.Source(ctx, "", 0, false, 0)
		asserts.NoError(err)
		asserts.NotEmpty(sourceURL)
		asserts.Contains(sourceURL, "sign=")
	}

	// 下载
	{
		file := model.File{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test.jpg",
		}
		ctx := context.WithValue(ctx, fsctx.FileModelCtx, file)
		sourceURL, err := handler.Source(ctx, "", 0, true, 0)
		asserts.NoError(err)
		asserts.NotEmpty(sourceURL)
		asserts.Contains(sourceURL, "sign=")
		asserts.Contains(sourceURL, "download")
	}

	// 无法获取上下文
	{
		sourceURL, err := handler.Source(ctx, "", 0, false, 0)
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
		sourceURL, err := handler.Source(ctx, "", 0, false, 0)
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
		sourceURL, err := handler.Source(ctx, "", 0, false, 0)
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
		downloadURL, err := handler.Source(ctx, "", 10, true, 0)
		asserts.NoError(err)
		asserts.Contains(downloadURL, "sign=")
	}

	// 无法获取上下文
	{
		downloadURL, err := handler.Source(ctx, "", 10, true, 0)
		asserts.Error(err)
		asserts.Empty(downloadURL)
	}
}

func TestHandler_Token(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{},
	}
	ctx := context.Background()
	upSession := &serializer.UploadSession{SavePath: "TestHandler_Token"}
	_, err := handler.Token(ctx, 10, upSession, &fsctx.FileStream{})
	asserts.NoError(err)

	file, _ := os.Create("TestHandler_Token")
	defer func() {
		file.Close()
		os.Remove("TestHandler_Token")
	}()

	_, err = handler.Token(ctx, 10, upSession, &fsctx.FileStream{})
	asserts.Error(err)
	asserts.Contains(err.Error(), "already exist")
}

func TestDriver_CancelToken(t *testing.T) {
	a := assert.New(t)
	handler := Driver{}
	a.NoError(handler.CancelToken(context.Background(), &serializer.UploadSession{}))
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
