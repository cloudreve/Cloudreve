package middleware

import (
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/mocks/controllermock"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
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
		"X-Cr-Site-Id":           {"expectedSiteID"},
		"X-Cr-Site-Url":          {"expectedSiteURL"},
		"X-Cr-Cloudreve-Version": {"expectedMasterVersion"},
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
		c.Request.Header.Set("X-Cr-Node-Id", "unknown")
		slaveRPCSignRequiredFunc(c)
		a.True(c.IsAborted())
	}

	// node id not exist
	{
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Cr-Node-Id", "38")
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
		c.Request.Header.Set("X-Cr-Node-Id", "38")
		c.Request = auth.SignRequest(authInstance, c.Request, 0)
		slaveRPCSignRequiredFunc(c)
		a.False(c.IsAborted())
	}
}

func TestUseSlaveAria2Instance(t *testing.T) {
	a := assert.New(t)

	// MasterSiteID not set
	{
		testController := &controllermock.SlaveControllerMock{}
		useSlaveAria2InstanceFunc := UseSlaveAria2Instance(testController)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		useSlaveAria2InstanceFunc(c)
		a.True(c.IsAborted())
	}

	// Cannot get aria2 instances
	{
		testController := &controllermock.SlaveControllerMock{}
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
		testController := &controllermock.SlaveControllerMock{}
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
