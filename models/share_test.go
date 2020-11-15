package model

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestShare_Create(t *testing.T) {
	asserts := assert.New(t)
	share := Share{UserID: 1}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()
		id, err := share.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.EqualValues(2, id)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		id, err := share.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualValues(0, id)
	}
}

func TestGetShareByHashID(t *testing.T) {
	asserts := assert.New(t)
	conf.SystemConfig.HashIDSalt = ""

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		res := GetShareByHashID("x9T4")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotNil(res)
	}

	// 查询失败
	{
		mock.ExpectQuery("SELECT(.+)").
			WillReturnError(errors.New("error"))
		res := GetShareByHashID("x9T4")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Nil(res)
	}

	// ID解码失败
	{
		res := GetShareByHashID("empty")
		asserts.Nil(res)
	}

}

func TestShare_IsAvailable(t *testing.T) {
	asserts := assert.New(t)

	// 下载剩余次数为0
	{
		share := Share{}
		asserts.False(share.IsAvailable())
	}

	// 时效过期
	{
		expires := time.Unix(10, 10)
		share := Share{
			RemainDownloads: -1,
			Expires:         &expires,
		}
		asserts.False(share.IsAvailable())
	}

	// 源对象为目录，但不存在
	{
		share := Share{
			RemainDownloads: -1,
			SourceID:        2,
			IsDir:           true,
		}
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		asserts.False(share.IsAvailable())
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 源对象为目录，存在
	{
		share := Share{
			RemainDownloads: -1,
			SourceID:        2,
			IsDir:           false,
		}
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(13))
		asserts.True(share.IsAvailable())
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 用户被封禁
	{
		share := Share{
			RemainDownloads: -1,
			SourceID:        2,
			IsDir:           true,
			User:            User{Status: Baned},
		}
		asserts.False(share.IsAvailable())
	}
}

func TestShare_GetCreator(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	share := Share{UserID: 1}
	res := share.Creator()
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.EqualValues(1, res.ID)
}

func TestShare_Source(t *testing.T) {
	asserts := assert.New(t)

	// 目录
	{
		share := Share{IsDir: true, SourceID: 3}
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
		asserts.EqualValues(3, share.Source().(*Folder).ID)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 文件
	{
		share := Share{IsDir: false, SourceID: 3}
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
		asserts.EqualValues(3, share.Source().(*File).ID)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestShare_CanBeDownloadBy(t *testing.T) {
	asserts := assert.New(t)
	share := Share{}

	// 未登录，无权
	{
		user := &User{
			Group: Group{
				OptionsSerialized: GroupOption{
					ShareDownload: false,
				},
			},
		}
		asserts.Error(share.CanBeDownloadBy(user))
	}

	// 已登录，无权
	{
		user := &User{
			Model: gorm.Model{ID: 1},
			Group: Group{
				OptionsSerialized: GroupOption{
					ShareDownload: false,
				},
			},
		}
		asserts.Error(share.CanBeDownloadBy(user))
	}

	// 未登录，需要积分
	{
		user := &User{
			Group: Group{
				OptionsSerialized: GroupOption{
					ShareDownload: true,
				},
			},
		}
		asserts.Error(share.CanBeDownloadBy(user))
	}

	// 成功
	{
		user := &User{
			Model: gorm.Model{ID: 1},
			Group: Group{
				OptionsSerialized: GroupOption{
					ShareDownload: true,
				},
			},
		}
		asserts.NoError(share.CanBeDownloadBy(user))
	}
}

func TestShare_WasDownloadedBy(t *testing.T) {
	asserts := assert.New(t)
	share := Share{
		Model: gorm.Model{ID: 1},
	}

	// 已登录，已下载
	{
		user := User{
			Model: gorm.Model{
				ID: 1,
			},
		}
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		cache.Set("share_1_1", true, 0)
		asserts.True(share.WasDownloadedBy(&user, c))
	}
}

func TestShare_DownloadBy(t *testing.T) {
	asserts := assert.New(t)
	share := Share{
		Model: gorm.Model{ID: 1},
	}
	user := User{
		Model: gorm.Model{
			ID: 1,
		},
	}
	cache.Deletes([]string{"1_1"}, "share_")
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := share.DownloadBy(&user, c)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
	_, ok := cache.Get("share_1_1")
	asserts.True(ok)
}

func TestShare_Viewed(t *testing.T) {
	asserts := assert.New(t)
	share := Share{}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	share.Viewed()
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.EqualValues(1, share.Views)
}

func TestShare_UpdateAndDelete(t *testing.T) {
	asserts := assert.New(t)
	share := Share{}

	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := share.Update(map[string]interface{}{"id": 1})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := share.Delete()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := DeleteShareBySourceIDs([]uint{1}, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

}

func TestListShares(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2).AddRow(2))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	res, total := ListShares(1, 1, 10, "desc", true)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(res, 2)
	asserts.Equal(2, total)
}

func TestSearchShares(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").
		WithArgs("", sqlmock.AnyArg(), "%1%2%").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	res, total := SearchShares(1, 10, "id", "1 2")
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(res, 1)
	asserts.Equal(1, total)
}
