package middleware

import (
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"net/http/httptest"
	"testing"
)

func TestMasterMetadata(t *testing.T) {
	a := assert.New(t)
	masterMetaDataFunc := MasterMetadata()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)

	c.Request.Header = map[string][]string{
		"X-Site-Id":           {"expectedSiteID"},
		"X-Site-Url":          {"expectedSiteURL"},
		"X-Cloudreve-Version": {"expectedMasterVersion"},
	}
	masterMetaDataFunc(c)
	siteID, _ := c.Get("MasterSiteID")
	siteURL, _ := c.Get("MasterSiteURL")
	siteVersion, _ := c.Get("MasterVersion")

	a.Equal("expectedSiteID", siteID.(string))
	a.Equal("expectedSiteURL", siteURL.(string))
	a.Equal("expectedMasterVersion", siteVersion.(string))
}

func TestSlaveRPCSignRequired(t *testing.T) {
	a := assert.New(t)
	np := &cluster.NodePool{}
	np.Init()
	slaveRPCSignRequiredFunc := SlaveRPCSignRequired(np)
	rec := httptest.NewRecorder()

	// id parse failed
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Node-Id", "unknown")
		slaveRPCSignRequiredFunc(c)
		a.True(c.IsAborted())
	}

	// node id not exist
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Node-Id", "38")
		slaveRPCSignRequiredFunc(c)
		a.True(c.IsAborted())
	}

	// success
	{
		authInstance := auth.HMACAuth{SecretKey: []byte("")}
		np.Add(&model.Node{Model: gorm.Model{
			ID: 38,
		}})

		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Request.Header.Set("X-Node-Id", "38")
		c.Request = auth.SignRequest(authInstance, c.Request, 0)
		slaveRPCSignRequiredFunc(c)
		a.False(c.IsAborted())
	}
}

type SlaveControllerMock struct {
	testMock.Mock
}

func (s SlaveControllerMock) HandleHeartBeat(pingReq *serializer.NodePingReq) (serializer.NodePingResp, error) {
	args := s.Called(pingReq)
	return args.Get(0).(serializer.NodePingResp), args.Error(1)
}

func (s SlaveControllerMock) GetAria2Instance(s2 string) (common.Aria2, error) {
	args := s.Called(s2)
	return args.Get(0).(common.Aria2), args.Error(1)
}

func (s SlaveControllerMock) SendNotification(s3 string, s2 string, message mq.Message) error {
	args := s.Called(s3, s2, message)
	return args.Error(0)
}

func (s SlaveControllerMock) SubmitTask(s3 string, i interface{}, s2 string, f func(interface{})) error {
	args := s.Called(s3, i, s2, f)
	return args.Error(0)
}

func (s SlaveControllerMock) GetMasterInfo(s2 string) (*cluster.MasterInfo, error) {
	args := s.Called(s2)
	return args.Get(0).(*cluster.MasterInfo), args.Error(1)
}

func (s SlaveControllerMock) GetOneDriveToken(s2 string, u uint) (string, error) {
	args := s.Called(s2, u)
	return args.String(0), args.Error(1)
}

func TestUseSlaveAria2Instance(t *testing.T) {
	a := assert.New(t)

	// MasterSiteID not set
	{
		testController := &SlaveControllerMock{}
		useSlaveAria2InstanceFunc := UseSlaveAria2Instance(testController)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		useSlaveAria2InstanceFunc(c)
		a.True(c.IsAborted())
	}

	// Cannot get aria2 instances
	{
		testController := &SlaveControllerMock{}
		useSlaveAria2InstanceFunc := UseSlaveAria2Instance(testController)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("MasterSiteID", "expectedSiteID")
		testController.On("GetAria2Instance", "expectedSiteID").Return(&common.DummyAria2{}, errors.New("error"))
		useSlaveAria2InstanceFunc(c)
		a.True(c.IsAborted())
		testController.AssertExpectations(t)
	}

	// Success
	{
		testController := &SlaveControllerMock{}
		useSlaveAria2InstanceFunc := UseSlaveAria2Instance(testController)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("MasterSiteID", "expectedSiteID")
		testController.On("GetAria2Instance", "expectedSiteID").Return(&common.DummyAria2{}, nil)
		useSlaveAria2InstanceFunc(c)
		a.False(c.IsAborted())
		res, _ := c.Get("MasterAria2Instance")
		a.NotNil(res)
		testController.AssertExpectations(t)
	}
}
