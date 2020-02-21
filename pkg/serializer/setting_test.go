package serializer

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckSettingValue(t *testing.T) {
	asserts := assert.New(t)

	asserts.Equal("", checkSettingValue(map[string]string{}, "key"))
	asserts.Equal("123", checkSettingValue(map[string]string{"key": "123"}, "key"))
}

func TestBuildSiteConfig(t *testing.T) {
	asserts := assert.New(t)

	res := BuildSiteConfig(map[string]string{"not exist": ""}, &model.User{})
	asserts.Equal("", res.Data.(SiteConfig).SiteName)

	res = BuildSiteConfig(map[string]string{"siteName": "123"}, &model.User{})
	asserts.Equal("123", res.Data.(SiteConfig).SiteName)

	res = BuildSiteConfig(map[string]string{"qq_login": "1"}, &model.User{})
	asserts.Equal(true, res.Data.(SiteConfig).QQLogin)
	asserts.Equal(uint(0), res.Data.(SiteConfig).User.ID)

	// 非空用户
	res = BuildSiteConfig(map[string]string{"qq_login": "1"}, &model.User{
		Model: gorm.Model{
			ID: 5,
		},
	})
	asserts.Equal(uint(5), res.Data.(SiteConfig).User.ID)
}

func TestBuildTaskList(t *testing.T) {
	asserts := assert.New(t)
	tasks := []model.Task{{}}

	res := BuildTaskList(tasks, 1)
	asserts.NotNil(res)
}
