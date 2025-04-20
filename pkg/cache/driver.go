package cache

import (
	"encoding/gob"
)

func init() {
	gob.Register(map[string]itemWithTTL{})
}

// Driver 键值缓存存储容器
type Driver interface {
	// 设置值，ttl为过期时间，单位为秒
	Set(key string, value any, ttl int) error

	// 取值，并返回是否成功
	Get(key string) (any, bool)

	// 批量取值，返回成功取值的map即不存在的值
	Gets(keys []string, prefix string) (map[string]any, []string)

	// 批量设置值，所有的key都会加上prefix前缀
	Sets(values map[string]any, prefix string) error

	// Delete values by [Prefix + key]. If no ket is presented, all keys with given prefix will be deleted.
	Delete(prefix string, keys ...string) error

	// Save in-memory cache to disk
	Persist(path string) error

	// Restore cache from disk
	Restore(path string) error

	// Remove all entries
	DeleteAll() error
}
