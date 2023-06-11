package wopi

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/cachemock"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/requestmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"net/url"
	"testing"
)

var mock sqlmock.Sqlmock

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}
	model.DB, _ = gorm.Open("mysql", db)
	defer db.Close()
	m.Run()
}

func TestNewSession(t *testing.T) {
	a := assert.New(t)
	endpoint, _ := url.Parse("http://localhost:8001/hosting/discovery")
	client := &client{
		cache: cache.NewMemoStore(),
		config: config{
			discoveryEndpoint: endpoint,
		},
	}

	// Discovery failed
	{
		expectedErr := errors.New("error")
		mockHttp := &requestmock.RequestMock{}
		client.http = mockHttp
		mockHttp.On(
			"Request",
			"GET",
			endpoint.String(),
			testMock.Anything,
			testMock.Anything,
		).Return(&request.Response{
			Err: expectedErr,
		})
		res, err := client.NewSession(0, &model.File{}, ActionPreview)
		a.Nil(res)
		a.ErrorIs(err, expectedErr)
		mockHttp.AssertExpectations(t)
	}

	// not supported ext
	{
		client.discovery = &WopiDiscovery{}
		client.actions = make(map[string]map[string]Action)
		res, err := client.NewSession(0, &model.File{}, ActionPreview)
		a.Nil(res)
		a.ErrorIs(err, ErrActionNotSupported)
	}

	// preferred action not supported
	{
		client.discovery = &WopiDiscovery{}
		client.actions = map[string]map[string]Action{
			".doc": {},
		}
		res, err := client.NewSession(0, &model.File{Name: "1.doc"}, ActionPreview)
		a.Nil(res)
		a.ErrorIs(err, ErrActionNotSupported)
	}

	// src url cannot be parsed
	{
		client.discovery = &WopiDiscovery{}
		client.actions = map[string]map[string]Action{
			".doc": {
				string(ActionPreviewFallback): Action{
					Urlsrc: string([]byte{0x7f}),
				},
			},
		}
		res, err := client.NewSession(0, &model.File{Name: "1.doc"}, ActionEdit)
		a.Nil(res)
		a.ErrorContains(err, "invalid control character in URL")
	}

	// all pass - default placeholder
	{
		client.discovery = &WopiDiscovery{}
		client.actions = map[string]map[string]Action{
			".doc": {
				string(ActionPreviewFallback): Action{
					Urlsrc: "https://doc.com/doc",
				},
			},
		}
		res, err := client.NewSession(0, &model.File{Name: "1.doc"}, ActionEdit)
		a.NotNil(res)
		a.NoError(err)
		resUrl := res.ActionURL.String()
		a.Contains(resUrl, wopiSrcParamDefault)
	}

	// all pass - with placeholders
	{
		client.discovery = &WopiDiscovery{}
		client.actions = map[string]map[string]Action{
			".doc": {
				string(ActionPreviewFallback): Action{
					Urlsrc: "https://doc.com/doc?origin=preserved&<dc=DC_LLCC&><notsuported=DISABLE_ASYNC&><src=WOPI_SOURCE&>",
				},
			},
		}
		res, err := client.NewSession(0, &model.File{Name: "1.doc"}, ActionEdit)
		a.NotNil(res)
		a.NoError(err)
		resUrl := res.ActionURL.String()
		a.Contains(resUrl, "origin=preserved")
		a.Contains(resUrl, "dc=lng")
		a.Contains(resUrl, "src=")
		a.NotContains(resUrl, "notsuported")
	}

	// cache operation failed
	{
		mockCache := &cachemock.CacheClientMock{}
		expectedErr := errors.New("error")
		client.cache = mockCache
		client.discovery = &WopiDiscovery{}
		client.actions = map[string]map[string]Action{
			".doc": {
				string(ActionPreviewFallback): Action{
					Urlsrc: "https://doc.com/doc",
				},
			},
		}
		mockCache.On("Set", testMock.Anything, testMock.Anything, testMock.Anything).Return(expectedErr)
		res, err := client.NewSession(0, &model.File{Name: "1.doc"}, ActionEdit)
		a.Nil(res)
		a.ErrorIs(err, expectedErr)
	}
}

func TestInit(t *testing.T) {
	a := assert.New(t)

	// not enabled
	{
		a.Nil(Default)
		Default = &client{}
		Init()
		a.Nil(Default)
	}

	// throw error
	{
		a.Nil(Default)
		cache.Set("setting_wopi_enabled", "1", 0)
		cache.Set("setting_wopi_endpoint", string([]byte{0x7f}), 0)
		Init()
		a.Nil(Default)
	}

	// all pass
	{
		a.Nil(Default)
		cache.Set("setting_wopi_enabled", "1", 0)
		cache.Set("setting_wopi_endpoint", "", 0)
		Init()
		a.NotNil(Default)
	}
}
