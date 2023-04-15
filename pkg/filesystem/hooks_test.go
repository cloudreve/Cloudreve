package filesystem

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/requestmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestGenericBeforeUpload(t *testing.T) {
	asserts := assert.New(t)
	file := &fsctx.FileStream{
		Size: 5,
		Name: "1.txt",
	}
	ctx := context.Background()
	cache.Set("pack_size_0", uint64(0), 0)
	fs := FileSystem{
		User: &model.User{
			Storage: 0,
			Group: model.Group{
				MaxStorage: 11,
			},
		},
		Policy: &model.Policy{
			MaxSize: 4,
			OptionsSerialized: model.PolicyOption{
				FileType: []string{"txt"},
			},
		},
	}

	asserts.Error(HookValidateFile(ctx, &fs, file))

	file.Size = 1
	file.Name = "1"
	asserts.Error(HookValidateFile(ctx, &fs, file))

	file.Name = "1.txt"
	asserts.NoError(HookValidateFile(ctx, &fs, file))

	file.Name = "1.t/xt"
	asserts.Error(HookValidateFile(ctx, &fs, file))
}

func TestGenericAfterUploadCanceled(t *testing.T) {
	asserts := assert.New(t)
	file := &fsctx.FileStream{
		Size:     5,
		Name:     "TestGenericAfterUploadCanceled",
		SavePath: "TestGenericAfterUploadCanceled",
	}
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{},
	}

	// 成功
	{
		mockHandler := &FileHeaderMock{}
		fs.Handler = mockHandler
		mockHandler.On("Delete", testMock.Anything, testMock.Anything).Return([]string{}, nil)
		err := HookDeleteTempFile(ctx, &fs, file)
		asserts.NoError(err)
		mockHandler.AssertExpectations(t)
	}

	// 失败
	{
		mockHandler := &FileHeaderMock{}
		fs.Handler = mockHandler
		mockHandler.On("Delete", testMock.Anything, testMock.Anything).Return([]string{}, errors.New(""))
		err := HookDeleteTempFile(ctx, &fs, file)
		asserts.NoError(err)
		mockHandler.AssertExpectations(t)
	}

}

