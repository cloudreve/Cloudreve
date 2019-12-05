package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemoStore(t *testing.T) {
	asserts := assert.New(t)

	store := NewMemoStore()
	asserts.NotNil(store)
	asserts.NotNil(store.Store)
}

func TestMemoStore_Set(t *testing.T) {
	asserts := assert.New(t)

	store := NewMemoStore()
	err := store.Set("KEY", "vAL")
	asserts.NoError(err)

	val, ok := store.Store.Load("KEY")
	asserts.True(ok)
	asserts.Equal("vAL", val)
}

func TestMemoStore_Get(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	// 正常情况
	{
		_ = store.Set("string", "string_val")
		val, ok := store.Get("string")
		asserts.Equal("string_val", val)
		asserts.True(ok)
	}

	// Key不存在
	{
		val, ok := store.Get("something")
		asserts.Equal(nil, val)
		asserts.False(ok)
	}

	// 存储struct
	{
		type testStruct struct {
			key int
		}
		test := testStruct{key: 233}
		_ = store.Set("struct", test)
		val, ok := store.Get("struct")
		asserts.True(ok)
		res, ok := val.(testStruct)
		asserts.True(ok)
		asserts.Equal(test, res)
	}

}
