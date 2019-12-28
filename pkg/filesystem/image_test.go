package filesystem

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func CreateTestImage() *os.File {
	file, err := os.Create("TestFileSystem_GenerateThumbnail.jpeg")
	alpha := image.NewAlpha(image.Rect(0, 0, 500, 200))
	jpeg.Encode(file, alpha, nil)
	if err != nil {
		fmt.Println(err)
	}
	_, _ = file.Seek(0, 0)
	return file
}

func TestFileSystem_GetThumb(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{ID: 1},
		},
	}
	ctx := context.Background()

	// 正常
	{
		testHandler := new(FileHeaderMock)
		testHandler.On("Thumb", testMock.Anything, "123.jpg").Return(&response.ContentResponse{URL: "123"}, nil)
		fs.Handler = testHandler
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(10, 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "pic_info", "source_name"}).
					AddRow(10, "10,10", "123.jpg"),
			)

		res, err := fs.GetThumb(ctx, 10)
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
		asserts.NoError(err)
		asserts.Equal("123", res.URL)
	}

	// 文件不存在
	{

		mock.ExpectQuery("SELECT(.+)").
			WithArgs(10, 1).
			WillReturnRows(
				sqlmock.NewRows(
					[]string{"id", "pic_info", "source_name"}),
			)

		_, err := fs.GetThumb(ctx, 10)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}

func TestFileSystem_GenerateThumbnail(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{ID: 1},
		},
	}
	ctx := context.Background()

	// 成功
	{
		src := CreateTestImage()
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "TestFileSystem_GenerateThumbnail.jpeg").Return(src, nil)
		fs.Handler = testHandler

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		file := &model.File{
			Model:      gorm.Model{ID: 1},
			Name:       "123.jpg",
			SourceName: "TestFileSystem_GenerateThumbnail.jpeg",
		}

		fs.GenerateThumbnail(ctx, file)
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
		asserts.True(util.Exists("TestFileSystem_GenerateThumbnail.jpeg" + conf.ThumbConfig.FileSuffix))

	}

	// 成功,不进行数据库更新
	{
		src := CreateTestImage()
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "TestFileSystem_GenerateThumbnail.jpeg").Return(src, nil)
		fs.Handler = testHandler

		file := &model.File{
			Name:       "123.jpg",
			SourceName: "TestFileSystem_GenerateThumbnail.jpeg",
		}

		fs.GenerateThumbnail(ctx, file)
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)
		asserts.True(util.Exists("TestFileSystem_GenerateThumbnail.jpeg" + conf.ThumbConfig.FileSuffix))

	}

	// 更新信息失败后删除文件
	{
		src := CreateTestImage()
		testHandler := new(FileHeaderMock)
		testHandler.On("Get", testMock.Anything, "TestFileSystem_GenerateThumbnail.jpeg").Return(src, nil)
		testHandler.On("Delete", testMock.Anything, testMock.Anything).Return([]string{}, nil)
		fs.Handler = testHandler

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WillReturnError(errors.New("error"))
		mock.ExpectRollback()

		file := &model.File{
			Model:      gorm.Model{ID: 1},
			Name:       "123.jpg",
			SourceName: "TestFileSystem_GenerateThumbnail.jpeg",
		}

		fs.GenerateThumbnail(ctx, file)
		asserts.NoError(mock.ExpectationsWereMet())
		testHandler.AssertExpectations(t)

	}

	// 不能生成缩略图
	{
		file := &model.File{
			Model:      gorm.Model{ID: 1},
			Name:       "123.123",
			SourceName: "TestFileSystem_GenerateThumbnail.jpeg",
		}

		fs.GenerateThumbnail(ctx, file)
		asserts.NoError(mock.ExpectationsWereMet())
	}

}

func TestFileSystem_GenerateThumbnailSize(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{
		User: &model.User{
			Model: gorm.Model{ID: 1},
		},
	}
	asserts.NotPanics(func() {
		_, _ = fs.GenerateThumbnailSize(0, 0)
	})

}
