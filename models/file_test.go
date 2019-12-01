package model

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFileByPathAndName(t *testing.T) {
	asserts := assert.New(t)

	fileRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "1.cia")
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(fileRows)
	file, _ := GetFileByPathAndName("/", "1.cia", 1)
	asserts.Equal("1.cia", file.Name)
	asserts.NoError(mock.ExpectationsWereMet())
}

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
	folder := &Folder{
		Model: gorm.Model{
			ID: 1,
		},
	}

	// 找不到
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnError(errors.New("error"))
	files, err := folder.GetChildFile()
	asserts.Error(err)
	asserts.Len(files, 0)
	asserts.NoError(mock.ExpectationsWereMet())

	// 找到了
	mock.ExpectQuery("SELECT(.+)folder_id(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("1.txt", 1).AddRow("2.txt", 2))
	files, err = folder.GetChildFile()
	asserts.NoError(err)
	asserts.Len(files, 2)
	asserts.NoError(mock.ExpectationsWereMet())

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

func TestGetFileByPaths(t *testing.T) {
	asserts := assert.New(t)
	paths := []string{"/我的目录/文件.txt", "/根目录文件.txt"}

	// 正常情况
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("/我的目录", "文件.txt", 1, "/", "根目录文件.txt", 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "文件.txt").
					AddRow(2, "根目录文件.txt"),
			)
		files, err := GetFileByPaths(paths, 1)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal([]File{
			File{
				Model: gorm.Model{ID: 1},
				Name:  "文件.txt",
			},
			File{
				Model: gorm.Model{ID: 2},
				Name:  "根目录文件.txt",
			},
		}, files)
	}

	// 出错
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("/我的目录", "文件.txt", 1, "/", "根目录文件.txt", 1).
			WillReturnError(errors.New("error"))
		files, err := GetFileByPaths(paths, 1)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Len(files, 0)
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
		mock.ExpectExec("UPDATE(.+)delete(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := DeleteFileByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)delete(.+)").
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		err := DeleteFileByIDs([]uint{1, 2, 3})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
}
