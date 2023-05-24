package model

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestGetPolicyByID(t *testing.T) {
	asserts := assert.New(t)

	cache.Deletes([]string{"22", "23"}, "policy_")
	// 缓存未命中
	{
		rows := sqlmock.NewRows([]string{"name", "type", "options"}).
			AddRow("默认存储策略", "local", "{\"od_redirect\":\"123\"}")
		mock.ExpectQuery("^SELECT(.+)").WillReturnRows(rows)
		policy, err := GetPolicyByID(uint(22))
		asserts.NoError(err)
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal("默认存储策略", policy.Name)
		asserts.Equal("123", policy.OptionsSerialized.OauthRedirect)

		rows = sqlmock.NewRows([]string{"name", "type", "options"})
		mock.ExpectQuery("^SELECT(.+)").WillReturnRows(rows)
		policy, err = GetPolicyByID(uint(23))
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}

	// 命中
	{
		policy, err := GetPolicyByID(uint(22))
		asserts.NoError(err)
		asserts.Equal("默认存储策略", policy.Name)
		asserts.Equal("123", policy.OptionsSerialized.OauthRedirect)

	}

}

func TestPolicy_BeforeSave(t *testing.T) {
	asserts := assert.New(t)

	testPolicy := Policy{
		OptionsSerialized: PolicyOption{
			OauthRedirect: "123",
		},
	}
	expected, _ := json.Marshal(testPolicy.OptionsSerialized)
	err := testPolicy.BeforeSave()
	asserts.NoError(err)
	asserts.Equal(string(expected), testPolicy.Options)

}

func TestPolicy_GeneratePath(t *testing.T) {
	asserts := assert.New(t)
	testPolicy := Policy{}

	testPolicy.DirNameRule = "{randomkey16}"
	asserts.Len(testPolicy.GeneratePath(1, "/"), 16)

	testPolicy.DirNameRule = "{randomkey8}"
	asserts.Len(testPolicy.GeneratePath(1, "/"), 8)

	testPolicy.DirNameRule = "{timestamp}"
	asserts.Equal(testPolicy.GeneratePath(1, "/"), strconv.FormatInt(time.Now().Unix(), 10))

	testPolicy.DirNameRule = "{uid}"
	asserts.Equal(testPolicy.GeneratePath(1, "/"), strconv.Itoa(int(1)))

	testPolicy.DirNameRule = "{datetime}"
	asserts.Len(testPolicy.GeneratePath(1, "/"), 14)

	testPolicy.DirNameRule = "{date}"
	asserts.Len(testPolicy.GeneratePath(1, "/"), 8)

	testPolicy.DirNameRule = "123{date}ss{datetime}"
	asserts.Len(testPolicy.GeneratePath(1, "/"), 27)

	testPolicy.DirNameRule = "/1/{path}/456"
	asserts.Condition(func() (success bool) {
		res := testPolicy.GeneratePath(1, "/23")
		return res == "/1/23/456" || res == "\\1\\23\\456"
	})

}

