package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExists(t *testing.T) {
	asserts := assert.New(t)
	asserts.True(Exists("io_test.go"))
	asserts.False(Exists("io_test.js"))
}
