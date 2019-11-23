package serializer

import (
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

	res := BuildSiteConfig(map[string]string{"not exist": ""})
	asserts.Equal("", res.Data.(SiteConfig).SiteName)

	res = BuildSiteConfig(map[string]string{"siteName": "123"})
	asserts.Equal("123", res.Data.(SiteConfig).SiteName)

	res = BuildSiteConfig(map[string]string{"qq_login": "1"})
	asserts.Equal(true, res.Data.(SiteConfig).QQLogin)
}
