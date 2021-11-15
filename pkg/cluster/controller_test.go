package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestInitController(t *testing.T) {
	assert.NotPanics(t, func() {
		InitController()
	})
}

func TestSlaveController_HandleHeartBeat(t *testing.T) {
	a := assert.New(t)
	c := &slaveController{
		masters: make(map[string]MasterInfo),
	}

	// first heart beat
	{
		_, err := c.HandleHeartBeat(&serializer.NodePingReq{
			SiteID: "1",
			Node:   &model.Node{},
		})
		a.NoError(err)

		_, err = c.HandleHeartBeat(&serializer.NodePingReq{
			SiteID: "2",
			Node:   &model.Node{},
		})
		a.NoError(err)

		a.Len(c.masters, 2)
	}

	// second heart beat, no fresh
	{
		_, err := c.HandleHeartBeat(&serializer.NodePingReq{
			SiteID:  "1",
			SiteURL: "http://127.0.0.1",
			Node:    &model.Node{},
		})
		a.NoError(err)
		a.Len(c.masters, 2)
		a.Empty(c.masters["1"].URL)
	}

	// second heart beat, fresh
	{
		_, err := c.HandleHeartBeat(&serializer.NodePingReq{
			SiteID:   "1",
			IsUpdate: true,
			SiteURL:  "http://127.0.0.1",
			Node:     &model.Node{},
		})
		a.NoError(err)
		a.Len(c.masters, 2)
		a.Equal("http://127.0.0.1", c.masters["1"].URL.String())
	}

	// second heart beat, fresh, url illegal
	{
		_, err := c.HandleHeartBeat(&serializer.NodePingReq{
			SiteID:   "1",
			IsUpdate: true,
			SiteURL:  string([]byte{0x7f}),
			Node:     &model.Node{},
		})
		a.Error(err)
		a.Len(c.masters, 2)
		a.Equal("http://127.0.0.1", c.masters["1"].URL.String())
	}
}

type nodeMock struct {
	testMock.Mock
}

func (n nodeMock) Init(node *model.Node) {
	n.Called(node)
}

func (n nodeMock) IsFeatureEnabled(feature string) bool {
	args := n.Called(feature)
	return args.Bool(0)
}

func (n nodeMock) SubscribeStatusChange(callback func(isActive bool, id uint)) {
	n.Called(callback)
}

func (n nodeMock) Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error) {
	args := n.Called(req)
	return args.Get(0).(*serializer.NodePingResp), args.Error(1)
}

func (n nodeMock) IsActive() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n nodeMock) GetAria2Instance() common.Aria2 {
	args := n.Called()
	return args.Get(0).(common.Aria2)
}

func (n nodeMock) ID() uint {
	args := n.Called()
	return args.Get(0).(uint)
}

func (n nodeMock) Kill() {
	n.Called()
}

func (n nodeMock) IsMater() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n nodeMock) MasterAuthInstance() auth.Auth {
	args := n.Called()
	return args.Get(0).(auth.Auth)
}

func (n nodeMock) SlaveAuthInstance() auth.Auth {
	args := n.Called()
	return args.Get(0).(auth.Auth)
}

func (n nodeMock) DBModel() *model.Node {
	args := n.Called()
	return args.Get(0).(*model.Node)
}

func TestSlaveController_GetAria2Instance(t *testing.T) {
	a := assert.New(t)
	mockNode := &nodeMock{}
	mockNode.On("GetAria2Instance").Return(&common.DummyAria2{})
	c := &slaveController{
		masters: map[string]MasterInfo{
			"1": {Instance: mockNode},
		},
	}

	// node node found
	{
		res, err := c.GetAria2Instance("2")
		a.Nil(res)
		a.Equal(ErrMasterNotFound, err)
	}

	// node found
	{
		res, err := c.GetAria2Instance("1")
		a.NotNil(res)
		a.NoError(err)
		mockNode.AssertExpectations(t)
	}

}

type requestMock struct {
	testMock.Mock
}

