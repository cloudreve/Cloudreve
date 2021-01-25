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
		asserts.Equal("123", policy.OptionsSerialized.OdRedirect)

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
		asserts.Equal("123", policy.OptionsSerialized.OdRedirect)

	}

}

func TestPolicy_BeforeSave(t *testing.T) {
	asserts := assert.New(t)

	testPolicy := Policy{
		OptionsSerialized: PolicyOption{
			OdRedirect: "123",
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
		asserts.Equal("${filename}", testPolicy.GenerateFileName(1, ""))
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

		// 支持{originname}的策略
		testPolicy.Type = "local"
		testPolicy.FileNameRule = "123{originname}"
		asserts.Equal("123123.txt", testPolicy.GenerateFileName(1, "123.txt"))

		testPolicy.Type = "qiniu"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123123.txt", testPolicy.GenerateFileName(1, "123.txt"))

		testPolicy.Type = "oss"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123${filename}", testPolicy.GenerateFileName(1, ""))

		testPolicy.Type = "upyun"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123{filename}{.suffix}", testPolicy.GenerateFileName(1, ""))

		testPolicy.Type = "qiniu"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123$(fname)", testPolicy.GenerateFileName(1, ""))

		testPolicy.Type = "local"
		testPolicy.FileNameRule = "{uid}123{originname}"
		asserts.Equal("1123", testPolicy.GenerateFileName(1, ""))
	}

}

func TestPolicy_IsDirectlyPreview(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Type: "local"}
	asserts.True(policy.IsDirectlyPreview())
	policy.Type = "remote"
	asserts.False(policy.IsDirectlyPreview())
}

func TestPolicy_GetUploadURL(t *testing.T) {
	asserts := assert.New(t)

	// 本地
	{
		cache.Set("setting_siteURL", "http://127.0.0.1", 0)
		policy := Policy{Type: "local", Server: "http://127.0.0.1"}
		asserts.Equal("/api/v3/file/upload", policy.GetUploadURL())
	}

	// 远程
	{
		policy := Policy{Type: "remote", Server: "http://127.0.0.1"}
		asserts.Equal("http://127.0.0.1/api/v3/slave/upload", policy.GetUploadURL())
	}

	// OSS
	{
		policy := Policy{Type: "oss", BucketName: "base", Server: "127.0.0.1"}
		asserts.Equal("https://base.127.0.0.1", policy.GetUploadURL())
	}

	// cos
	{
		policy := Policy{Type: "cos", BaseURL: "base", Server: "http://127.0.0.1"}
		asserts.Equal("http://127.0.0.1", policy.GetUploadURL())
	}

	// upyun
	{
		policy := Policy{Type: "upyun", BucketName: "base", Server: "http://127.0.0.1"}
		asserts.Equal("https://v0.api.upyun.com/base", policy.GetUploadURL())
	}

	// 未知
	{
		policy := Policy{Type: "unknown", Server: "http://127.0.0.1"}
		asserts.Equal("http://127.0.0.1", policy.GetUploadURL())
	}

	// S3 未填写自动生成
	{
		policy := Policy{
			Type:              "s3",
			Server:            "",
			BucketName:        "bucket",
			OptionsSerialized: PolicyOption{Region: "us-east"},
		}
		asserts.Equal("https://bucket.s3.us-east.amazonaws.com/", policy.GetUploadURL())
	}

	// s3 自己指定
	{
		policy := Policy{
			Type:              "s3",
			Server:            "https://s3.us-east.amazonaws.com/",
			BucketName:        "bucket",
			OptionsSerialized: PolicyOption{Region: "us-east"},
		}
		asserts.Equal("https://s3.us-east.amazonaws.com/bucket", policy.GetUploadURL())
	}

}

func TestPolicy_IsPathGenerateNeeded(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Type: "qiniu"}
	asserts.True(policy.IsPathGenerateNeeded())
	policy.Type = "remote"
	asserts.False(policy.IsPathGenerateNeeded())
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
	err := policy.UpdateAccessKey("123")
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(err)
}

func TestPolicy_Props(t *testing.T) {
	asserts := assert.New(t)
	policy := Policy{Type: "onedrive"}
	asserts.False(policy.IsThumbGenerateNeeded())
	asserts.True(policy.IsPathGenerateNeeded())
	asserts.True(policy.IsTransitUpload(4))
	asserts.False(policy.IsTransitUpload(5 * 1024 * 1024))
	asserts.True(policy.CanStructureBeListed())
	policy.Type = "local"
	asserts.True(policy.IsThumbGenerateNeeded())
	asserts.True(policy.IsPathGenerateNeeded())
	asserts.False(policy.CanStructureBeListed())
}

func TestPolicy_IsThumbExist(t *testing.T) {
	asserts := assert.New(t)

	testCases := []struct {
		name   string
		expect bool
		policy string
	}{
		{
			"1.png",
			false,
			"unknown",
		},
		{
			"1.png",
			false,
			"local",
		},
		{
			"1.png",
			true,
			"cos",
		},
		{
			"1",
			false,
			"cos",
		},
		{
			"1.txt.png",
			true,
			"cos",
		},
		{
			"1.png.txt",
			false,
			"cos",
		},
		{
			"1",
			true,
			"onedrive",
		},
	}

	for _, testCase := range testCases {
		policy := Policy{Type: testCase.policy}
		asserts.Equal(testCase.expect, policy.IsThumbExist(testCase.name))
	}
}
