package cache

import (
	"sync"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// MemoStore 内存存储驱动
type MemoStore struct {
	Store *sync.Map
}

// item 存储的对象
type itemWithTTL struct {
	expires int64
	value   interface{}
}

func newItem(value interface{}, expires int) itemWithTTL {
	expires64 := int64(expires)
	if expires > 0 {
		expires64 = time.Now().Unix() + expires64
	}
	return itemWithTTL{
		value:   value,
		expires: expires64,
	}
}

// getValue 从itemWithTTL中取值
func getValue(item interface{}, ok bool) (interface{}, bool) {
	if !ok {
		return nil, ok
	}

	var itemObj itemWithTTL
	if itemObj, ok = item.(itemWithTTL); !ok {
		return item, true
	}

	if itemObj.expires > 0 && itemObj.expires < time.Now().Unix() {
		return nil, false
	}

	return itemObj.value, ok

}

// GarbageCollect 回收已过期的缓存
func (store *MemoStore) GarbageCollect() {
	store.Store.Range(func(key, value interface{}) bool {
		if item, ok := value.(itemWithTTL); ok {
			if item.expires > 0 && item.expires < time.Now().Unix() {
				util.Log().Debug("回收垃圾[%s]", key.(string))
				store.Store.Delete(key)
			}
		}
		return true
	})
}

// NewMemoStore 新建内存存储
func NewMemoStore() *MemoStore {
	return &MemoStore{
		Store: &sync.Map{},
	}
}

// Set 存储值
func (store *MemoStore) Set(key string, value interface{}, ttl int) error {
	store.Store.Store(key, newItem(value, ttl))
	return nil
}

// Get 取值
func (store *MemoStore) Get(key string) (interface{}, bool) {
	return getValue(store.Store.Load(key))
}

// Gets 批量取值
func (store *MemoStore) Gets(keys []string, prefix string) (map[string]interface{}, []string) {
	var res = make(map[string]interface{})
	var notFound = make([]string, 0, len(keys))

	for _, key := range keys {
		if value, ok := getValue(store.Store.Load(prefix + key)); ok {
			res[key] = value
		} else {
			notFound = append(notFound, key)
		}
	}

	return res, notFound
}

// Sets 批量设置值
func (store *MemoStore) Sets(values map[string]interface{}, prefix string) error {
	for key, value := range values {
		store.Store.Store(prefix+key, value)
	}
	return nil
}

// Delete 批量删除值
func (store *MemoStore) Delete(keys []string, prefix string) error {
	for _, key := range keys {
		store.Store.Delete(prefix + key)
	}
	return nil
}
