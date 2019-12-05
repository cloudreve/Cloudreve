package cache

// Store 缓存存储器
var Store Driver = NewMemoStore()

// Driver 键值缓存存储容器
type Driver interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool)
}
