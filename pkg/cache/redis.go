package cache

import (
	"bytes"
	"encoding/gob"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

// RedisStore redis存储驱动
type RedisStore struct {
	pool *redis.Pool
}

type item struct {
	Value interface{}
}

// NewRedisStore 创建新的redis存储
func NewRedisStore(size int, network, address, password, database string) *RedisStore {
	return &RedisStore{
		pool: &redis.Pool{
			MaxIdle:     size,
			IdleTimeout: 240 * time.Second,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			Dial: func() (redis.Conn, error) {
				db, err := strconv.Atoi(database)
				if err != nil {
					return nil, err
				}

				c, err := redis.Dial(
					network,
					address,
					redis.DialDatabase(db),
					redis.DialPassword(password),
				)
				if err != nil {
					util.Log().Warning("无法创建Redis连接：%s", err)
					return nil, err
				}
				return c, nil
			},
		},
	}
}

// Set 存储值
func (store *RedisStore) Set(key string, value interface{}) error {
	rc := store.pool.Get()
	defer rc.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	storeValue := item{
		Value: value,
	}
	err := enc.Encode(storeValue)
	if err != nil {
		return err
	}

	if rc.Err() == nil {
		_, err := rc.Do("SET", key, buffer.Bytes())
		if err != nil {
			return err
		}
		return nil
	}

	return rc.Err()
}

// Get 取值
func (store *RedisStore) Get(key string) (interface{}, bool) {
	rc := store.pool.Get()
	defer rc.Close()

	v, err := redis.Bytes(rc.Do("GET", key))
	if err != nil {
		return nil, false
	}

	var res item
	buffer := bytes.NewReader(v)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&res)
	if err != nil {
		return nil, false
	}

	return res.Value, true

}

// Gets 批量取值
func (store *RedisStore) Gets(keys []string, prefix string) (map[string]interface{}, []string) {
	rc := store.pool.Get()
	defer rc.Close()

	var queryKeys = make([]string, len(keys))
	for key, value := range keys {
		queryKeys[key] = prefix + value
	}

	v, err := redis.ByteSlices(rc.Do("MGET", redis.Args{}.AddFlat(queryKeys)...))
	if err != nil {
		return nil, keys
	}

	var res = make(map[string]interface{})
	var missed = make([]string, 0, len(keys))

	for key, value := range v {
		var decoded item
		buffer := bytes.NewReader(value)
		dec := gob.NewDecoder(buffer)
		err = dec.Decode(&decoded)
		if err != nil || decoded.Value == nil {
			missed = append(missed, keys[key])
		} else {
			res[keys[key]] = decoded.Value
		}
	}
	// 解码所得值
	return res, missed
}

// Sets 批量设置值
func (store *RedisStore) Sets(values map[string]interface{}, prefix string) error {
	rc := store.pool.Get()
	defer rc.Close()
	var setValues = make(map[string]interface{})

	// 编码待设置值
	for key, value := range values {
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		storeValue := item{
			Value: value,
		}
		err := enc.Encode(storeValue)
		if err != nil {
			return err
		}
		setValues[prefix+key] = buffer.Bytes()
	}

	if rc.Err() == nil {
		_, err := rc.Do("MSET", redis.Args{}.AddFlat(setValues)...)
		if err != nil {
			return err
		}
		return nil
	}

	return rc.Err()
}
