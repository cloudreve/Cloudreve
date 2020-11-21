package filesystem

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
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

func TestFileSystem_ValidateLegalName(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{}
	asserts.True(fs.ValidateLegalName(ctx, "1.txt"))
	asserts.True(fs.ValidateLegalName(ctx, "1-1.txt"))
	asserts.True(fs.ValidateLegalName(ctx, "1？1.txt"))
	asserts.False(fs.ValidateLegalName(ctx, "1:1.txt"))
	asserts.False(fs.ValidateLegalName(ctx, "../11.txt"))
	asserts.False(fs.ValidateLegalName(ctx, "/11.txt"))
	asserts.False(fs.ValidateLegalName(ctx, "\\11.txt"))
	asserts.False(fs.ValidateLegalName(ctx, ""))
	asserts.False(fs.ValidateLegalName(ctx, "1.tx t "))
	asserts.True(fs.ValidateLegalName(ctx, "1.tx t"))
}

func TestFileSystem_ValidateCapacity(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	cache.Set("pack_size_0", uint64(0), 0)
	fs := FileSystem{
		User: &model.User{
			Storage: 10,
			Group: model.Group{
				MaxStorage: 11,
			},
		},
	}

	asserts.True(fs.ValidateCapacity(ctx, 1))
	asserts.Equal(uint64(11), fs.User.Storage)

	fs.User.Storage = 5
	asserts.False(fs.ValidateCapacity(ctx, 10))
	asserts.Equal(uint64(5), fs.User.Storage)
}

func TestFileSystem_ValidateFileSize(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Policy: model.Policy{
				MaxSize: 10,
			},
		},
	}

	asserts.True(fs.ValidateFileSize(ctx, 5))
	asserts.True(fs.ValidateFileSize(ctx, 10))
	asserts.False(fs.ValidateFileSize(ctx, 11))

	// 无限制
	fs.User.Policy.MaxSize = 0
	asserts.True(fs.ValidateFileSize(ctx, 11))
}

func TestFileSystem_ValidateExtension(t *testing.T) {
	asserts := assert.New(t)
	ctx := context.Background()
	fs := FileSystem{
		User: &model.User{
			Policy: model.Policy{
				OptionsSerialized: model.PolicyOption{
					FileType: nil,
				},
			},
		},
	}

	asserts.True(fs.ValidateExtension(ctx, "1"))
	asserts.True(fs.ValidateExtension(ctx, "1.txt"))

	fs.User.Policy.OptionsSerialized.FileType = []string{}
	asserts.True(fs.ValidateExtension(ctx, "1"))
	asserts.True(fs.ValidateExtension(ctx, "1.txt"))

	fs.User.Policy.OptionsSerialized.FileType = []string{"txt", "jpg"}
	asserts.False(fs.ValidateExtension(ctx, "1"))
	asserts.False(fs.ValidateExtension(ctx, "1.jpg.png"))
	asserts.True(fs.ValidateExtension(ctx, "1.txt"))
	asserts.True(fs.ValidateExtension(ctx, "1.png.jpg"))
	asserts.True(fs.ValidateExtension(ctx, "1.png.jpG"))
	asserts.False(fs.ValidateExtension(ctx, "1.png"))
}
