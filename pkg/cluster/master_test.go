package cluster

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMasterNode_Init(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{}
	m.Init(&model.Node{Status: model.NodeSuspend})
	a.Equal(model.NodeSuspend, m.DBModel().Status)
	m.Init(&model.Node{Aria2Enabled: true})
}

func TestMasterNode_DummyMethods(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{
		Model: &model.Node{},
	}

	m.Model.ID = 5
	a.Equal(m.Model.ID, m.ID())

	res, err := m.Ping(&serializer.NodePingReq{})
	a.NoError(err)
	a.NotNil(res)

	a.True(m.IsActive())
	a.True(m.IsMater())

	m.SubscribeStatusChange(func(isActive bool, id uint) {})
}

func TestMasterNode_IsFeatureEnabled(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{
		Model: &model.Node{},
	}

	a.False(m.IsFeatureEnabled("aria2"))
	a.False(m.IsFeatureEnabled("random"))
	m.Model.Aria2Enabled = true
	a.True(m.IsFeatureEnabled("aria2"))
}

func TestMasterNode_AuthInstance(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{
		Model: &model.Node{},
	}

	a.NotNil(m.MasterAuthInstance())
	a.NotNil(m.SlaveAuthInstance())
}

func TestMasterNode_Kill(t *testing.T) {
	m := &MasterNode{
		Model: &model.Node{},
	}

	m.Kill()

	caller, _ := rpc.New(context.Background(), "http://", "", 0, nil)
	m.aria2RPC.Caller = caller
	m.Kill()
}

func TestMasterNode_GetAria2Instance(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{
		Model:    &model.Node{},
		aria2RPC: rpcService{},
	}

	m.aria2RPC.parent = m

	a.NotNil(m.GetAria2Instance())
	m.Model.Aria2Enabled = true
	a.NotNil(m.GetAria2Instance())
	m.aria2RPC.Initialized = true
	a.NotNil(m.GetAria2Instance())
}

func TestRpcService_Init(t *testing.T) {
	a := assert.New(t)
	m := &MasterNode{
		Model: &model.Node{
			Aria2OptionsSerialized: model.Aria2Option{
				Options: "{",
			},
		},
		aria2RPC: rpcService{},
	}
	m.aria2RPC.parent = m

	// failed to decode address
	{
		m.Model.Aria2OptionsSerialized.Server = string([]byte{0x7f})
		a.Error(m.aria2RPC.Init())
	}

	// failed to decode options
	{
		m.Model.Aria2OptionsSerialized.Server = ""
		a.Error(m.aria2RPC.Init())
	}

	// failed to initialized
	{
		m.Model.Aria2OptionsSerialized.Server = ""
		m.Model.Aria2OptionsSerialized.Options = "{}"
		caller, _ := rpc.New(context.Background(), "http://", "", 0, nil)
		m.aria2RPC.Caller = caller
		a.Error(m.aria2RPC.Init())
		a.False(m.aria2RPC.Initialized)
	}
}

func getTestRPCNode() *MasterNode {
	m := &MasterNode{
		Model: &model.Node{
			Aria2OptionsSerialized: model.Aria2Option{},
		},
		aria2RPC: rpcService{
			options: &clientOptions{
				Options: map[string]interface{}{"1": "1"},
			},
		},
	}
	m.aria2RPC.parent = m
	caller, _ := rpc.New(context.Background(), "http://", "", 0, nil)
	m.aria2RPC.Caller = caller
	return m
}

func TestRpcService_CreateTask(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNode()

	res, err := m.aria2RPC.CreateTask(&model.Download{}, map[string]interface{}{"1": "1"})
	a.Error(err)
	a.Empty(res)
}

func TestRpcService_Status(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNode()

	res, err := m.aria2RPC.Status(&model.Download{})
	a.Error(err)
	a.Empty(res)
}

func TestRpcService_Cancel(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNode()

	a.Error(m.aria2RPC.Cancel(&model.Download{}))
}

func TestRpcService_Select(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNode()

	a.NotNil(m.aria2RPC.GetConfig())
	a.Error(m.aria2RPC.Select(&model.Download{}, []int{1, 2, 3}))
}

func TestRpcService_DeleteTempFile(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNode()
	fdName := "TestRpcService_DeleteTempFile"
	a.NoError(os.Mkdir(fdName, 0644))

	a.NoError(m.aria2RPC.DeleteTempFile(&model.Download{Parent: fdName}))
	time.Sleep(500 * time.Millisecond)
	a.False(util.Exists(fdName))
}
