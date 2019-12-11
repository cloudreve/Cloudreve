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

func TestMemoStore_Gets(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	err := store.Set("1", "1,val")
	err = store.Set("2", "2,val")
	err = store.Set("3", "3,val")
	err = store.Set("4", "4,val")
	asserts.NoError(err)

	// 全部命中
	{
		values, miss := store.Gets([]string{"1", "2", "3", "4"}, "")
		asserts.Len(values, 4)
		asserts.Len(miss, 0)
	}

	// 命中一半
	{
		values, miss := store.Gets([]string{"1", "2", "9", "10"}, "")
		asserts.Len(values, 2)
		asserts.Equal([]string{"9", "10"}, miss)
	}
}

func TestMemoStore_Sets(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	err := store.Sets(map[string]interface{}{
		"1": "1.val",
		"2": "2.val",
		"3": "3.val",
		"4": "4.val",
	}, "test_")
	asserts.NoError(err)

	vals, miss := store.Gets([]string{"1", "2", "3", "4"}, "test_")
	asserts.Len(miss, 0)
	asserts.Equal(map[string]interface{}{
		"1": "1.val",
		"2": "2.val",
		"3": "3.val",
		"4": "4.val",
	}, vals)
}

func TestMemoStore_Delete(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	err := store.Sets(map[string]interface{}{
		"1": "1.val",
		"2": "2.val",
		"3": "3.val",
		"4": "4.val",
	}, "test_")
	asserts.NoError(err)

	err = store.Delete([]string{"1", "2"}, "test_")
	asserts.NoError(err)
	values, miss := store.Gets([]string{"1", "2", "3", "4"}, "test_")
	asserts.Equal([]string{"1", "2"}, miss)
	asserts.Equal(map[string]interface{}{"3": "3.val", "4": "4.val"}, values)
}
