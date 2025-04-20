package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSet(t *testing.T) {
	asserts := assert.New(t)

	asserts.NoError(Set("123", "321", -1))
}

func TestGet(t *testing.T) {
	asserts := assert.New(t)
	asserts.NoError(Set("123", "321", -1))

	value, ok := Get("123")
	asserts.True(ok)
	asserts.Equal("321", value)

	value, ok = Get("not_exist")
	asserts.False(ok)
}

func TestDeletes(t *testing.T) {
	asserts := assert.New(t)
	asserts.NoError(Set("123", "321", -1))
	err := Deletes([]string{"123"}, "")
	asserts.NoError(err)
	_, exist := Get("123")
	asserts.False(exist)
}

func TestGetSettings(t *testing.T) {
	asserts := assert.New(t)
	asserts.NoError(Set("test_1", "1", -1))

	values, missed := GetSettings([]string{"1", "2"}, "test_")
	asserts.Equal(map[string]string{"1": "1"}, values)
	asserts.Equal([]string{"2"}, missed)
}

func TestSetSettings(t *testing.T) {
	asserts := assert.New(t)

	err := SetSettings(map[string]string{"3": "3", "4": "4"}, "test_")
	asserts.NoError(err)
	value1, _ := Get("test_3")
	value2, _ := Get("test_4")
	asserts.Equal("3", value1)
	asserts.Equal("4", value2)
}

func TestInit(t *testing.T) {
	asserts := assert.New(t)

	asserts.NotPanics(func() {
		Init()
	})
}
