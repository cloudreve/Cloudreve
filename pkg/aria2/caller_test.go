package aria2

import (
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestRPCService_Init(t *testing.T) {
	asserts := assert.New(t)
	caller := &RPCService{}
	asserts.Error(caller.Init("ws://", "", 1, nil))
	asserts.NoError(caller.Init("http://127.0.0.1", "", 1, nil))
}

func TestRPCService_Status(t *testing.T) {
	asserts := assert.New(t)
	caller := &RPCService{}
	asserts.NoError(caller.Init("http://127.0.0.1", "", 1, nil))

	_, err := caller.Status(&model.Download{})
	asserts.Error(err)
}

func TestRPCService_Cancel(t *testing.T) {
	asserts := assert.New(t)
	caller := &RPCService{}
	asserts.NoError(caller.Init("http://127.0.0.1", "", 1, nil))

	err := caller.Cancel(&model.Download{Parent: "test"})
	asserts.Error(err)
}

func TestRPCService_Select(t *testing.T) {
	asserts := assert.New(t)
	caller := &RPCService{}
	asserts.NoError(caller.Init("http://127.0.0.1", "", 1, nil))

	err := caller.Select(&model.Download{Parent: "test"}, []int{1, 2, 3})
	asserts.Error(err)
}

func TestRPCService_CreateTask(t *testing.T) {
	asserts := assert.New(t)
	caller := &RPCService{}
	asserts.NoError(caller.Init("http://127.0.0.1", "", 1, nil))
	cache.Set("setting_aria2_temp_path", "test", 0)
	err := caller.CreateTask(&model.Download{Parent: "test"}, map[string]interface{}{"1": "1"})
	asserts.Error(err)
}
