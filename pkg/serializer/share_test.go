package serializer

import (
	"testing"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestBuildShareList(t *testing.T) {
	asserts := assert.New(t)
	timeNow := time.Now()

	shares := []model.Share{
		{
			Expires: &timeNow,
			File: model.File{
				Model: gorm.Model{ID: 1},
			},
		},
		{
			Folder: model.Folder{
				Model: gorm.Model{ID: 1},
			},
		},
	}

	res := BuildShareList(shares, 2)
	asserts.Equal(0, res.Code)
}

func TestBuildShareResponse(t *testing.T) {
	asserts := assert.New(t)

	// 未解锁
	{
		share := &model.Share{
			User:      model.User{Model: gorm.Model{ID: 1}},
			Downloads: 1,
		}
		res := BuildShareResponse(share, false)
		asserts.EqualValues(0, res.Downloads)
		asserts.True(res.Locked)
		asserts.NotNil(res.Creator)
	}

	// 已解锁，非目录
	{
		expires := time.Now().Add(time.Duration(10) * time.Second)
		share := &model.Share{
			User:      model.User{Model: gorm.Model{ID: 1}},
			Downloads: 1,
			Expires:   &expires,
			File: model.File{
				Model: gorm.Model{ID: 1},
			},
		}
		res := BuildShareResponse(share, true)
		asserts.EqualValues(1, res.Downloads)
		asserts.False(res.Locked)
		asserts.NotEmpty(res.Expire)
		asserts.NotNil(res.Creator)
	}

	// 已解锁，是目录
	{
		expires := time.Now().Add(time.Duration(10) * time.Second)
		share := &model.Share{
			User:      model.User{Model: gorm.Model{ID: 1}},
			Downloads: 1,
			Expires:   &expires,
			Folder: model.Folder{
				Model: gorm.Model{ID: 1},
			},
			IsDir: true,
		}
		res := BuildShareResponse(share, true)
		asserts.EqualValues(1, res.Downloads)
		asserts.False(res.Locked)
		asserts.NotEmpty(res.Expire)
		asserts.NotNil(res.Creator)
	}
}
