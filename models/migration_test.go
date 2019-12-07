package model

import (
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigration(t *testing.T) {
	asserts := assert.New(t)
	conf.DatabaseConfig.Type = "sqlite3"
	DB, _ = gorm.Open("sqlite3", ":memory:")

	asserts.NotPanics(func() {
		migration()
	})
	conf.DatabaseConfig.Type = "mysql"
	DB = mockDB
}
