package fsctx

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

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
	{
		file := FileStream{
			File: ioutil.NopCloser(strings.NewReader("123")),
		}
		err := file.Close()
		asserts.NoError(err)
	}

	{
		file := FileStream{}
		err := file.Close()
		asserts.NoError(err)
	}
}

func TestFileStream_Seek(t *testing.T) {
	asserts := assert.New(t)
	f, _ := os.CreateTemp("", "*")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	{
		file := FileStream{
			File:   f,
			Seeker: f,
		}
		res, err := file.Seek(0, io.SeekStart)
		asserts.NoError(err)
		asserts.EqualValues(0, res)
	}

	{
		file := FileStream{}
		res, err := file.Seek(0, io.SeekStart)
		asserts.Error(err)
		asserts.EqualValues(0, res)
	}
}

func TestFileStream_Info(t *testing.T) {
	a := assert.New(t)
	file := FileStream{}
	a.NotNil(file.Info())

	file.SetSize(10)
	a.EqualValues(10, file.Info().Size)

	file.SetModel(&model.File{})
	a.NotNil(file.Info().Model)
}
