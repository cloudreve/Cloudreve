package model

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestFile_Create(t *testing.T) {
	asserts := assert.New(t)
	file := File{
		Name: "123",
	}

	// 无法插入文件记录
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := file.Create()
		asserts.Error(err)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 无法更新用户容量
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := file.Create()
		asserts.Error(err)
		asserts.NoError(mock.ExpectationsWereMet())
	}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := file.Create()
		asserts.NoError(err)
		asserts.Equal(uint(5), file.ID)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestFile_AfterFind(t *testing.T) {
	a := assert.New(t)

	// metadata not empty
	{
		file := File{
			Name:     "123",
			Metadata: "{\"name\":\"123\"}",
		}

		a.NoError(file.AfterFind())
		a.Equal("123", file.MetadataSerialized["name"])
	}

	// metadata empty
	{
		file := File{
			Name:     "123",
			Metadata: "",
		}
		a.Nil(file.MetadataSerialized)
		a.NoError(file.AfterFind())
		a.NotNil(file.MetadataSerialized)
	}
}

func TestFile_BeforeSave(t *testing.T) {
	a := assert.New(t)

	// metadata not empty
	{
		file := File{
			Name: "123",
			MetadataSerialized: map[string]string{
				"name": "123",
			},
		}

		a.NoError(file.BeforeSave())
		a.Equal("{\"name\":\"123\"}", file.Metadata)
	}

	// metadata empty
	{
		file := File{
			Name: "123",
		}
		a.NoError(file.BeforeSave())
		a.Equal("", file.Metadata)
	}
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

func TestGetUploadPlaceholderFiles(t *testing.T) {
	a := assert.New(t)

	mock.ExpectQuery("SELECT(.+)upload_session_id(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1"))
	files := GetUploadPlaceholderFiles(1)
	a.NoError(mock.ExpectationsWereMet())
	a.Len(files, 1)
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

func TestRemoveFilesWithSoftLinks_EmptyArg(t *testing.T) {
	asserts := assert.New(t)
	// 传入空
	{
		mock.ExpectQuery("SELECT(.+)files(.+)")
		file, err := RemoveFilesWithSoftLinks([]File{})
		asserts.Error(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(len(file), 0)
		DB.Find(&File{})
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

	// 传入空文件列表
	{
		file, err := RemoveFilesWithSoftLinks([]File{})
		asserts.NoError(err)
		asserts.Empty(file)
	}

	// 全都没有
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("2.txt", 24, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(files, file)
	}

	// 第二个是软链
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("2.txt", 24, 2).
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
			WithArgs("1.txt", 23, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 23, "1.txt"),
			)
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("2.txt", 24, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "policy_id", "source_name"}))
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(files[1:], file)
	}
	// 全部是软链
	{
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("1.txt", 23, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 23, "1.txt"),
			)
		mock.ExpectQuery("SELECT(.+)files(.+)").
			WithArgs("2.txt", 24, 2).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(3, 24, "2.txt"),
			)
		file, err := RemoveFilesWithSoftLinks(files)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(file, 0)
	}
}

func TestDeleteFiles(t *testing.T) {
	a := assert.New(t)

	// uid 不一致
	{
		err := DeleteFiles([]*File{{UserID: 2}}, 1)
		a.Contains("user id not consistent", err.Error())
	}

	// 删除失败
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := DeleteFiles([]*File{{UserID: 1}}, 1)
		a.NoError(mock.ExpectationsWereMet())
		a.Error(err)
	}

	// 无法变更用户容量
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		err := DeleteFiles([]*File{{UserID: 1}}, 1)
		a.NoError(mock.ExpectationsWereMet())
		a.Error(err)
	}

	// 文件脏读
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 0))
		mock.ExpectRollback()
		err := DeleteFiles([]*File{{Size: 1, UserID: 1}, {Size: 2, UserID: 1}}, 1)
		a.NoError(mock.ExpectationsWereMet())
		a.Error(err)
		a.Contains("file size is dirty", err.Error())
	}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)").WithArgs(uint64(3), sqlmock.AnyArg(), uint(1)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := DeleteFiles([]*File{{Size: 1, UserID: 1}, {Size: 2, UserID: 1}}, 1)
		a.NoError(mock.ExpectationsWereMet())
		a.NoError(err)
	}

	// 成功,  关联用户不存在
	{
		mock.ExpectBegin()
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectExec("DELETE(.+)").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()
		err := DeleteFiles([]*File{{Size: 1, UserID: 1}, {Size: 2, UserID: 1}}, 0)
		a.NoError(mock.ExpectationsWereMet())
		a.NoError(err)
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

func TestGetFilesByUploadSession(t *testing.T) {
	a := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, "sessionID").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).AddRow(4, "4.txt"))
	files, err := GetFilesByUploadSession("sessionID", 1)
	a.NoError(err)
	a.NoError(mock.ExpectationsWereMet())
	a.Equal("4.txt", files.Name)
}

