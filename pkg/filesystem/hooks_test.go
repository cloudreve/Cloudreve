package filesystem

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestGenericBeforeUpload(t *testing.T) {
	asserts := assert.New(t)
	file := local.FileStream{
		Size: 5,
		Name: "1.txt",
	}
	cache.Set("pack_size_0", uint64(0), 0)
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, file)
	fs := FileSystem{
		User: &model.User{
			Storage: 0,
			Group: model.Group{
				MaxStorage: 11,
			},
			Policy: model.Policy{
				MaxSize: 4,
				OptionsSerialized: model.PolicyOption{
					FileType: []string{"txt"},
				},
			},
		},
	}

	asserts.Error(HookValidateFile(ctx, &fs))

	file.Size = 1
	file.Name = "1"
	ctx = context.WithValue(context.Background(), fsctx.FileHeaderCtx, file)
	asserts.Error(HookValidateFile(ctx, &fs))

	file.Name = "1.txt"
	ctx = context.WithValue(context.Background(), fsctx.FileHeaderCtx, file)
	asserts.NoError(HookValidateFile(ctx, &fs))

	file.Name = "1.t/xt"
	ctx = context.WithValue(context.Background(), fsctx.FileHeaderCtx, file)
	asserts.Error(HookValidateFile(ctx, &fs))
}

func TestGenericAfterUploadCanceled(t *testing.T) {
	asserts := assert.New(t)
	f, err := os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()
	file := local.FileStream{
		Size: 5,
		Name: "TestGenericAfterUploadCanceled",
	}
	ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "TestGenericAfterUploadCanceled")
	ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)
	fs := FileSystem{
		User:    &model.User{Storage: 5},
		Handler: local.Driver{},
	}

	// 成功
	err = HookDeleteTempFile(ctx, &fs)
	asserts.NoError(err)
	err = HookGiveBackCapacity(ctx, &fs)
	asserts.NoError(err)
	asserts.Equal(uint64(0), fs.User.Storage)

	f, err = os.Create("TestGenericAfterUploadCanceled")
	asserts.NoError(err)
	f.Close()

	// 容量不能再降低
	err = HookGiveBackCapacity(ctx, &fs)
	asserts.Error(err)

	//文件不存在
	fs.User.Storage = 5
	err = HookDeleteTempFile(ctx, &fs)
	asserts.NoError(err)
}

func TestGenericAfterUpload(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{
				ID: 1,
			},
		},
	}

	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{
		VirtualPath: "/我的文件",
		Name:        "test.txt",
	})
	ctx = context.WithValue(ctx, fsctx.SavePathCtx, "")

	// 正常
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := GenericAfterUpload(ctx, &fs)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 路径不存在
	mock.ExpectQuery("SELECT(.+)folders(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}),
	)
	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrRootProtected, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 文件已存在
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnRows(
		mock.NewRows([]string{"name"}).AddRow("test.txt"),
	)
	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrFileExisted, err)
	asserts.NoError(mock.ExpectationsWereMet())

	// 插入失败
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
	// 1
	mock.ExpectQuery("SELECT(.+)").
		WithArgs(1, 1, "我的文件").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(2, 1))

	mock.ExpectQuery("SELECT(.+)files(.+)").WillReturnError(errors.New("not found"))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)files(.+)").WillReturnError(errors.New("error"))
	mock.ExpectRollback()

	err = GenericAfterUpload(ctx, &fs)
	asserts.Equal(ErrInsertFileRecord, err)
	asserts.NoError(mock.ExpectationsWereMet())

}

