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
func (store *MemoStore) Set(key string, value interface{}) error {
	store.Store.Store(key, value)
	return nil
}

// Get 取值
func (store *MemoStore) Get(key string) (interface{}, bool) {
	return store.Store.Load(key)
}
