package serializer

import (
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestBuildFinishedListResponse(t *testing.T) {
	asserts := assert.New(t)
	tasks := []model.Download{
		{
			StatusInfo: rpc.StatusInfo{
				Files: []rpc.FileInfo{
					{
						Path: "/file/name.txt",
					},
				},
			},
			Task: &model.Task{
				Model: gorm.Model{},
				Error: "error",
			},
		},
		{
			StatusInfo: rpc.StatusInfo{
				Files: []rpc.FileInfo{
					{
						Path: "/file/name1.txt",
					},
					{
						Path: "/file/name2.txt",
					},
				},
			},
		},
	}
	tasks[1].StatusInfo.BitTorrent.Info.Name = "name.txt"
	res := BuildFinishedListResponse(tasks).Data.([]FinishedListResponse)
	asserts.Len(res, 2)
	asserts.Equal("name.txt", res[1].Name)
	asserts.Equal("name.txt", res[0].Name)
	asserts.Equal("name.txt", res[0].Files[0].Path)
	asserts.Equal("name1.txt", res[1].Files[0].Path)
	asserts.Equal("name2.txt", res[1].Files[1].Path)
	asserts.EqualValues(0, res[0].TaskStatus)
	asserts.Equal("error", res[0].TaskError)
}

func TestBuildDownloadingResponse(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("setting_aria2_interval", "10", 0)
	tasks := []model.Download{
		{
			StatusInfo: rpc.StatusInfo{
				Files: []rpc.FileInfo{
					{
						Path: "/file/name.txt",
					},
				},
			},
			Task: &model.Task{
				Model: gorm.Model{},
				Error: "error",
			},
		},
		{
			StatusInfo: rpc.StatusInfo{
				Files: []rpc.FileInfo{
					{
						Path: "/file/name1.txt",
					},
					{
						Path: "/file/name2.txt",
					},
				},
			},
		},
	}
	tasks[1].StatusInfo.BitTorrent.Info.Name = "name.txt"

	res := BuildDownloadingResponse(tasks).Data.([]DownloadListResponse)
	asserts.Len(res, 2)
	asserts.Equal("name1.txt", res[1].Name)
	asserts.Equal("name.txt", res[0].Name)
	asserts.Equal("name.txt", res[0].Info.Files[0].Path)
	asserts.Equal("name1.txt", res[1].Info.Files[0].Path)
	asserts.Equal("name2.txt", res[1].Info.Files[1].Path)
}
