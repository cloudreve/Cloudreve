package local

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func TestFileStream_GetFileName(t *testing.T) {
	asserts := assert.New(t)
	file := FileStream{Name: "123"}
	asserts.Equal("123", file.GetFileName())
}

func TestFileStream_GetMIMEType(t *testing.T) {
	asserts := assert.New(t)
	file := FileStream{MIMEType: "123"}
	asserts.Equal("123", file.GetMIMEType())
}

func TestFileStream_GetSize(t *testing.T) {
	asserts := assert.New(t)
	file := FileStream{Size: 123}
	asserts.Equal(uint64(123), file.GetSize())
}

func TestFileStream_Read(t *testing.T) {
	asserts := assert.New(t)
	file := FileStream{
		File: ioutil.NopCloser(strings.NewReader("123")),
	}
	var p = make([]byte, 3)
	{
		n, err := file.Read(p)
		asserts.Equal(3, n)
		asserts.NoError(err)
	}
}

func TestFileStream_Close(t *testing.T) {
	asserts := assert.New(t)
	file := FileStream{
		File: ioutil.NopCloser(strings.NewReader("123")),
	}
	err := file.Close()
	asserts.NoError(err)
}
