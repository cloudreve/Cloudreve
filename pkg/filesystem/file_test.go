package filesystem

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileSystem_AddFile(t *testing.T) {
	asserts := assert.New(t)
	file := local.FileStream{
		Size: 5,
		Name: "1.txt",
	}
	folder := model.Folder{
		Model: gorm.Model{
			ID: 1,
		},
	}
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, file)
	ctx = context.WithValue(ctx, fsctx.SavePathCtx, "/Uploads/1_sad.txt")

	_, err := fs.AddFile(ctx, &folder)

	asserts.Error(err)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	f, err := fs.AddFile(ctx, &folder)

	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Equal("/Uploads/1_sad.txt", f.SourceName)
}

func TestFileSystem_GetContent(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}

	// 文件不存在
	rs, err := fs.GetContent(ctx, "not exist file")
	asserts.Equal(ErrObjectNotExist, err)
	asserts.Nil(rs)
	fs.CleanTargets()

	// 未知存储策略
	file, err := os.Create("TestFileSystem_GetContent.txt")
	asserts.NoError(err)
	_ = file.Close()

	cache.Deletes([]string{"1"}, "policy_")
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "TestFileSystem_GetContent.txt", 1))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "unknown"))

	rs, err = fs.GetContent(ctx, "/TestFileSystem_GetContent.txt")
	asserts.Error(err)
	asserts.NoError(mock.ExpectationsWereMet())
	fs.CleanTargets()

	// 打开文件失败
	cache.Deletes([]string{"1"}, "policy_")
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "TestFileSystem_GetContent.txt", 1))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source_name"}).AddRow(1, "local", "not exist"))

	rs, err = fs.GetContent(ctx, "/TestFileSystem_GetContent.txt")
	asserts.Equal(serializer.CodeIOFailed, err.(serializer.AppError).Code)
	asserts.NoError(mock.ExpectationsWereMet())
	fs.CleanTargets()

	// 打开成功
	cache.Deletes([]string{"1"}, "policy_")
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetContent.txt", 1, "TestFileSystem_GetContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	rs, err = fs.GetContent(ctx, "/TestFileSystem_GetContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_GetDownloadContent(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
			Policy: model.Policy{
				Model: gorm.Model{
					ID: 1,
				},
			},
		},
	}
	file, err := os.Create("TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	_ = file.Close()

	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetDownloadContent.txt", 1, "TestFileSystem_GetDownloadContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	// 无限速
	_, err = fs.GetDownloadContent(ctx, "/TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	fs.CleanTargets()

	// 有限速
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id", "source_name"}).AddRow(1, "TestFileSystem_GetDownloadContent.txt", 1, "TestFileSystem_GetDownloadContent.txt"))
	mock.ExpectQuery("SELECT(.+)poli(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "local"))

	fs.User.Group.SpeedLimit = 1
	_, err = fs.GetDownloadContent(ctx, "/TestFileSystem_GetDownloadContent.txt")
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestFileSystem_GroupFileByPolicy(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	files := []model.File{
		model.File{
			PolicyID: 1,
			Name:     "1_1.txt",
		},
		model.File{
			PolicyID: 2,
			Name:     "2_1.txt",
		},
		model.File{
			PolicyID: 3,
			Name:     "3_1.txt",
		},
		model.File{
			PolicyID: 2,
			Name:     "2_2.txt",
		},
		model.File{
			PolicyID: 1,
			Name:     "1_2.txt",
		},
	}
	fs := FileSystem{}
	policyGroup := fs.GroupFileByPolicy(ctx, files)
	asserts.Equal(map[uint][]*model.File{
		1: {&files[0], &files[4]},
		2: {&files[1], &files[3]},
		3: {&files[2]},
	}, policyGroup)
}

func TestFileSystem_deleteGroupedFile(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{}
	files := []model.File{
		{
			PolicyID:   1,
			Name:       "1_1.txt",
			SourceName: "1_1.txt",
			Policy:     model.Policy{Model: gorm.Model{ID: 1}, Type: "local"},
		},
		{
			PolicyID:   2,
			Name:       "2_1.txt",
			SourceName: "2_1.txt",
			Policy:     model.Policy{Model: gorm.Model{ID: 1}, Type: "local"},
		},
		{
			PolicyID:   3,
			Name:       "3_1.txt",
			SourceName: "3_1.txt",
			Policy:     model.Policy{Model: gorm.Model{ID: 1}, Type: "local"},
		},
		{
			PolicyID:   2,
			Name:       "2_2.txt",
			SourceName: "2_2.txt",
			Policy:     model.Policy{Model: gorm.Model{ID: 1}, Type: "local"},
		},
		{
			PolicyID:   1,
			Name:       "1_2.txt",
			SourceName: "1_2.txt",
			Policy:     model.Policy{Model: gorm.Model{ID: 1}, Type: "local"},
		},
	}

	// 全部失败
	{
		failed := fs.deleteGroupedFile(ctx, fs.GroupFileByPolicy(ctx, files))
		asserts.Equal(map[uint][]string{
			1: {"1_1.txt", "1_2.txt"},
			2: {"2_1.txt", "2_2.txt"},
			3: {"3_1.txt"},
		}, failed)
	}
	// 部分失败
	{
		file, err := os.Create("1_1.txt")
		asserts.NoError(err)
		_ = file.Close()
		failed := fs.deleteGroupedFile(ctx, fs.GroupFileByPolicy(ctx, files))
		asserts.Equal(map[uint][]string{
			1: {"1_2.txt"},
			2: {"2_1.txt", "2_2.txt"},
			3: {"3_1.txt"},
		}, failed)
	}
	// 部分失败,包含整组未知存储策略导致的失败
	{
		file, err := os.Create("1_1.txt")
		asserts.NoError(err)
		_ = file.Close()

		files[1].Policy.Type = "unknown"
		files[3].Policy.Type = "unknown"
		failed := fs.deleteGroupedFile(ctx, fs.GroupFileByPolicy(ctx, files))
		asserts.Equal(map[uint][]string{
			1: {"1_2.txt"},
			2: {"2_1.txt", "2_2.txt"},
			3: {"3_1.txt"},
		}, failed)
	}
}

func TestFileSystem_GetSource(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	auth.General = auth.HMACAuth{SecretKey: []byte("123")}

	// 正常
	{
		fs := FileSystem{
			User: &model.User{Model: gorm.Model{ID: 1}},
		}
		// 清空缓存
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(2, 35, "1.txt"),
			)
		// 查找上传策略
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "type", "is_origin_link_enable"}).
					AddRow(35, "local", true),
			)
		// 查找站点URL
		mock.ExpectQuery("SELECT(.+)").WithArgs("siteURL").WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(1, "https://cloudreve.org"))

		sourceURL, err := fs.GetSource(ctx, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.NotEmpty(sourceURL)
		fs.CleanTargets()
	}

	// 文件不存在
	{
		fs := FileSystem{
			User: &model.User{Model: gorm.Model{ID: 1}},
		}
		// 清空缓存
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}),
			)

		sourceURL, err := fs.GetSource(ctx, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(ErrObjectNotExist.Code, err.(serializer.AppError).Code)
		asserts.Empty(sourceURL)
		fs.CleanTargets()
	}

	// 未知上传策略
	{
		fs := FileSystem{
			User: &model.User{Model: gorm.Model{ID: 1}},
		}
		// 清空缓存
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(2, 35, "1.txt"),
			)
		// 查找上传策略
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "type", "is_origin_link_enable"}).
					AddRow(35, "?", true),
			)

		sourceURL, err := fs.GetSource(ctx, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Empty(sourceURL)
		fs.CleanTargets()
	}

	// 不允许获取外链
	{
		fs := FileSystem{
			User: &model.User{Model: gorm.Model{ID: 1}},
		}
		// 清空缓存
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(2, 1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "policy_id", "source_name"}).
					AddRow(2, 35, "1.txt"),
			)
		// 查找上传策略
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "type", "is_origin_link_enable"}).
					AddRow(35, "local", false),
			)

		sourceURL, err := fs.GetSource(ctx, 2)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Equal(serializer.CodePolicyNotAllowed, err.(serializer.AppError).Code)
		asserts.Empty(sourceURL)
		fs.CleanTargets()
	}
}

