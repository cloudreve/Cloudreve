package common

import (
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/stretchr/testify/assert"
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

	a.Equal(GetStatus(rpc.StatusInfo{Status: "complete"}), Complete)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "active",
		BitTorrent: rpc.BitTorrentInfo{Mode: ""}}), Downloading)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "active",
		BitTorrent:  rpc.BitTorrentInfo{Mode: "single"},
		TotalLength: "100", CompletedLength: "50"}), Downloading)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "active",
		BitTorrent:  rpc.BitTorrentInfo{Mode: "multi"},
		TotalLength: "100", CompletedLength: "100"}), Seeding)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "waiting"}), Ready)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "paused"}), Paused)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "error"}), Error)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "removed"}), Canceled)
	a.Equal(GetStatus(rpc.StatusInfo{Status: "unknown"}), Unknown)
}
