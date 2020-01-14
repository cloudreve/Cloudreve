package serializer

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildUser(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{MaxSize: 1024 * 1024},
	}
	res := BuildUser(user)
	asserts.Equal("1.00mb", res.Policy.MaxSize)

}

func TestBuildUserResponse(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{MaxSize: 1024 * 1024},
	}
	res := BuildUserResponse(user)
	asserts.Equal("1.00mb", res.Data.(User).Policy.MaxSize)
}

func TestBuildUserStorageResponse(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("pack_size_0", uint64(0), 0)

	{
		user := model.User{
			Storage: 0,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(0), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(10), res.Data.(storage).Free)
	}
	{
		user := model.User{
			Storage: 6,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(6), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(4), res.Data.(storage).Free)
	}
	{
		user := model.User{
			Storage: 20,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(20), res.Data.(storage).Used)
		asserts.Equal(uint64(10), res.Data.(storage).Total)
		asserts.Equal(uint64(0), res.Data.(storage).Free)
	}
	{
		cache.Set("pack_size_0", uint64(1), 0)
		user := model.User{
			Storage: 6,
			Group:   model.Group{MaxStorage: 10},
		}
		res := BuildUserStorageResponse(user)
		asserts.Equal(uint64(6), res.Data.(storage).Used)
		asserts.Equal(uint64(11), res.Data.(storage).Total)
		asserts.Equal(uint64(5), res.Data.(storage).Free)
	}
}