func TestFileSystem_GetDownloadURL(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{Model: gorm.Model{ID: 1}},
	}
	auth.General = auth.HMACAuth{SecretKey: []byte("123")}

	// 正常
	{
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		err = cache.Deletes([]string{"35"}, "policy_")
		err = cache.Deletes([]string{"download_timeout"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "1.txt", 1))
		// 查找上传策略
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "type", "is_origin_link_enable"}).
					AddRow(35, "local", true),
			)
		// 相关设置
		mock.ExpectQuery("SELECT(.+)").WithArgs("download_timeout").WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(1, "20"))
		mock.ExpectQuery("SELECT(.+)").WithArgs("siteURL").WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(1, "https://cloudreve.org"))
		downloadURL, err := fs.GetDownloadURL(ctx, "/1.txt", "download_timeout")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.NotEmpty(downloadURL)
		fs.CleanTargets()
	}

	// 文件不存在
	{
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		err = cache.Deletes([]string{"35"}, "policy_")
		err = cache.Deletes([]string{"download_timeout"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}))

		downloadURL, err := fs.GetDownloadURL(ctx, "/1.txt", "download_timeout")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Empty(downloadURL)
		fs.CleanTargets()
	}

	// 未知存储策略
	{
		err := cache.Deletes([]string{"siteURL"}, "setting_")
		err = cache.Deletes([]string{"35"}, "policy_")
		err = cache.Deletes([]string{"download_timeout"}, "setting_")
		asserts.NoError(err)
		// 查找文件
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policy_id"}).AddRow(1, "1.txt", 1))
		// 查找上传策略
		mock.ExpectQuery("SELECT(.+)").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "type", "is_origin_link_enable"}).
					AddRow(35, "unknown", true),
			)

		downloadURL, err := fs.GetDownloadURL(ctx, "/1.txt", "download_timeout")
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Empty(downloadURL)
		fs.CleanTargets()
	}
}

