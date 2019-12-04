package model

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigration(t *testing.T) {
	asserts := assert.New(t)
	gin.SetMode(gin.TestMode)

	asserts.NotPanics(func() {
		migration()
	})
}
