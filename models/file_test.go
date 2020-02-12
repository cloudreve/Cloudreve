package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFile_Create(t *testing.T) {
	asserts := assert.New(t)
	file := File{
		Name: "123",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()
	fileID, err := file.Create()
	asserts.NoError(err)
	asserts.Equal(uint(5), fileID)
	asserts.Equal(uint(5), file.ID)
	asserts.NoError(mock.ExpectationsWereMet())

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()
	fileID, err = file.Create()
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestFolder_GetChildFile(t *testing.T) {
	asserts := assert.New(t)
	folder := Folder{Model: gorm.Model{ID: 1}, Name: "/"}
	// 存在
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, "1.txt").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1.txt"))
		file, err := folder.GetChildFile("1.txt")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal("1.txt", file.Name)
		asserts.Equal("/", file.Position)
	}

	// 不存在
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, "1.txt").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		_, err := folder.GetChildFile("1.txt")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestFolder_GetChildFiles(t *testing.T) {
	asserts := assert.New(t)
	folder := &Folder{
		Model: gorm.Model{
			ID: 1,
		},
		Position: "/123",
		Name:     "456",
	}

	// 找不到
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnError(errors.New("error"))
	files, err := folder.GetChildFiles()
	asserts.Error(err)
	asserts.Len(files, 0)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找到了
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("1.txt", 1).AddRow("2.txt", 2))
	files, err = folder.GetChildFiles()
	asserts.NoError(err)
	asserts.Len(files, 2)
	asserts.Equal("/123/456", files[0].Position)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestGetFilesByIDs(t *testing.T) {
	asserts := assert.New(t)

	// 出错
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3, 1).
			WillReturnError(errors.New("error"))
		folders, err := GetFilesByIDs([]uint{1, 2, 3}, 1)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Len(folders, 0)
	}

	// 部分找到
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1"))
		folders, err := GetFilesByIDs([]uint{1, 2, 3}, 1)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(folders, 1)
	}

	// 忽略UID查找
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2, 3).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1"))
		folders, err := GetFilesByIDs([]uint{1, 2, 3}, 0)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(folders, 1)
	}
}

func TestGetChildFilesOfFolders(t *testing.T) {
	asserts := assert.New(t)
	testFolder := []Folder{
		Folder{
			Model: gorm.Model{ID: 3},
		},
		Folder{
			Model: gorm.Model{ID: 4},
		}, Folder{
			Model: gorm.Model{ID: 5},
		},
	}

	// 出错
	{
		mock.ExpectQuery("SELECT(.+)folder_id").WithArgs(3, 4, 5).WillReturnError(errors.New("not found"))
		files, err := GetChildFilesOfFolders(&testFolder)
		asserts.Error(err)
		asserts.Len(files, 0)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 找到2个
	{
		mock.ExpectQuery("SELECT(.+)folder_id").
			WithArgs(3, 4, 5).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(3, "3").
				AddRow(4, "4"),
			)
		files, err := GetChildFilesOfFolders(&testFolder)
		asserts.NoError(err)
		asserts.Len(files, 2)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 全部找到
	{
		mock.ExpectQuery("SELECT(.+)folder_id").
			WithArgs(3, 4, 5).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(3, "3").
				AddRow(4, "4").
				AddRow(5, "5"),
			)
		files, err := GetChildFilesOfFolders(&testFolder)
		asserts.NoError(err)
		asserts.Len(files, 3)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestFile_GetPolicy(t *testing.T) {
	asserts := assert.New(t)

	// 空策略
	{
		file := File{
			PolicyID: 23,
		}
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name"}).
					AddRow(23, "name"),
			)
		file.GetPolicy()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint(23), file.Policy.ID)
	}

	// 非空策略
	{
		file := File{
			PolicyID: 23,
			Policy:   Policy{Model: gorm.Model{ID: 24}},
		}
		file.GetPolicy()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint(24), file.Policy.ID)
	}
}

func TestRemoveFilesWithSoftLinks(t *testing.T) {
	asserts := assert.New(t)
	files := []File{
		File{
			Model:      gorm.Model{ID: 1},
			SourceName: "1.txt",
			PolicyID:   23,
		},
		File{
			Model:      gorm.Model{ID: 2},
			SourceName: "2.txt",
			PolicyID:   24,
		},
	}

	// 全都没有
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1, "2.txt", 24, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(files, file)
	}
	// 查询出错
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1, "2.txt", 24, 2).
			WillReturnError(errors.New("error"))
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Nil(file)
	}
	// 第二个是软链
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1, "2.txt", 24, 2).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 24, "2.txt"),
			)
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(files[:1], file)
	}
	// 第一个是软链
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1, "2.txt", 24, 2).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 23, "1.txt"),
			)
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(files[1:], file)
	}
	// 全部是软链
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1, "2.txt", 24, 2).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 24, "2.txt").
					AddRow(4, 23, "1.txt"),
			)
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(file, 0)
	}
}

func TestDeleteFileByIDs(t *testing.T) {
	asserts := assert.New(t)

	// 出错
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := DeleteFileByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		err := DeleteFileByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
}

func TestGetFilesByParentIDs(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 4, 5, 6).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).
				AddRow(4, "4.txt").
				AddRow(5, "5.txt").
				AddRow(6, "6.txt"),
		)
	files, err := GetFilesByParentIDs([]uint{4, 5, 6}, 1)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Len(files, 3)
}

func TestFile_Updates(t *testing.T) {
	asserts := assert.New(t)
	file := File{Model: gorm.Model{ID: 1}}
	// rename
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs("newName", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := file.Rename("newName")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// UpdatePicInfo
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs(10, sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := file.UpdateSize(10)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// UpdatePicInfo
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs("1,1", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := file.UpdatePicInfo("1,1")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
}

func TestFile_FileInfoInterface(t *testing.T) {
	asserts := assert.New(t)
	file := File{
		Model: gorm.Model{
			UpdatedAt: time.Date(2019, 12, 21, 12, 40, 0, 0, time.UTC),
		},
		Name:       "test_name",
		SourceName: "",
		UserID:     0,
		Size:       10,
		PicInfo:    "",
		FolderID:   0,
		PolicyID:   0,
		Policy:     Policy{},
		Position:   "/test",
	}

	name := file.GetName()
	asserts.Equal("test_name", name)

	size := file.GetSize()
	asserts.Equal(uint64(10), size)

	asserts.Equal(time.Date(2019, 12, 21, 12, 40, 0, 0, time.UTC), file.ModTime())
	asserts.False(file.IsDir())
	asserts.Equal("/test", file.GetPosition())
}

func TestGetFilesByKeywords(t *testing.T) {
	asserts := assert.New(t)

	// 未指定用户
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs("k1", "k2").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		res, err := GetFilesByKeywords(0, "k1", "k2")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(res, 1)
	}

	// 指定用户
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, "k1", "k2").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		res, err := GetFilesByKeywords(1, "k1", "k2")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(res, 1)
	}
}
