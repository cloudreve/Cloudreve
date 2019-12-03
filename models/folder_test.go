package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFolderByPath(t *testing.T) {
	asserts := assert.New(t)

	folderRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(folderRows)
	folder, _ := GetFolderByPath("/测试/test", 1)
	asserts.Equal("test", folder.Name)
	asserts.NoError(mock.ExpectationsWereMet())

	folderRows = sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(folderRows)
	folder, err := GetFolderByPath("/测试/test", 1)
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFolder_Create(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Name: "new folder",
	}

	// 插入成功
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()
	fid, err := folder.Create()
	asserts.NoError(err)
	asserts.Equal(uint(5), fid)
	asserts.NoError(mock.ExpectationsWereMet())

	// 插入失败
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()
	fid, err = folder.Create()
	asserts.Error(err)
	asserts.Equal(uint(0), fid)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFolder_GetChildFolder(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Model: gorm.Model{
			ID: 1,
		},
	}

	// 找不到
	mock.ExpectQuery("SELECT(.+)parent_id(.+)").WithArgs(1).WillReturnError(errors.New("error"))
	files, err := folder.GetChildFolder()
	asserts.Error(err)
	asserts.Len(files, 0)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找到了
	mock.ExpectQuery("SELECT(.+)parent_id(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("1.txt", 1).AddRow("2.txt", 2))
	files, err = folder.GetChildFolder()
	asserts.NoError(err)
	asserts.Len(files, 2)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestGetRecursiveChildFolder(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	dirs := []string{"/目录1", "/目录2"}

	// 正常
	{
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1, util.BuildRegexp(dirs, "^", "/", "|"), 1, "/目录1", "/目录2").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "sub1").
					AddRow(2, "sub2").
					AddRow(3, "sub3"),
			)
		subs, err := GetRecursiveChildFolder(dirs, 1, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(subs, 3)
	}
	// 出错
	{
		mock.ExpectQuery("SELECT(.+)folders(.+)").
			WithArgs(1, util.BuildRegexp(dirs, "^", "/", "|"), 1, "/目录1", "/目录2").
			WillReturnError(errors.New("233"))
		subs, err := GetRecursiveChildFolder(dirs, 1, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Len(subs, 0)
	}
}

func TestGetRecursiveChildFolderSQLite(t *testing.T) {
	conf.DatabaseConfig.Type = "sqlite3"
	asserts := assert.New(t)

	// 测试目录结构
	//      1
	//     2  3
	//   4  5   6

	// 查询第一层
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, "/test").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "folder1"),
		)
	// 查询第二层
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).
				AddRow(2, "folder2").
				AddRow(3, "folder3"),
		)
	// 查询第三层
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 2, 3).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).
				AddRow(4, "folder4").
				AddRow(5, "folder5").
				AddRow(6, "folder6"),
		)
	// 查询第四层
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 4, 5, 6).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}),
		)

	folders, err := GetRecursiveChildFolder([]string{"/test"}, 1, true)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(folders, 6)
}

func TestDeleteFolderByIDs(t *testing.T) {
	asserts := assert.New(t)

	// 出错
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := DeleteFolderByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		err := DeleteFolderByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
}

func TestFolder_MoveOrCopyFileTo(t *testing.T) {
	asserts := assert.New(t)
	// 当前目录
	folder := Folder{
		OwnerID:          1,
		PositionAbsolute: "/test",
	}
	// 目标目录
	dstFolder := Folder{
		Model:            gorm.Model{ID: 10},
		PositionAbsolute: "/dst",
	}

	// 复制文件
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(
				"1.txt",
				"2.txt",
				1,
				"/test",
			).WillReturnRows(
			sqlmock.NewRows([]string{"id", "size"}).
				AddRow(1, 10).
				AddRow(2, 20),
		)
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		storage, err := folder.MoveOrCopyFileTo(
			[]string{"1.txt", "2.txt"},
			&dstFolder,
			true,
		)
		asserts.NoError(err)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint64(30), storage)
	}

	// 复制文件, 检索文件出错
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(
				"1.txt",
				"2.txt",
				1,
				"/test",
			).WillReturnError(errors.New("error"))

		storage, err := folder.MoveOrCopyFileTo(
			[]string{"1.txt", "2.txt"},
			&dstFolder,
			true,
		)
		asserts.Error(err)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint64(0), storage)
	}

	// 复制文件,第二个文件插入出错
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(
				"1.txt",
				"2.txt",
				1,
				"/test",
			).WillReturnRows(
			sqlmock.NewRows([]string{"id", "size"}).
				AddRow(1, 10).
				AddRow(2, 20),
		)
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		storage, err := folder.MoveOrCopyFileTo(
			[]string{"1.txt", "2.txt"},
			&dstFolder,
			true,
		)
		asserts.Error(err)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint64(10), storage)
	}

	// 移动文件 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst", 10, sqlmock.AnyArg(), "1.txt", "2.txt", 1, "/test").
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()
		storage, err := folder.MoveOrCopyFileTo(
			[]string{"1.txt", "2.txt"},
			&dstFolder,
			false,
		)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(uint64(0), storage)
	}

	// 移动文件 出错
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst", 10, sqlmock.AnyArg(), "1.txt", "2.txt", 1, "/test").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		storage, err := folder.MoveOrCopyFileTo(
			[]string{"1.txt", "2.txt"},
			&dstFolder,
			false,
		)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)
	}
}
