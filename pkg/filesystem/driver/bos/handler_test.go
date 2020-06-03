package bos

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestDriver_Get(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "accesskey",
			SecretKey:  "secretkey",
			BucketName: "file",
			Server:     "http://39.105.48.70:9002",
		},
	}
	path := "TestFile.txt_44c873c0fe0a4733a28e701c4e24cf6d"
	resp, err := handler.Get(context.Background(), path)
	asserts.Error(err)
	asserts.Nil(resp)
}

func TestDriver_Put(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "accesskey",
			SecretKey:  "secretkey",
			BucketName: "file",
			Server:     "http://39.105.48.70:9002",
		},
	}

	dst := "TestFile.txt"
	err := handler.Put(context.Background(), ioutil.NopCloser(strings.NewReader("666")), dst, 3)
	asserts.Error(err)
}

func TestDriver_Thumb(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey: "accesskey",
			SecretKey: "secretkey",
			Server:    "http://39.105.48.70:9002",
		},
	}

	resp, err := handler.Thumb(context.Background(), "./999.txt")
	asserts.Error(err)
	asserts.Nil(resp)
}

func TestDriver_Delete(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "accesskey",
			SecretKey:  "secretkey",
			BucketName: "file",
			Server:     "http://39.105.48.70:9002",
		},
	}

	resp, err := handler.Delete(context.Background(), []string{"TestFile.txt_3e46f40e0e774376979add820f142653"})
	asserts.Error(err)
	log.Println(resp)
}

func TestDriver_List(t *testing.T) {
	asserts := assert.New(t)
	handler := Driver{
		Policy: &model.Policy{
			AccessKey:  "accesskey",
			SecretKey:  "secretkey",
			BucketName: "439C4A6DC3F38B825D03F0357729A1E22B39FFCE",
			Server:     "http://123.56.171.188:9002",
		},
	}

	resp, err := handler.List(context.Background(), "1", false)
	asserts.Error(err)
	log.Println(resp)
}