func TestFileSystem_Use(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{}

	hook := func(ctx context.Context, fs *FileSystem) error {
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

	hook := func(ctx context.Context, fs *FileSystem) error {
		fs.User.Storage++
		return nil
	}

	// 一个
	fs.Use("BeforeUpload", hook)
	err := fs.Trigger(ctx, "BeforeUpload")
	asserts.NoError(err)
	asserts.Equal(uint64(1), fs.User.Storage)

	// 多个
	fs.Use("BeforeUpload", hook)
	fs.Use("BeforeUpload", hook)
	err = fs.Trigger(ctx, "BeforeUpload")
	asserts.NoError(err)
	asserts.Equal(uint64(4), fs.User.Storage)
}

func TestHookIsFileExist(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{
			ID: 1,
		},
	}}
	ctx := context.WithValue(context.Background(), fsctx.PathCtx, "/test.txt")
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "test.txt").WillReturnRows(
			sqlmock.NewRows([]string{"Name"}).AddRow("s"),
		)
		err := HookIsFileExist(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}
	{
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow(1, 1))
		mock.ExpectQuery("SELECT(.+)").WithArgs(uint(1), "test.txt").WillReturnRows(
			sqlmock.NewRows([]string{"Name"}),
		)
		err := HookIsFileExist(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

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
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{Size: 10})
	{
		err := HookValidateCapacity(ctx, fs)
		asserts.NoError(err)
	}
	{
		err := HookValidateCapacity(ctx, fs)
		asserts.Error(err)
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
		err := HookResetPolicy(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文文件不存在
	{
		cache.Deletes([]string{"2"}, "policy_")
		ctx := context.Background()
		err := HookResetPolicy(ctx, fs)
		asserts.Error(err)
	}
}

func TestHookChangeCapacity(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("pack_size_1", uint64(0), 0)

	// 容量增加 失败
	{
		fs := &FileSystem{User: &model.User{
			Model: gorm.Model{ID: 1},
		}}

		newFile := local.FileStream{Size: 10}
		oldFile := model.File{Size: 9}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, oldFile)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, newFile)
		err := HookChangeCapacity(ctx, fs)
		asserts.Equal(ErrInsufficientCapacity, err)
	}

	// 容量增加 成功
	{
		fs := &FileSystem{User: &model.User{
			Model: gorm.Model{ID: 1},
			Group: model.Group{MaxStorage: 1},
		}}

		newFile := local.FileStream{Size: 10}
		oldFile := model.File{Size: 9}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, oldFile)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, newFile)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs(1, sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		err := HookChangeCapacity(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal(uint64(1), fs.User.Storage)
	}

	// 容量减少
	{
		fs := &FileSystem{User: &model.User{
			Model:   gorm.Model{ID: 1},
			Storage: 1,
		}}

		newFile := local.FileStream{Size: 9}
		oldFile := model.File{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, oldFile)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, newFile)
		err := HookChangeCapacity(ctx, fs)
		asserts.NoError(err)
		asserts.Equal(uint64(0), fs.User.Storage)
	}
}

