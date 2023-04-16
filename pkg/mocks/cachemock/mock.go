package cachemock

import "github.com/stretchr/testify/mock"

type CacheClientMock struct {
	mock.Mock
}

func (c CacheClientMock) Set(key string, value interface{}, ttl int) error {
	return c.Called(key, value, ttl).Error(0)
}

func (c CacheClientMock) Get(key string) (interface{}, bool) {
	args := c.Called(key)
	return args.Get(0), args.Bool(1)
}

func (c CacheClientMock) Gets(keys []string, prefix string) (map[string]interface{}, []string) {
	args := c.Called(keys, prefix)
	return args.Get(0).(map[string]interface{}), args.Get(1).([]string)
}

func (c CacheClientMock) Sets(values map[string]interface{}, prefix string) error {
	return c.Called(values).Error(0)
}

func (c CacheClientMock) Delete(keys []string, prefix string) error {
	return c.Called(keys, prefix).Error(0)
}

func (c CacheClientMock) Persist(path string) error {
	return c.Called(path).Error(0)
}

func (c CacheClientMock) Restore(path string) error {
	return c.Called(path).Error(0)
}
