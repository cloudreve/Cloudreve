package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSlaveNode_InitAndKill(t *testing.T) {
	a := assert.New(t)
	n := &SlaveNode{
		callback: func(b bool, u uint) {

		},
	}

	a.NotPanics(func() {
		n.Init(&model.Node{})
		time.Sleep(time.Millisecond * 500)
		n.Init(&model.Node{})
		n.Kill()
	})
}

func TestSlaveNode_DummyMethods(t *testing.T) {
	a := assert.New(t)
	m := &SlaveNode{
		Model: &model.Node{},
	}

	m.Model.ID = 5
	a.Equal(m.Model.ID, m.ID())
	a.Equal(m.Model.ID, m.DBModel().ID)

	a.False(m.IsActive())
	a.False(m.IsMater())

	m.SubscribeStatusChange(func(isActive bool, id uint) {})
}

func TestSlaveNode_IsFeatureEnabled(t *testing.T) {
	a := assert.New(t)
	m := &SlaveNode{
		Model: &model.Node{},
	}

	a.False(m.IsFeatureEnabled("aria2"))
	a.False(m.IsFeatureEnabled("random"))
	m.Model.Aria2Enabled = true
	a.True(m.IsFeatureEnabled("aria2"))
}

func TestSlaveNode_Ping(t *testing.T) {
	a := assert.New(t)
	m := &SlaveNode{
		Model: &model.Node{},
	}

	// master return error code
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "POST", "heartbeat", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.Ping(&serializer.NodePingReq{})
		a.Error(err)
		a.Nil(res)
		a.Equal(1, err.(serializer.AppError).Code)
	}

	// return unexpected json
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "POST", "heartbeat", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"233\"}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.Ping(&serializer.NodePingReq{})
		a.Error(err)
		a.Nil(res)
	}

	// return success
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "POST", "heartbeat", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"{}\"}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.Ping(&serializer.NodePingReq{})
		a.NoError(err)
		a.NotNil(res)
	}
}

func TestSlaveNode_GetAria2Instance(t *testing.T) {
	a := assert.New(t)
	m := &SlaveNode{
		Model: &model.Node{},
	}

	a.NotNil(m.GetAria2Instance())
	m.Model.Aria2Enabled = true
	a.NotNil(m.GetAria2Instance())
	a.NotNil(m.GetAria2Instance())
}

func TestSlaveNode_StartPingLoop(t *testing.T) {
	callbackCount := 0
	finishedChan := make(chan struct{})
	mockRequest := requestMock{}
	mockRequest.On("Request", "POST", "heartbeat", testMock.Anything, testMock.Anything).Return(&request.Response{
		Response: &http.Response{
			StatusCode: 404,
		},
	})
	m := &SlaveNode{
		Active: true,
		Model:  &model.Node{},
		callback: func(b bool, u uint) {
			callbackCount++
			if callbackCount == 2 {
				close(finishedChan)
			}
			if callbackCount == 1 {
				mockRequest.AssertExpectations(t)
				mockRequest = requestMock{}
				mockRequest.On("Request", "POST", "heartbeat", testMock.Anything, testMock.Anything).Return(&request.Response{
					Response: &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"{}\"}")),
					},
				})
			}
		},
	}
	cache.Set("setting_slave_ping_interval", "0", 0)
	cache.Set("setting_slave_recover_interval", "0", 0)
	cache.Set("setting_slave_node_retry", "1", 0)

	m.caller.Client = &mockRequest
	go func() {
		select {
		case <-finishedChan:
			m.Kill()
		}
	}()
	m.StartPingLoop()
	mockRequest.AssertExpectations(t)
}

func TestSlaveNode_AuthInstance(t *testing.T) {
	a := assert.New(t)
	m := &SlaveNode{
		Model: &model.Node{},
	}

	a.NotNil(m.MasterAuthInstance())
	a.NotNil(m.SlaveAuthInstance())
}

func TestSlaveNode_ChangeStatus(t *testing.T) {
	a := assert.New(t)
	isActive := false
	m := &SlaveNode{
		Model: &model.Node{},
		callback: func(b bool, u uint) {
			isActive = b
		},
	}

	a.NotPanics(func() {
		m.changeStatus(false)
	})
	m.changeStatus(true)
	a.True(isActive)
}

func getTestRPCNodeSlave() *SlaveNode {
	m := &SlaveNode{
		Model: &model.Node{},
	}
	m.caller.parent = m
	return m
}

func TestSlaveCaller_CreateTask(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNodeSlave()

	// master return 404
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/task", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.CreateTask(&model.Download{}, nil)
		a.Empty(res)
		a.Error(err)
	}

	// master return error
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/task", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.CreateTask(&model.Download{}, nil)
		a.Empty(res)
		a.Error(err)
	}

	// master return success
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/task", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"res\"}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.CreateTask(&model.Download{}, nil)
		a.Equal("res", res)
		a.NoError(err)
	}
}

func TestSlaveCaller_Status(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNodeSlave()

	// master return 404
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/status", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.Status(&model.Download{})
		a.Empty(res.Status)
		a.Error(err)
	}

	// master return error
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/status", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.Status(&model.Download{})
		a.Empty(res.Status)
		a.Error(err)
	}

	// master return success
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/status", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"re456456s\"}")),
			},
		})
		m.caller.Client = mockRequest
		res, err := m.caller.Status(&model.Download{})
		a.Empty(res.Status)
		a.NoError(err)
	}
}

func TestSlaveCaller_Cancel(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNodeSlave()

	// master return 404
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/cancel", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Cancel(&model.Download{})
		a.Error(err)
	}

	// master return error
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/cancel", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Cancel(&model.Download{})
		a.Error(err)
	}

	// master return success
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/cancel", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"res\"}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Cancel(&model.Download{})
		a.NoError(err)
	}
}

func TestSlaveCaller_Select(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNodeSlave()
	m.caller.Init()
	m.caller.GetConfig()

	// master return 404
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/select", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Select(&model.Download{}, nil)
		a.Error(err)
	}

	// master return error
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/select", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Select(&model.Download{}, nil)
		a.Error(err)
	}

	// master return success
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/select", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"res\"}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.Select(&model.Download{}, nil)
		a.NoError(err)
	}
}

func TestSlaveCaller_DeleteTempFile(t *testing.T) {
	a := assert.New(t)
	m := getTestRPCNodeSlave()
	m.caller.Init()
	m.caller.GetConfig()

	// master return 404
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/delete", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 404,
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.DeleteTempFile(&model.Download{})
		a.Error(err)
	}

	// master return error
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/delete", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.DeleteTempFile(&model.Download{})
		a.Error(err)
	}

	// master return success
	{
		mockRequest := requestMock{}
		mockRequest.On("Request", "POST", "aria2/delete", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"res\"}")),
			},
		})
		m.caller.Client = mockRequest
		err := m.caller.DeleteTempFile(&model.Download{})
		a.NoError(err)
	}
}