func TestFile_Updates(t *testing.T) {
	asserts := assert.New(t)
	file := File{Model: gorm.Model{ID: 1}}

	// rename
	{
		// not reset thumb
		{
			file := File{Model: gorm.Model{ID: 1}}
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE(.+)files(.+)SET(.+)").WithArgs("", "newName", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			err := file.Rename("newName")
			asserts.NoError(mock.ExpectationsWereMet())
			asserts.NoError(err)
		}

		// thumb not available, rename base name only
		{
			file := File{Model: gorm.Model{ID: 1}, Name: "1.txt", MetadataSerialized: map[string]string{
				ThumbStatusMetadataKey: ThumbStatusNotAvailable,
			},
				Metadata: "{}"}
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE(.+)files(.+)SET(.+)").WithArgs("{}", "newName.txt", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			err := file.Rename("newName.txt")
			asserts.NoError(mock.ExpectationsWereMet())
			asserts.NoError(err)
			asserts.Equal(ThumbStatusNotAvailable, file.MetadataSerialized[ThumbStatusMetadataKey])
		}

		// thumb not available, rename base name only
		{
			file := File{Model: gorm.Model{ID: 1}, Name: "1.txt", MetadataSerialized: map[string]string{
				ThumbStatusMetadataKey: ThumbStatusNotAvailable,
			}}
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE(.+)files(.+)SET(.+)").WithArgs("{}", "newName.jpg", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			err := file.Rename("newName.jpg")
			asserts.NoError(mock.ExpectationsWereMet())
			asserts.NoError(err)
			asserts.Empty(file.MetadataSerialized[ThumbStatusMetadataKey])
		}
	}

	// UpdatePicInfo
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs("1,1", 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := file.UpdatePicInfo("1,1")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// UpdateSourceName
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs("", "newName", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := file.UpdateSourceName("newName")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
}

func TestFile_UpdateSize(t *testing.T) {
	a := assert.New(t)

	// 增加成功
	{
		file := File{Size: 10}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 11, sqlmock.AnyArg(), 10).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)+(.+)").WithArgs(uint64(1), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		a.NoError(file.UpdateSize(11))
		a.NoError(mock.ExpectationsWereMet())
	}

	// 减少成功
	{
		file := File{Size: 10}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 8, sqlmock.AnyArg(), 10).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)-(.+)").WithArgs(uint64(2), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		a.NoError(file.UpdateSize(8))
		a.NoError(mock.ExpectationsWereMet())
	}

	// 文件更新失败
	{
		file := File{Size: 10}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 8, sqlmock.AnyArg(), 10).WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		a.Error(file.UpdateSize(8))
		a.NoError(mock.ExpectationsWereMet())
	}

	// 用户容量更新失败
	{
		file := File{Size: 10}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 8, sqlmock.AnyArg(), 10).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)storage(.+)-(.+)").WithArgs(uint64(2), sqlmock.AnyArg()).WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		a.Error(file.UpdateSize(8))
		a.NoError(mock.ExpectationsWereMet())
	}
}

func TestFile_PopChunkToFile(t *testing.T) {
	a := assert.New(t)
	timeNow := time.Now()
	file := File{}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(file.PopChunkToFile(&timeNow, "1,1"))
}

func TestFile_CanCopy(t *testing.T) {
	a := assert.New(t)
	file := File{}
	a.True(file.CanCopy())
	file.UploadSessionID = &file.Name
	a.False(file.CanCopy())
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
		res, err := GetFilesByKeywords(0, nil, "k1", "k2")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(res, 1)
	}

	// 指定用户
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, "k1", "k2").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		res, err := GetFilesByKeywords(1, nil, "k1", "k2")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(res, 1)
	}

	// 指定父目录
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, 12, "k1", "k2").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		res, err := GetFilesByKeywords(1, []uint{12}, "k1", "k2")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Len(res, 1)
	}
}

func TestFile_CreateOrGetSourceLink(t *testing.T) {
	a := assert.New(t)
	file := &File{}
	file.ID = 1

	// 已存在，返回老的 SourceLink
	{
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		res, err := file.CreateOrGetSourceLink()
		a.NoError(err)
		a.EqualValues(2, res.ID)
		a.NoError(mock.ExpectationsWereMet())
	}

	// 不存在，插入失败
	{
		expectedErr := errors.New("error")
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)source_links(.+)").WillReturnError(expectedErr)
		mock.ExpectRollback()
		res, err := file.CreateOrGetSourceLink()
		a.Nil(res)
		a.ErrorIs(err, expectedErr)
		a.NoError(mock.ExpectationsWereMet())
	}

	// 成功
	{
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectBegin()
		mock.ExpectExec("INSERT(.+)source_links(.+)").WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()
		res, err := file.CreateOrGetSourceLink()
		a.NoError(err)
		a.EqualValues(2, res.ID)
		a.EqualValues(file.ID, res.File.ID)
		a.NoError(mock.ExpectationsWereMet())
	}
}

func TestFile_UpdateMetadata(t *testing.T) {
	a := assert.New(t)
	file := &File{}
	file.ID = 1

	// 更新失败
	{
		expectedErr := errors.New("error")
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs(sqlmock.AnyArg(), 1).WillReturnError(expectedErr)
		mock.ExpectRollback()
		a.ErrorIs(file.UpdateMetadata(map[string]string{"1": "1"}), expectedErr)
		a.NoError(mock.ExpectationsWereMet())
	}

	// 成功
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs(sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		a.NoError(file.UpdateMetadata(map[string]string{"1": "1"}))
		a.NoError(mock.ExpectationsWereMet())
		a.Equal("1", file.MetadataSerialized["1"])
	}
}

func TestFile_ShouldLoadThumb(t *testing.T) {
	a := assert.New(t)
	file := &File{
		MetadataSerialized: map[string]string{},
	}
	file.ID = 1

	// 无缩略图
	{
		file.MetadataSerialized[ThumbStatusMetadataKey] = ThumbStatusNotAvailable
		a.False(file.ShouldLoadThumb())
	}

	// 有缩略图
	{
		file.MetadataSerialized[ThumbStatusMetadataKey] = ThumbStatusExist
		a.True(file.ShouldLoadThumb())
	}
}

func TestFile_ThumbFile(t *testing.T) {
	a := assert.New(t)
	file := &File{
		SourceName:         "test",
		MetadataSerialized: map[string]string{},
	}
	file.ID = 1

	a.Equal("test._thumb", file.ThumbFile())
}
