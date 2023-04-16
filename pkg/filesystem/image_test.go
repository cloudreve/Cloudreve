package filesystem

import (
	"context"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/thumbmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	testMock "github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystem_GetThumb(t *testing.T) {
	a := assert.New(t)
	fs := &FileSystem{User: &model.User{}}

	// file not found
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnError(errors.New("error"))
		res, err := fs.GetThumb(context.Background(), 1)
		a.ErrorIs(err, ErrObjectNotExist)
		a.Nil(res)
		a.NoError(mock.ExpectationsWereMet())
	}

	// thumb not exist
	{
		fs.SetTargetFile(&[]model.File{{
			MetadataSerialized: map[string]string{
				model.ThumbStatusMetadataKey: model.ThumbStatusNotAvailable,
			},
			Policy: model.Policy{Type: "mock"},
		}})
		fs.FileTarget[0].Policy.ID = 1

		res, err := fs.GetThumb(context.Background(), 1)
		a.ErrorIs(err, ErrObjectNotExist)
		a.Nil(res)
	}

	// thumb not initialized, also failed to generate
	{
		fs.CleanTargets()
		fs.SetTargetFile(&[]model.File{{
			Policy: model.Policy{Type: "mock"},
			Size:   31457281,
		}})
		testHandller2 := new(FileHeaderMock)
		testHandller2.On("Thumb", testMock.Anything, &fs.FileTarget[0]).Return(&response.ContentResponse{}, driver.ErrorThumbNotExist)
		fs.Handler = testHandller2
		fs.FileTarget[0].Policy.ID = 1
		res, err := fs.GetThumb(context.Background(), 1)
		a.Contains(err.Error(), "file too large")
		a.Nil(res.Content)
	}

	// thumb not initialized, failed to get source
	{
		fs.CleanTargets()
		fs.SetTargetFile(&[]model.File{{
			Policy: model.Policy{Type: "mock"},
		}})
		testHandller2 := new(FileHeaderMock)
		testHandller2.On("Thumb", testMock.Anything, &fs.FileTarget[0]).Return(&response.ContentResponse{}, driver.ErrorThumbNotExist)
		testHandller2.On("Get", testMock.Anything, "").Return(MockRSC{}, errors.New("error"))
		fs.Handler = testHandller2
		fs.FileTarget[0].Policy.ID = 1
		res, err := fs.GetThumb(context.Background(), 1)
		a.Contains(err.Error(), "error")
		a.Nil(res.Content)
	}

	// thumb not initialized, no available generators
	{
		thumb.Generators = []thumb.Generator{}
		fs.CleanTargets()
		fs.SetTargetFile(&[]model.File{{
			Policy: model.Policy{Type: "local"},
		}})
		testHandller2 := new(FileHeaderMock)
		testHandller2.On("Thumb", testMock.Anything, &fs.FileTarget[0]).Return(&response.ContentResponse{}, driver.ErrorThumbNotExist)
		testHandller2.On("Get", testMock.Anything, "").Return(MockRSC{}, nil)
		fs.Handler = testHandller2
		fs.FileTarget[0].Policy.ID = 1
		res, err := fs.GetThumb(context.Background(), 1)
		a.ErrorIs(err, thumb.ErrNotAvailable)
		a.Nil(res)
	}

	// thumb not initialized, thumb generated but cannot be open
	{
		mockGenerator := &thumbmock.GeneratorMock{}
		thumb.Generators = []thumb.Generator{mockGenerator}
		fs.CleanTargets()
		fs.SetTargetFile(&[]model.File{{
			Policy: model.Policy{Type: "mock"},
		}})
		cache.Set("setting_thumb_vips_enabled", "1", 0)
		testHandller2 := new(FileHeaderMock)
		testHandller2.On("Thumb", testMock.Anything, &fs.FileTarget[0]).Return(&response.ContentResponse{}, driver.ErrorThumbNotExist)
		testHandller2.On("Get", testMock.Anything, "").Return(MockRSC{}, nil)
		mockGenerator.On("Generate", testMock.Anything, testMock.Anything, testMock.Anything, testMock.Anything, testMock.Anything).
			Return(&thumb.Result{Path: "not_exit_thumb"}, nil)

		fs.Handler = testHandller2
		fs.FileTarget[0].Policy.ID = 1
		res, err := fs.GetThumb(context.Background(), 1)
		a.Contains(err.Error(), "failed to open temp thumb")
		a.Nil(res.Content)
		testHandller2.AssertExpectations(t)
		mockGenerator.AssertExpectations(t)
	}
}

func TestFileSystem_ThumbWorker(t *testing.T) {
	asserts := assert.New(t)

	asserts.NotPanics(func() {
		getThumbWorker().addWorker()
		getThumbWorker().releaseWorker()
	})
}
