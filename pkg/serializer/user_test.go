package serializer

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
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

func TestBuildUser(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{MaxSize: 1024 * 1024},
	}
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	res := BuildUser(user)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Equal("1.00mb", res.Policy.MaxSize)

}

func TestBuildUserResponse(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{MaxSize: 1024 * 1024},
	}
	res := BuildUserResponse(user)
	asserts.Equal("1.00mb", res.Data.(User).Policy.MaxSize)
}

func TestBuildUserStorageResponse(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("pack_size_0", uint64(0), 0)

	{
		user := model.User{
			Storage: 0,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(0), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(10), res.Data.(storage).Free)
	}
	{
		user := model.User{
			Storage: 6,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(6), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(4), res.Data.(storage).Free)
	}
	{
		user := model.User{
			Storage: 20,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(20), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(0), res.Data.(storage).Free)
	}
	{
		user := model.User{
			Storage: 6,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(6), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(4), res.Data.(storage).Free)
	}
}

func TestBuildTagRes(t *testing.T) {
	asserts := assert.New(t)
	tags := []model.Tag{
		{
			Type:       0,
			Expression: "exp",
		},
		{
			Type:       1,
			Expression: "exp",
		},
	}
	res := buildTagRes(tags)
	asserts.Len(res, 2)
	asserts.Equal("", res[0].Expression)
	asserts.Equal("exp", res[1].Expression)
}

func TestBuildWebAuthnList(t *testing.T) {
	asserts := assert.New(t)
	credentials := []webauthn.Credential{{}}
	res := BuildWebAuthnList(credentials)
	asserts.Len(res, 1)
}
