package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDotPathToStandardPath(t *testing.T) {
	asserts := assert.New(t)

	asserts.Equal("/", DotPathToStandardPath(""))
	asserts.Equal("/目录", DotPathToStandardPath("目录"))
	asserts.Equal("/目录/目录2", DotPathToStandardPath("目录,目录2"))
}

func TestFillSlash(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal("/", FillSlash("/"))
	asserts.Equal("/", FillSlash(""))
	asserts.Equal("/123/", FillSlash("/123"))
}

func TestRemoveSlash(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal("/", RemoveSlash("/"))
	asserts.Equal("/123/1236", RemoveSlash("/123/1236"))
	asserts.Equal("/123/1236", RemoveSlash("/123/1236/"))
}

func TestSplitPath(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal([]string{}, SplitPath(""))
	asserts.Equal([]string{}, SplitPath("1"))
	asserts.Equal([]string{"/"}, SplitPath("/"))
	asserts.Equal([]string{"/", "123", "321"}, SplitPath("/123/321"))
}
