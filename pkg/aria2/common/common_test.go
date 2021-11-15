package common

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDummyAria2(t *testing.T) {
	a := assert.New(t)
	d := &DummyAria2{}

	a.NoError(d.Init())

	res, err := d.CreateTask(&model.Download{}, map[string]interface{}{})
	a.Empty(res)
	a.Error(err)

	_, err = d.Status(&model.Download{})
	a.Error(err)

	err = d.Cancel(&model.Download{})
	a.Error(err)

	err = d.Select(&model.Download{}, []int{})
	a.Error(err)

	configRes := d.GetConfig()
	a.NotNil(configRes)

	err = d.DeleteTempFile(&model.Download{})
	a.Error(err)
}

func TestGetStatus(t *testing.T) {
	a := assert.New(t)

	a.Equal(GetStatus("complete"), Complete)
	a.Equal(GetStatus("active"), Downloading)
	a.Equal(GetStatus("waiting"), Ready)
	a.Equal(GetStatus("paused"), Paused)
	a.Equal(GetStatus("error"), Error)
	a.Equal(GetStatus("removed"), Canceled)
	a.Equal(GetStatus("unknown"), Unknown)
}
