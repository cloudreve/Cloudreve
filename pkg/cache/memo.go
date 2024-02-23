package cache

import (
	"encoding/gob"
	"fmt"
	"os"
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
	Expires int64
	Value   interface{}
}

const DefaultCacheFile = "cache_persist.bin"

func newItem(value interface{}, expires int) itemWithTTL {
	expires64 := int64(expires)
	if expires > 0 {
		expires64 = time.Now().Unix() + expires64
	}
	return itemWithTTL{
		Value:   value,
		Expires: expires64,
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

	if itemObj.Expires > 0 && itemObj.Expires < time.Now().Unix() {
		return nil, false
	}

	return itemObj.Value, ok

}

// GarbageCollect 回收已过期的缓存
func (store *MemoStore) GarbageCollect() {
	store.Store.Range(func(key, value interface{}) bool {
		if item, ok := value.(itemWithTTL); ok {
			if item.Expires > 0 && item.Expires < time.Now().Unix() {
				util.Log().Debug("Cache %q is garbage collected.", key.(string))
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
		store.Store.Store(prefix+key, newItem(value, 0))
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

// Persist write memory store into cache
func (store *MemoStore) Persist(path string) error {
	persisted := make(map[string]itemWithTTL)
	store.Store.Range(func(key, value interface{}) bool {
		v, ok := store.Store.Load(key)
		if _, ok := getValue(v, ok); ok {
			persisted[key.(string)] = v.(itemWithTTL)
		}

		return true
	})

	res, err := serializer(persisted)
	if err != nil {
		return fmt.Errorf("failed to serialize cache: %s", err)
	}

	// err = os.WriteFile(path, res, 0644)
	file, err := util.CreatNestedFile(path)
	if err == nil {
		_, err = file.Write(res)
		file.Chmod(0644)
		file.Close()
	}
	return err
}

// Restore memory cache from disk file
func (store *MemoStore) Restore(path string) error {
	if !util.Exists(path) {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %s", err)
	}

	defer func() {
		f.Close()
		os.Remove(path)
	}()

	persisted := &item{}
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&persisted); err != nil {
		return fmt.Errorf("unknown cache file format: %s", err)
	}

	items := persisted.Value.(map[string]itemWithTTL)
	loaded := 0
	for k, v := range items {
		if _, ok := getValue(v, true); ok {
			loaded++
			store.Store.Store(k, v)
		} else {
			util.Log().Debug("Persisted cache %q is expired.", k)
		}
	}

	util.Log().Info("Restored %d items from %q into memory cache.", loaded, path)
	return nil
}
