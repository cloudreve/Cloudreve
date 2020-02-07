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

func TestCreatNestedFile(t *testing.T) {
	asserts := assert.New(t)

	// 父目录不存在
	{
		file, err := CreatNestedFile("test/nest.txt")
		asserts.NoError(err)
		asserts.NoError(file.Close())
		asserts.FileExists("test/nest.txt")
	}

	// 父目录存在
	{
		file, err := CreatNestedFile("test/direct.txt")
		asserts.NoError(err)
		asserts.NoError(file.Close())
		asserts.FileExists("test/direct.txt")
	}
}

func TestIsEmpty(t *testing.T) {
	asserts := assert.New(t)

	asserts.False(IsEmpty(""))
	asserts.False(IsEmpty("not_exist"))
}
