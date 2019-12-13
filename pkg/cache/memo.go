package cache

import "sync"

// MemoStore 内存存储驱动
type MemoStore struct {
	Store *sync.Map
}

// NewMemoStore 新建内存存储
func NewMemoStore() *MemoStore {
	return &MemoStore{
		Store: &sync.Map{},
	}
}

// Set 存储值
func (store *MemoStore) Set(key string, value interface{}, ttl int) error {
	store.Store.Store(key, value)
	return nil
}

// Get 取值
func (store *MemoStore) Get(key string) (interface{}, bool) {
	return store.Store.Load(key)
}

// Gets 批量取值
func (store *MemoStore) Gets(keys []string, prefix string) (map[string]interface{}, []string) {
	var res = make(map[string]interface{})
	var notFound = make([]string, 0, len(keys))

	for _, key := range keys {
		if value, ok := store.Store.Load(prefix + key); ok {
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
