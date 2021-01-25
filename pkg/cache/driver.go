package cache

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/gin-gonic/gin"
)

// Store 缓存存储器
var Store Driver = NewMemoStore()

// Init 初始化缓存
func Init() {
	//Store = NewRedisStore(10, "tcp", "127.0.0.1:6379", "", "0")
	//return
	if conf.RedisConfig.Server != "" && gin.Mode() != gin.TestMode {
		Store = NewRedisStore(
			10,
			conf.RedisConfig.Network,
			conf.RedisConfig.Server,
			conf.RedisConfig.Password,
			conf.RedisConfig.DB,
		)
	}
}

// Driver 键值缓存存储容器
type Driver interface {
	// 设置值，ttl为过期时间，单位为秒
	Set(key string, value interface{}, ttl int) error

	// 取值，并返回是否成功
	Get(key string) (interface{}, bool)

	// 批量取值，返回成功取值的map即不存在的值
	Gets(keys []string, prefix string) (map[string]interface{}, []string)

	// 批量设置值，所有的key都会加上prefix前缀
	Sets(values map[string]interface{}, prefix string) error

	// 删除值
	Delete(keys []string, prefix string) error
}

// Set 设置缓存值
func Set(key string, value interface{}, ttl int) error {
	return Store.Set(key, value, ttl)
}

// Get 获取缓存值
func Get(key string) (interface{}, bool) {
	return Store.Get(key)
}

// Deletes 删除值
func Deletes(keys []string, prefix string) error {
	return Store.Delete(keys, prefix)
}

// GetSettings 根据名称批量获取设置项缓存
func GetSettings(keys []string, prefix string) (map[string]string, []string) {
	raw, miss := Store.Gets(keys, prefix)

	res := make(map[string]string, len(raw))
	for k, v := range raw {
		res[k] = v.(string)
	}

	return res, miss
}

// SetSettings 批量设置站点设置缓存
func SetSettings(values map[string]string, prefix string) error {
	var toBeSet = make(map[string]interface{}, len(values))
	for key, value := range values {
		toBeSet[key] = interface{}(value)
	}
	return Store.Sets(toBeSet, prefix)
}
