package model

import (
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigration(t *testing.T) {
	asserts := assert.New(t)
	DB, _ = gorm.Open("sqlite3", ":memory:")

	asserts.NotPanics(func() {
		migration()
	})
	DB = mockDB
}
