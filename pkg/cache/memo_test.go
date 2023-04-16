package cache

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
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
	err := store.Set("KEY", "vAL", -1)
	asserts.NoError(err)

	val, ok := store.Store.Load("KEY")
	asserts.True(ok)
	asserts.Equal("vAL", val.(itemWithTTL).Value)
}

func TestMemoStore_Get(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	// 正常情况
	{
		_ = store.Set("string", "string_val", -1)
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
		_ = store.Set("struct", test, -1)
		val, ok := store.Get("struct")
		asserts.True(ok)
		res, ok := val.(testStruct)
		asserts.True(ok)
		asserts.Equal(test, res)
	}

	// 过期
	{
		_ = store.Set("string", "string_val", 1)
		time.Sleep(time.Duration(2) * time.Second)
		val, ok := store.Get("string")
		asserts.Nil(val)
		asserts.False(ok)
	}

}

func TestMemoStore_Gets(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()

	err := store.Set("1", "1,val", -1)
	err = store.Set("2", "2,val", -1)
	err = store.Set("3", "3,val", -1)
	err = store.Set("4", "4,val", -1)
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

func TestMemoStore_GarbageCollect(t *testing.T) {
	asserts := assert.New(t)
	store := NewMemoStore()
	store.Set("test", 1, 1)
	time.Sleep(time.Duration(2000) * time.Millisecond)
	store.GarbageCollect()
	_, ok := store.Get("test")
	asserts.False(ok)
}

func TestMemoStore_PersistFailed(t *testing.T) {
	a := assert.New(t)
	store := NewMemoStore()
	type testStruct struct{ v string }
	store.Set("test", 1, 0)
	store.Set("test2", testStruct{v: "test"}, 0)
	err := store.Persist(filepath.Join(t.TempDir(), "TestMemoStore_PersistFailed"))
	a.Error(err)
}

func TestMemoStore_PersistAndRestore(t *testing.T) {
	a := assert.New(t)
	store := NewMemoStore()
	store.Set("test", 1, 0)
	// already expired
	store.Store.Store("test2", itemWithTTL{Value: "test", Expires: 1})
	// expired after persist
	store.Set("test3", 1, 1)
	temp := filepath.Join(t.TempDir(), "TestMemoStore_PersistFailed")

	// Persist
	err := store.Persist(temp)
	a.NoError(err)
	a.FileExists(temp)

	time.Sleep(2 * time.Second)
	// Restore
	store2 := NewMemoStore()
	err = store2.Restore(temp)
	a.NoError(err)
	test, testOk := store2.Get("test")
	a.EqualValues(1, test)
	a.True(testOk)
	test2, test2Ok := store2.Get("test2")
	a.Nil(test2)
	a.False(test2Ok)
	test3, test3Ok := store2.Get("test3")
	a.Nil(test3)
	a.False(test3Ok)

	a.NoFileExists(temp)
}
