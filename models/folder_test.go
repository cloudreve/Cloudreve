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

func TestFolder_MoveOrCopyFolderTo_Copy(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	// 父目录
	parFolder := Folder{
		OwnerID:          1,
		PositionAbsolute: "/",
	}
	// 目标目录
	dstFolder := Folder{
		Model:            gorm.Model{ID: 10},
		PositionAbsolute: "/dst",
	}

	// 测试复制目录结构
	//       test(2)
	//    1(3)    2.txt
	//  3(4) 4.txt

	// 正常情况 成功
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(2, 1, "test", "/", "/test"),
			)

		// 更新顶级目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit()
		// 更新子目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(6, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(7, 1))
		mock.ExpectCommit()

		// 获取子目录下的所有子文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 3, 2, 4).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "folder_id", "dir", "size"}).
					AddRow(1, 2, "/test", 10).
					AddRow(2, 3, "/test/1", 20),
			)

		// 更新子文件记录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(uint64(30), storage)

	}

	// 处理子目录时死循环避免
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/1").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(2, 1, "test", "/", "/test"),
			)

		// 更新顶级目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit()
		// 更新子目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(6, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(6, 1))
		mock.ExpectCommit()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)

	}

	// 检索子目录出错
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnError(errors.New("error"))

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)

	}

	// 寻找原始目录出错
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnError(errors.New("error"))

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)

	}

	// 更新顶级目录出错
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(2, 1, "test", "/", "/test"),
			)

		// 更新顶级目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)

	}

	// 复制子目录，一个成功一个失败
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(2, 1, "test", "/", "/test"),
			)

		// 更新顶级目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit()
		// 更新子目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(6, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)

	}

	// 复制文件，一个成功一个失败
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 查找顶级待复制目录
		mock.ExpectQuery("SELECT(.+)").
			WithArgs("/test", 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(2, 1, "test", "/", "/test"),
			)

		// 更新顶级目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit()
		// 更新子目录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(6, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").
			WillReturnResult(sqlmock.NewResult(7, 1))
		mock.ExpectCommit()

		// 获取子目录下的所有子文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 3, 2, 4).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "folder_id", "dir", "size"}).
					AddRow(1, 2, "/test", 10).
					AddRow(2, 3, "/test/1", 20),
			)

		// 更新子文件记录
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, true)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(uint64(10), storage)

	}

}

func TestFolder_MoveOrCopyFolderTo_Move(t *testing.T) {
	conf.DatabaseConfig.Type = "mysql"
	asserts := assert.New(t)
	// 父目录
	parFolder := Folder{
		OwnerID:          1,
		PositionAbsolute: "/",
	}
	// 目标目录
	dstFolder := Folder{
		Model:            gorm.Model{ID: 10},
		PositionAbsolute: "/dst",
	}

	// 测试复制目录结构
	//       test(2)
	//    1(3)    2.txt
	//  3(4) 4.txt

	// 正常情况 成功
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 更改顶级要移动目录的父目录指向
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs(10, "/dst", "/dst/", sqlmock.AnyArg(), "/test", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		// 移动子目录
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst/test", "/dst/test/1", sqlmock.AnyArg(), 3).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst/test/1", "/dst/test/1/3", sqlmock.AnyArg(), 4).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		// 获取子文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 3, 2, 4).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "folder_id", "dir", "size"}).
					AddRow(1, 2, "/test", 10).
					AddRow(2, 3, "/test/1", 20),
			)
		// 移动子文件
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst/test", sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("/dst/test/1", sqlmock.AnyArg(), 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, false)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(uint64(0), storage)
	}

	// 无法移动顶层目录
	{
		// 查找所有递归子目录，包括自身
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, sqlmock.AnyArg(), 1, "/test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "parent_id", "name", "position", "position_absolute"}).
					AddRow(3, 2, "1", "/test", "/test/1").
					AddRow(2, 1, "test", "/", "/test").
					AddRow(4, 3, "3", "/test/1", "/test/1/3"),
			)
		// 更改顶级要移动目录的父目录指向
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs(10, "/dst", "/dst/", sqlmock.AnyArg(), "/test", 1).
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		storage, err := parFolder.MoveOrCopyFolderTo([]string{"/test"}, &dstFolder, false)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(uint64(0), storage)
	}
}
