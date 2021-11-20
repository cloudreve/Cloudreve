package serializer

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewError(t *testing.T) {
	a := assert.New(t)
	err := NewError(400, "Bad Request", errors.New("error"))
	a.Error(err)
	a.EqualValues(400, err.Code)

	err.WithError(errors.New("error2"))
	a.Equal("error2", err.RawError.Error())
	a.Equal("Bad Request", err.Error())

	resp := &Response{
		Code:  400,
		Msg:   "Bad Request",
		Error: "error",
	}
	err = NewErrorFromResponse(resp)
	a.Error(err)
}

func TestDBErr(t *testing.T) {
	a := assert.New(t)
	resp := DBErr("", nil)
	a.NotEmpty(resp.Msg)

	resp = ParamErr("", nil)
	a.NotEmpty(resp.Msg)
}

func TestErr(t *testing.T) {
	a := assert.New(t)
	err := NewError(400, "Bad Request", errors.New("error"))
	resp := Err(400, "", err)
	a.Equal("Bad Request", resp.Msg)
}