func TestPolicy_GenerateFileName(t *testing.T) {
	asserts := assert.New(t)
	// 重命名关闭
	{
		testPolicy := Policy{
			AutoRename: false,
		}
		testPolicy.FileNameRule = "{randomkey16}"
		asserts.Equal("123.txt", testPolicy.GenerateFileName(1, "123.txt"))

		testPolicy.Type = "oss"
		asserts.Equal("origin", testPolicy.GenerateFileName(1, "origin"))
	}

	// 重命名开启
	{
		testPolicy := Policy{
			AutoRename: true,
		}

		testPolicy.FileNameRule = "{randomkey16}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 16)

		testPolicy.FileNameRule = "{randomkey8}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 8)

		testPolicy.FileNameRule = "{timestamp}"
		asserts.Equal(testPolicy.GenerateFileName(1, "123.txt"), strconv.FormatInt(time.Now().Unix(), 10))

		testPolicy.FileNameRule = "{uid}"
		asserts.Equal(testPolicy.GenerateFileName(1, "123.txt"), strconv.Itoa(int(1)))

		testPolicy.FileNameRule = "{datetime}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 14)

		testPolicy.FileNameRule = "{date}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 8)

		testPolicy.FileNameRule = "123{date}ss{datetime}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 27)

		testPolicy.FileNameRule = "{originname_without_ext}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 3)

		testPolicy.FileNameRule = "{originname_without_ext}_{randomkey8}{ext}"
		asserts.Len(testPolicy.GenerateFileName(1, "123.txt"), 16)

		// 支持{originname}的策略
		testPolicy.Type = "local"
		testPolicy.FileNameRule = "123{originname}"
		asserts.Equal("123123.txt", testPolicy.GenerateFileName(1, "123.txt"))

		testPolicy.Type = "qiniu"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123123.txt", testPolicy.GenerateFileName(1, "123.txt"))

		testPolicy.Type = "oss"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123123321", testPolicy.GenerateFileName(1, "123321"))

		testPolicy.Type = "upyun"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123123321", testPolicy.GenerateFileName(1, "123321"))

		testPolicy.Type = "qiniu"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123123321", testPolicy.GenerateFileName(1, "123321"))

		testPolicy.Type = "local"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123", testPolicy.GenerateFileName(1, ""))

		testPolicy.Type = "local"
		testPolicy.FileNameRule = "{ext}123{uuid}"
		asserts.Contains(testPolicy.GenerateFileName(1, "123.txt"), ".txt123")
	}

}

func TestPolicy_IsDirectlyPreview(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Type: "local"}
	asserts.True(policy.IsDirectlyPreview())
	policy.Type = "remote"
	asserts.False(policy.IsDirectlyPreview())
}

func TestPolicy_ClearCache(t *testing.T) {
	asserts := assert.New(t)
	cache.Set("policy_202", 1, 0)
	policy := Policy{Model: gorm.Model{ID: 202}}
	policy.ClearCache()
	_, ok := cache.Get("policy_202")
	asserts.False(ok)
}

func TestPolicy_UpdateAccessKey(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Model: gorm.Model{ID: 202}}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	policy.AccessKey = "123"
	err := policy.SaveAndClearCache()
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
}

func TestPolicy_Props(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Type: "onedrive"}
	policy.OptionsSerialized.PlaceholderWithSize = true
	asserts.False(policy.IsThumbGenerateNeeded())
	asserts.False(policy.IsTransitUpload(4))
	asserts.False(policy.IsTransitUpload(5 * 1024 * 1024))
	asserts.True(policy.CanStructureBeListed())
	asserts.True(policy.IsUploadPlaceholderWithSize())
	policy.Type = "local"
	asserts.True(policy.IsThumbGenerateNeeded())
	asserts.False(policy.CanStructureBeListed())
	asserts.False(policy.IsUploadPlaceholderWithSize())
	policy.Type = "remote"
	asserts.True(policy.IsUploadPlaceholderWithSize())
}

func TestPolicy_UpdateAccessKeyAndClearCache(t *testing.T) {
	a := assert.New(t)
	cache.Set("policy_1331", Policy{}, 3600)
	p := &Policy{}
	p.ID = 1331
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)").WithArgs("ak", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	a.NoError(p.UpdateAccessKeyAndClearCache("ak"))
	a.NoError(mock.ExpectationsWereMet())
	_, ok := cache.Get("policy_1331")
	a.False(ok)
}

func TestPolicy_CouldProxyThumb(t *testing.T) {
	a := assert.New(t)
	p := &Policy{Type: "local"}

	// local policy
	{
		a.False(p.CouldProxyThumb())
	}

	// feature not enabled
	{
		p.Type = "remote"
		cache.Set("setting_thumb_proxy_enabled", "0", 0)
		a.False(p.CouldProxyThumb())
	}

	// list not contain current policy
	{
		p.ID = 2
		cache.Set("setting_thumb_proxy_enabled", "1", 0)
		cache.Set("setting_thumb_proxy_policy", "[1]", 0)
		a.False(p.CouldProxyThumb())
	}

	// enabled
	{
		p.ID = 2
		cache.Set("setting_thumb_proxy_enabled", "1", 0)
		cache.Set("setting_thumb_proxy_policy", "[2]", 0)
		a.True(p.CouldProxyThumb())
	}

	cache.Deletes([]string{"thumb_proxy_enabled", "thumb_proxy_policy"}, "setting_")
}