func TestGenericAfterUpload(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
		},
		Policy: &model.Policy{},
	}

	ctx := context.Background()
	file := &fsctx.FileStream{
		VirtualPath: "/我的文件",
		Name:        "test.txt",
	}

	// 正常
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	mock.ExpectQuery("SELECT(.+)files").
		WithArgs(1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs("我的文件", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE(.+)storage(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := GenericAfterUpload(ctx, &fs, file)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 文件已存在
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	mock.ExpectQuery("SELECT(.+)files").
		WithArgs(1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs("我的文件", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("test.txt"),
	)
	err = GenericAfterUpload(ctx, &fs, file)
	asserts.Equal(ErrFileExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 文件已存在, 且为上传占位符
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	mock.ExpectQuery("SELECT(.+)files").
		WithArgs(1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs("我的文件", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnRows(
		mock.NewRows([]string{"name", "upload_session_id"}).AddRow("test.txt", "1"),
	)
	err = GenericAfterUpload(ctx, &fs, file)
	asserts.Equal(ErrFileUploadSessionExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 插入失败
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	mock.ExpectQuery("SELECT(.+)files").
		WithArgs(1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs("我的文件", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))

	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()

	err = GenericAfterUpload(ctx, &fs, file)
	asserts.Equal(ErrInsertFileRecord, err)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestFileSystem_Use(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{}

	hook := func(ctx context.Context, fs *FileSystem, fileHeader fsctx.FileHeader) error {
		return nil
	}

	// 添加一个
	fs.Use("BeforeUpload", hook)
	asserts.Len(fs.Hooks["BeforeUpload"], 1)

	// 添加一个
	fs.Use("BeforeUpload", hook)
	asserts.Len(fs.Hooks["BeforeUpload"], 2)

	// 不存在
	fs.Use("BeforeUpload2333", hook)

	asserts.NotPanics(func() {
		for _, hookName := range []string{
			"AfterUpload",
			"AfterValidateFailed",
			"AfterUploadCanceled",
			"BeforeFileDownload",
		} {
			fs.Use(hookName, hook)
		}
	})

}

func TestFileSystem_Trigger(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{},
	}
	ctx := context.Background()

	hook := func(ctx context.Context, fs *FileSystem, fileHeader fsctx.FileHeader) error {
		fs.User.Storage++
		return nil
	}

	// 一个
	fs.Use("BeforeUpload", hook)
	err := fs.Trigger(ctx, "BeforeUpload", nil)
	asserts.NoError(err)
	asserts.Equal(uint64(1), fs.User.Storage)

	// 多个
	fs.Use("BeforeUpload", hook)
	fs.Use("BeforeUpload", hook)
	err = fs.Trigger(ctx, "BeforeUpload", nil)
	asserts.NoError(err)
	asserts.Equal(uint64(4), fs.User.Storage)

	// 多个,有失败
	fs.Use("BeforeUpload", func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
		return errors.New("error")
	})
	fs.Use("BeforeUpload", func(ctx context.Context, fs *FileSystem, file fsctx.FileHeader) error {
		asserts.Fail("following hooks executed")
		return nil
	})
	err = fs.Trigger(ctx, "BeforeUpload", nil)
	asserts.Error(err)
}

func TestHookValidateCapacity(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("pack_size_1", uint64(0), 0)
	fs := &FileSystem{User: &model.User{
		Model:   gorm.Model{ID: 1},
		Storage: 0,
		Group: model.Group{
			MaxStorage: 11,
		},
	}}
	ctx := context.Background()
	file := &fsctx.FileStream{Size: 11}
	{
		err := HookValidateCapacity(ctx, fs, file)
		asserts.NoError(err)
	}
	{
		file.Size = 12
		err := HookValidateCapacity(ctx, fs, file)
		asserts.Error(err)
	}
}

func TestHookValidateCapacityDiff(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Group: model.Group{
			MaxStorage: 11,
		},
	}}
	file := model.File{Size: 10}
	ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)

	// 无需操作
	{
		a.NoError(HookValidateCapacityDiff(ctx, fs, &fsctx.FileStream{Size: 10}))
	}

	// 需要验证
	{
		a.Error(HookValidateCapacityDiff(ctx, fs, &fsctx.FileStream{Size: 12}))
	}

}

func TestHookResetPolicy(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	// 成功
	{
		file := model.File{PolicyID: 2}
		cache.Deletes([]string{"2"}, "policy_")
		mock.ExpectQuery("SELECT(.+)policies(.+)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(2, "local"))
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, file)
		err := HookResetPolicy(ctx, fs, nil)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文文件不存在
	{
		cache.Deletes([]string{"2"}, "policy_")
		ctx := context.Background()
		err := HookResetPolicy(ctx, fs, nil)
		asserts.Error(err)
	}
}

func TestHookCleanFileContent(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	file := &fsctx.FileStream{SavePath: "123/123"}
	handlerMock := FileHeaderMock{}
	handlerMock.On("Put", testMock.Anything, testMock.Anything).Return(errors.New("error"))
	fs.Handler = handlerMock
	err := HookCleanFileContent(context.Background(), fs, file)
	asserts.Error(err)
	handlerMock.AssertExpectations(t)
}

func TestHookClearFileSize(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	// 成功
	{
		ctx := context.WithValue(
			context.Background(),
			fsctx.FileModelCtx,
			model.File{Model: gorm.Model{ID: 1}, Size: 10},
		)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").
			WithArgs("", 0, sqlmock.AnyArg(), 1, 10).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)users(.+)").
			WithArgs(10, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := HookClearFileSize(ctx, fs, nil)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文对象不存在
	{
		ctx := context.Background()
		err := HookClearFileSize(ctx, fs, nil)
		asserts.Error(err)
	}

}

func TestHookUpdateSourceName(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	// 成功
	{
		originFile := model.File{
			Model:      gorm.Model{ID: 1},
			SourceName: "new.txt",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs("", "new.txt", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := HookUpdateSourceName(ctx, fs, nil)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文错误
	{
		ctx := context.Background()
		err := HookUpdateSourceName(ctx, fs, nil)
		asserts.Error(err)
	}
}

func TestGenericAfterUpdate(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	// 成功 是图像文件
	{
		originFile := model.File{
			Model:   gorm.Model{ID: 1},
			PicInfo: "1,1",
		}
		newFile := &fsctx.FileStream{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)

		handlerMock := FileHeaderMock{}
		handlerMock.On("Delete", testMock.Anything, []string{"._thumb"}).Return([]string{}, nil)
		fs.Handler = handlerMock
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)files(.+)").
			WithArgs("", 10, sqlmock.AnyArg(), 1, 0).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE(.+)users(.+)").
			WithArgs(10, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := GenericAfterUpdate(ctx, fs, newFile)

		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 原始文件上下文不存在
	{
		newFile := &fsctx.FileStream{Size: 10}
		ctx := context.Background()
		err := GenericAfterUpdate(ctx, fs, newFile)
		asserts.Error(err)
	}

	// 无法更新数据库容量
	// 成功 是图像文件
	{
		originFile := model.File{
			Model:   gorm.Model{ID: 1},
			PicInfo: "1,1",
		}
		newFile := &fsctx.FileStream{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs("", 10, sqlmock.AnyArg(), 1, 0).
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		err := GenericAfterUpdate(ctx, fs, newFile)

		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestSlaveAfterUpload(t *testing.T) {
	asserts := assert.New(t)
	conf.SystemConfig.Mode = "slave"
	fs, err := NewAnonymousFileSystem()
	conf.SystemConfig.Mode = "master"
	asserts.NoError(err)

	// 成功
	{
		clientMock := requestmock.RequestMock{}
		clientMock.On(
			"Request",
			"POST",
			"http://test/callbakc",
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: nil,
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":0}`)),
			},
		})
		request.GeneralClient = clientMock
		file := &fsctx.FileStream{
			Size:        10,
			VirtualPath: "/my",
			Name:        "test.txt",
			SavePath:    "/not_exist",
		}
		err := SlaveAfterUpload(&serializer.UploadSession{Callback: "http://test/callbakc"})(context.Background(), fs, file)
		clientMock.AssertExpectations(t)
		asserts.NoError(err)
	}

	// 跳过回调
	{
		file := &fsctx.FileStream{
			Size:        10,
			VirtualPath: "/my",
			Name:        "test.txt",
			SavePath:    "/not_exist",
		}
		err := SlaveAfterUpload(&serializer.UploadSession{})(context.Background(), fs, file)
		asserts.NoError(err)
	}
}

func TestFileSystem_CleanHooks(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{
		User: &model.User{
			Model: gorm.Model{ID: 1},
		},
		Hooks: map[string][]Hook{
			"hook1": []Hook{},
			"hook2": []Hook{},
			"hook3": []Hook{},
		},
	}

	// 清理一个
	{
		fs.CleanHooks("hook2")
		asserts.Len(fs.Hooks, 2)
		asserts.Contains(fs.Hooks, "hook1")
		asserts.Contains(fs.Hooks, "hook3")
	}

	// 清理全部
	{
		fs.CleanHooks("")
		asserts.Len(fs.Hooks, 0)
	}
}

func TestHookCancelContext(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{}
	ctx, cancel := context.WithCancel(context.Background())

	// empty ctx
	{
		asserts.NoError(HookCancelContext(ctx, fs, nil))
		select {
		case <-ctx.Done():
			t.Errorf("Channel should not be closed")
		default:

		}
	}

	// with cancel ctx
	{
		ctx = context.WithValue(ctx, fsctx.CancelFuncCtx, cancel)
		asserts.NoError(HookCancelContext(ctx, fs, nil))
		_, ok := <-ctx.Done()
		asserts.False(ok)
	}
}

func TestHookClearFileHeaderSize(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{Size: 10}
	a.NoError(HookClearFileHeaderSize(context.Background(), fs, file))
	a.EqualValues(0, file.Size)
}

func TestHookTruncateFileTo(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{}
	a.NoError(HookTruncateFileTo(0)(context.Background(), fs, file))

	fs.Handler = local.Driver{}
	a.Error(HookTruncateFileTo(0)(context.Background(), fs, file))
}

func TestHookChunkUploaded(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{
		AppendStart: 10,
		Size:        10,
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 20, sqlmock.AnyArg(), 1, 0).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE(.+)users(.+)").
		WithArgs(20, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(HookChunkUploaded(context.Background(), fs, file))
	a.NoError(mock.ExpectationsWereMet())
}

func TestHookChunkUploadFailed(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{
		AppendStart: 10,
		Size:        10,
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WithArgs("", 10, sqlmock.AnyArg(), 1, 0).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE(.+)users(.+)").
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(HookChunkUploadFailed(context.Background(), fs, file))
	a.NoError(mock.ExpectationsWereMet())
}

func TestHookPopPlaceholderToFile(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(HookPopPlaceholderToFile("1,1")(context.Background(), fs, file))
	a.NoError(mock.ExpectationsWereMet())
}

func TestHookPopPlaceholderToFileBySuffix(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{
		Policy: &model.Policy{Type: "cos"},
	}
	file := &fsctx.FileStream{
		Name: "1.png",
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	a.NoError(HookPopPlaceholderToFile("")(context.Background(), fs, file))
	a.NoError(mock.ExpectationsWereMet())
}

func TestHookDeleteUploadSession(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	cache.Set(UploadSessionCachePrefix+"TestHookDeleteUploadSession", "", 0)
	a.NoError(HookDeleteUploadSession("TestHookDeleteUploadSession")(context.Background(), fs, file))
	_, ok := cache.Get(UploadSessionCachePrefix + "TestHookDeleteUploadSession")
	a.False(ok)
}
func TestNewWebdavAfterUploadHook(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{}
	file := &fsctx.FileStream{
		Model: &model.File{
			Model: gorm.Model{ID: 1},
		},
	}

	req, _ := http.NewRequest("get", "http://localhost", nil)
	req.Header.Add("X-Oc-Mtime", "1681521402")
	req.Header.Add("OC-Checksum", "checksum")
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := NewWebdavAfterUploadHook(req)(context.Background(), fs, file)
	a.NoError(err)
	a.NoError(mock.ExpectationsWereMet())

}
