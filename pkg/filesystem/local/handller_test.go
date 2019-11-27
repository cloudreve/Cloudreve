package local

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
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
	asserts.Equal([]string{"test.file"}, list)
	asserts.NoError(err)

	file, err = os.Create("test.file")
	asserts.NoError(err)
	_ = file.Close()
	list, err = handler.Delete(ctx, []string{"test.file", "test.notexist"})
	asserts.Equal([]string{"test.file"}, list)
	asserts.Error(err)

	list, err = handler.Delete(ctx, []string{"test.notexist"})
	asserts.Equal([]string{}, list)
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
