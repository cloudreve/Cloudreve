package model

import (
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	asserts := assert.New(t)
	conf.DatabaseConfig.Type = "sqlite"
	DB, _ = gorm.Open("sqlite", ":memory:")

	asserts.NotPanics(func() {
		migration()
	})
	conf.DatabaseConfig.Type = "mysql"
	DB = mockDB
}
