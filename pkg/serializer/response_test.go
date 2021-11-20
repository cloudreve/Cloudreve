package serializer

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewResponseWithGobData(t *testing.T) {
	a := assert.New(t)
	type args struct {
		data interface{}
	}

	res := NewResponseWithGobData(args{})
	a.Equal(CodeInternalSetting, res.Code)

	res = NewResponseWithGobData("TestNewResponseWithGobData")
	a.Equal(0, res.Code)
	a.NotEmpty(res.Data)
}

func TestResponse_GobDecode(t *testing.T) {
	a := assert.New(t)
	res := NewResponseWithGobData("TestResponse_GobDecode")
	jsonContent, err := json.Marshal(res)
	a.NoError(err)
	resDecoded := &Response{}
	a.NoError(json.Unmarshal(jsonContent, resDecoded))
	var target string
	resDecoded.GobDecode(&target)
	a.Equal("TestResponse_GobDecode", target)
}