func TestFileSystem_GetPhysicalFileContent(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{},
	}

	// 文件不存在
	{
		rs, err := fs.GetPhysicalFileContent(ctx, "not_exist.txt")
		asserts.Error(err)
		asserts.Nil(rs)
	}

	// 成功
	{
		testFile, err := os.Create("GetPhysicalFileContent.txt")
		asserts.NoError(err)
		asserts.NoError(testFile.Close())

		rs, err := fs.GetPhysicalFileContent(ctx, "GetPhysicalFileContent.txt")
		asserts.NoError(err)
		asserts.NoError(rs.Close())
		asserts.NotNil(rs)
	}
}

func TestFileSystem_Preview(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()

	// 文件不存在
	{
		fs := FileSystem{
			User: &model.User{},
		}
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		resp, err := fs.Preview(ctx, "/1.txt", false)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Nil(resp)
	}

	// 直接返回文件内容,找不到文件
	{
		fs := FileSystem{
			User: &model.User{},
		}
		fs.FileTarget = []model.File{
			{
				PolicyID: 1,
				Policy: model.Policy{
					Model: gorm.Model{ID: 1},
					Type:  "local",
				},
			},
		}
		resp, err := fs.Preview(ctx, "/1.txt", false)
		asserts.Error(err)
		asserts.Nil(resp)
	}

	// 直接返回文件内容
	{
		fs := FileSystem{
			User: &model.User{},
		}
		fs.FileTarget = []model.File{
			{
				SourceName: "tests/file1.txt",
				PolicyID:   1,
				Policy: model.Policy{
					Model: gorm.Model{ID: 1},
					Type:  "local",
				},
			},
		}
		resp, err := fs.Preview(ctx, "/1.txt", false)
		asserts.NoError(err)
		asserts.NotNil(resp)
		asserts.False(resp.Redirect)
		asserts.NoError(resp.Content.Close())
	}

	// 需要重定向，成功
	{
		fs := FileSystem{
			User: &model.User{},
		}
		fs.FileTarget = []model.File{
			{
				SourceName: "tests/file1.txt",
				PolicyID:   1,
				Policy: model.Policy{
					Model: gorm.Model{ID: 1},
					Type:  "remote",
				},
			},
		}
		asserts.NoError(cache.Set("setting_preview_timeout", "233", 0))
		resp, err := fs.Preview(ctx, "/1.txt", false)
		asserts.NoError(err)
		asserts.NotNil(resp)
		asserts.True(resp.Redirect)
	}

	// 文本文件，大小超出限制
	{
		fs := FileSystem{
			User: &model.User{},
		}
		fs.FileTarget = []model.File{
			{
				SourceName: "tests/file1.txt",
				PolicyID:   1,
				Policy: model.Policy{
					Model: gorm.Model{ID: 1},
					Type:  "remote",
				},
				Size: 11,
			},
		}
		asserts.NoError(cache.Set("setting_maxEditSize", "10", 0))
		resp, err := fs.Preview(ctx, "/1.txt", true)
		asserts.Equal(ErrFileSizeTooBig, err)
		asserts.Nil(resp)
	}
}
