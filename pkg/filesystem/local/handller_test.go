package local

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
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
			dst:  "notexist:/S.TXT",
			err:  true,
		},
	}

	for _, testCase := range testCases {
		err := handler.Put(ctx, testCase.file, testCase.dst)
		if testCase.err {
			asserts.Error(err)
		} else {
			asserts.NoError(err)
			asserts.True(util.Exists(testCase.dst))
		}
	}
}