func (r requestMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	return r.Called(method, target, body, opts).Get(0).(*request.Response)
}

func TestSlaveController_SendNotification(t *testing.T) {
	a := assert.New(t)
	c := &slaveController{
		masters: map[string]MasterInfo{
			"1": {},
		},
	}

	// node not exit
	{
		a.Equal(ErrMasterNotFound, c.SendNotification("2", "", mq.Message{}))
	}

	// gob encode error
	{
		type randomType struct{}
		a.Error(c.SendNotification("1", "", mq.Message{
			Content: randomType{},
		}))
	}

	// return none 200
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "PUT", "/api/v3/slave/notification/s1", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{StatusCode: http.StatusConflict},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		a.Error(c.SendNotification("1", "s1", mq.Message{}))
		mockRequest.AssertExpectations(t)
	}

	// master return error
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "PUT", "/api/v3/slave/notification/s2", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		a.Equal(1, c.SendNotification("1", "s2", mq.Message{}).(serializer.AppError).Code)
		mockRequest.AssertExpectations(t)
	}

	// success
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "PUT", "/api/v3/slave/notification/s3", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":0}")),
			},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		a.NoError(c.SendNotification("1", "s3", mq.Message{}))
		mockRequest.AssertExpectations(t)
	}
}

func TestSlaveController_SubmitTask(t *testing.T) {
	a := assert.New(t)
	c := &slaveController{
		masters: map[string]MasterInfo{
			"1": {
				jobTracker: map[string]bool{},
			},
		},
	}

	// node not exit
	{
		a.Equal(ErrMasterNotFound, c.SubmitTask("2", "", "", nil))
	}

	// success
	{
		submitted := false
		a.NoError(c.SubmitTask("1", "", "hash", func(i interface{}) {
			submitted = true
		}))
		a.True(submitted)
	}

	// job already submitted
	{
		submitted := false
		a.NoError(c.SubmitTask("1", "", "hash", func(i interface{}) {
			submitted = true
		}))
		a.False(submitted)
	}
}

func TestSlaveController_GetMasterInfo(t *testing.T) {
	a := assert.New(t)
	c := &slaveController{
		masters: map[string]MasterInfo{
			"1": {},
		},
	}

	// node not exit
	{
		res, err := c.GetMasterInfo("2")
		a.Equal(ErrMasterNotFound, err)
		a.Nil(res)
	}

	// success
	{
		res, err := c.GetMasterInfo("1")
		a.NoError(err)
		a.NotNil(res)
	}
}

func TestSlaveController_GetOneDriveToken(t *testing.T) {
	a := assert.New(t)
	c := &slaveController{
		masters: map[string]MasterInfo{
			"1": {},
		},
	}

	// node not exit
	{
		res, err := c.GetOneDriveToken("2", 1)
		a.Equal(ErrMasterNotFound, err)
		a.Empty(res)
	}

	// return none 200
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "GET", "/api/v3/slave/credential/onedrive/1", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{StatusCode: http.StatusConflict},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		res, err := c.GetOneDriveToken("1", 1)
		a.Error(err)
		a.Empty(res)
		mockRequest.AssertExpectations(t)
	}

	// master return error
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "GET", "/api/v3/slave/credential/onedrive/1", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"code\":1}")),
			},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		res, err := c.GetOneDriveToken("1", 1)
		a.Equal(1, err.(serializer.AppError).Code)
		a.Empty(res)
		mockRequest.AssertExpectations(t)
	}

	// success
	{
		mockRequest := &requestMock{}
		mockRequest.On("Request", "GET", "/api/v3/slave/credential/onedrive/1", testMock.Anything, testMock.Anything).Return(&request.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{\"data\":\"expected\"}")),
			},
		})
		c := &slaveController{
			masters: map[string]MasterInfo{
				"1": {Client: mockRequest},
			},
		}
		res, err := c.GetOneDriveToken("1", 1)
		a.NoError(err)
		a.Equal("expected", res)
		mockRequest.AssertExpectations(t)
	}

}
