package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownload_Create(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		download := Download{GID: "1"}
		id, err := download.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.EqualValues(1, id)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		download := Download{GID: "1"}
		id, err := download.Create()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.EqualValues(0, id)
	}
}

func TestDownload_AfterFind(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		download := Download{Attrs: `{"gid":"123"}`}
		err := download.AfterFind()
		asserts.NoError(err)
		asserts.Equal("123", download.StatusInfo.Gid)
	}

	// 忽略空值
	{
		download := Download{Attrs: ``}
		err := download.AfterFind()
		asserts.NoError(err)
		asserts.Equal("", download.StatusInfo.Gid)
	}

	// 解析失败
	{
		download := Download{Attrs: `?`}
		err := download.BeforeSave()
		asserts.Error(err)
		asserts.Equal("", download.StatusInfo.Gid)
	}

}

func TestDownload_Save(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		download := Download{
			Model: gorm.Model{
				ID: 1,
			},
			Attrs: `{"gid":"123"}`,
		}
		err := download.Save()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal("123", download.StatusInfo.Gid)
	}

	// 失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		download := Download{
			Model: gorm.Model{
				ID: 1,
			},
		}
		err := download.Save()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestGetDownloadsByStatus(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WithArgs(0, 1).WillReturnRows(sqlmock.NewRows([]string{"gid"}).AddRow("0").AddRow("1"))
	res := GetDownloadsByStatus(0, 1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(res, 2)
}

func TestGetDownloadByGid(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WithArgs(2, "gid").WillReturnRows(sqlmock.NewRows([]string{"g_id"}).AddRow("1"))
	res, err := GetDownloadByGid("gid", 2)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
	asserts.Equal(res.GID, "1")
}

func TestDownload_GetOwner(t *testing.T) {
	asserts := assert.New(t)

	// 已经有User对象
	{
		download := &Download{User: &User{Nick: "nick"}}
		user := download.GetOwner()
		asserts.NotNil(user)
		asserts.Equal("nick", user.Nick)
	}

	// 无User对象
	{
		download := &Download{UserID: 3}
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"nick"}).AddRow("nick"))
		user := download.GetOwner()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NotNil(user)
		asserts.Equal("nick", user.Nick)
	}
}

func TestGetDownloadsByStatusAndUser(t *testing.T) {
	asserts := assert.New(t)

	// 列出全部
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, 1, 2).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2).AddRow(3))
		res := GetDownloadsByStatusAndUser(0, 1, 1, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Len(res, 2)
	}

	// 列出全部,分页
	{
		mock.ExpectQuery("SELECT(.+)DESC(.+)").WithArgs(1, 1, 2).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2).AddRow(3))
		res := GetDownloadsByStatusAndUser(2, 1, 1, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Len(res, 2)
	}
}

func TestDownload_Delete(t *testing.T) {
	asserts := assert.New(t)
	share := Download{}

	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := share.Delete()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

}