func TestHookCleanFileContent(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{User: &model.User{
		Model: gorm.Model{ID: 1},
	}}

	ctx := context.WithValue(context.Background(), fsctx.SavePathCtx, "123/123")
	handlerMock := FileHeaderMock{}
	handlerMock.On("Put", testMock.Anything, testMock.Anything, "123/123").Return(errors.New("error"))
	fs.Handler = handlerMock
	err := HookCleanFileContent(ctx, fs)
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
			model.File{Model: gorm.Model{ID: 1}},
		)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs(0, sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := HookClearFileSize(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文对象不存在
	{
		ctx := context.Background()
		err := HookClearFileSize(ctx, fs)
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
		mock.ExpectExec("UPDATE(.+)").WithArgs("new.txt", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := HookUpdateSourceName(ctx, fs)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 上下文错误
	{
		ctx := context.Background()
		err := HookUpdateSourceName(ctx, fs)
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
		newFile := local.FileStream{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, newFile)

		handlerMock := FileHeaderMock{}
		handlerMock.On("Delete", testMock.Anything, []string{"._thumb"}).Return([]string{}, nil)
		fs.Handler = handlerMock
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs(10, sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := GenericAfterUpdate(ctx, fs)

		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
	}

	// 新文件上下文不存在
	{
		originFile := model.File{
			Model:   gorm.Model{ID: 1},
			PicInfo: "1,1",
		}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)
		err := GenericAfterUpdate(ctx, fs)
		asserts.Error(err)
	}

	// 原始文件上下文不存在
	{
		newFile := local.FileStream{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, newFile)
		err := GenericAfterUpdate(ctx, fs)
		asserts.Error(err)
	}

	// 无法更新数据库容量
	// 成功 是图像文件
	{
		originFile := model.File{
			Model:   gorm.Model{ID: 1},
			PicInfo: "1,1",
		}
		newFile := local.FileStream{Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, originFile)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, newFile)

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WithArgs(10, sqlmock.AnyArg(), 1).
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		err := GenericAfterUpdate(ctx, fs)

		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestHookSlaveUploadValidate(t *testing.T) {
	asserts := assert.New(t)
	conf.SystemConfig.Mode = "slave"
	fs, err := NewAnonymousFileSystem()
	conf.SystemConfig.Mode = "master"
	asserts.NoError(err)

	// 正常
	{
		policy := serializer.UploadPolicy{
			SavePath:         "",
			MaxSize:          10,
			AllowedExtension: nil,
		}
		file := local.FileStream{Name: "1.txt", Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.UploadPolicyCtx, policy)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)
		asserts.NoError(HookSlaveUploadValidate(ctx, fs))
	}

	// 尺寸太大
	{
		policy := serializer.UploadPolicy{
			SavePath:         "",
			MaxSize:          10,
			AllowedExtension: nil,
		}
		file := local.FileStream{Name: "1.txt", Size: 11}
		ctx := context.WithValue(context.Background(), fsctx.UploadPolicyCtx, policy)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)
		asserts.Equal(ErrFileSizeTooBig, HookSlaveUploadValidate(ctx, fs))
	}

	// 文件名非法
	{
		policy := serializer.UploadPolicy{
			SavePath:         "",
			MaxSize:          10,
			AllowedExtension: nil,
		}
		file := local.FileStream{Name: "/1.txt", Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.UploadPolicyCtx, policy)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)
		asserts.Equal(ErrIllegalObjectName, HookSlaveUploadValidate(ctx, fs))
	}

	// 扩展名非法
	{
		policy := serializer.UploadPolicy{
			SavePath:         "",
			MaxSize:          10,
			AllowedExtension: []string{"jpg"},
		}
		file := local.FileStream{Name: "1.txt", Size: 10}
		ctx := context.WithValue(context.Background(), fsctx.UploadPolicyCtx, policy)
		ctx = context.WithValue(ctx, fsctx.FileHeaderCtx, file)
		asserts.Equal(ErrFileExtensionNotAllowed, HookSlaveUploadValidate(ctx, fs))
	}

}

type ClientMock struct {
	testMock.Mock
}

func (m ClientMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	args := m.Called(method, target, body, opts)
	return args.Get(0).(*request.Response)
}

func TestSlaveAfterUpload(t *testing.T) {
	asserts := assert.New(t)
	conf.SystemConfig.Mode = "slave"
	fs, err := NewAnonymousFileSystem()
	conf.SystemConfig.Mode = "master"
	asserts.NoError(err)

	// 成功
	{
		clientMock := ClientMock{}
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
		ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{
			Size:        10,
			VirtualPath: "/my",
			Name:        "test.txt",
		})
		ctx = context.WithValue(ctx, fsctx.UploadPolicyCtx, serializer.UploadPolicy{
			CallbackURL: "http://test/callbakc",
		})
		ctx = context.WithValue(ctx, fsctx.SavePathCtx, "/not_exist")
		err := SlaveAfterUpload(ctx, fs)
		clientMock.AssertExpectations(t)
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
		asserts.NoError(HookCancelContext(ctx, fs))
		select {
		case <-ctx.Done():
			t.Errorf("Channel should not be closed")
		default:

		}
	}

	// with cancel ctx
	{
		ctx = context.WithValue(ctx, fsctx.CancelFuncCtx, cancel)
		asserts.NoError(HookCancelContext(ctx, fs))
		_, ok := <-ctx.Done()
		asserts.False(ok)
	}
}

func TestHookGiveBackCapacity(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{
		User: &model.User{
			Model:   gorm.Model{ID: 1},
			Storage: 10,
		},
	}
	ctx := context.WithValue(context.Background(), fsctx.FileHeaderCtx, local.FileStream{Size: 1})

	// without once limit
	{
		asserts.NoError(HookGiveBackCapacity(ctx, fs))
		asserts.EqualValues(9, fs.User.Storage)
		asserts.NoError(HookGiveBackCapacity(ctx, fs))
		asserts.EqualValues(8, fs.User.Storage)
	}

	// with once limit
	{
		ctx = context.WithValue(ctx, fsctx.ValidateCapacityOnceCtx, &sync.Once{})
		asserts.NoError(HookGiveBackCapacity(ctx, fs))
		asserts.EqualValues(7, fs.User.Storage)
		asserts.NoError(HookGiveBackCapacity(ctx, fs))
		asserts.EqualValues(7, fs.User.Storage)
	}
}
