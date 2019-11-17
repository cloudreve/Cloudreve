package model

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestGetPolicyByID(t *testing.T) {
	asserts := assert.New(t)

	rows := sqlmock.NewRows([]string{"name", "type", "options"}).
		AddRow("默认上传策略", "local", "{\"op_name\":\"123\"}")
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND \\(\\(`policies`.`id` = 1\\)\\)(.+)$").WillReturnRows(rows)
	policy, err := GetPolicyByID(1)
	asserts.NoError(err)
	asserts.Equal("默认上传策略", policy.Name)
	asserts.Equal("123", policy.OptionsSerialized.OPName)

	rows = sqlmock.NewRows([]string{"name", "type", "options"})
	mock.ExpectQuery("^SELECT \\* FROM `(.+)` WHERE `(.+)`\\.`deleted_at` IS NULL AND \\(\\(`policies`.`id` = 1\\)\\)(.+)$").WillReturnRows(rows)
	policy, err = GetPolicyByID(1)
	asserts.Error(err)
}

func TestPolicy_BeforeSave(t *testing.T) {
	asserts := assert.New(t)

	testPolicy := Policy{
		OptionsSerialized: PolicyOption{
			OPName: "123",
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
	asserts.Len(testPolicy.GeneratePath(1), 16)

	testPolicy.DirNameRule = "{randomkey8}"
	asserts.Len(testPolicy.GeneratePath(1), 8)

	testPolicy.DirNameRule = "{timestamp}"
	asserts.Equal(testPolicy.GeneratePath(1), strconv.FormatInt(time.Now().Unix(), 10))

	testPolicy.DirNameRule = "{uid}"
	asserts.Equal(testPolicy.GeneratePath(1), strconv.Itoa(int(1)))

	testPolicy.DirNameRule = "{datetime}"
	asserts.Len(testPolicy.GeneratePath(1), 14)

	testPolicy.DirNameRule = "{date}"
	asserts.Len(testPolicy.GeneratePath(1), 8)

	testPolicy.DirNameRule = "123{date}ss{datetime}"
	asserts.Len(testPolicy.GeneratePath(1), 27)
}

func TestPolicy_GenerateFileName(t *testing.T) {
	asserts := assert.New(t)
	testPolicy := Policy{}

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
	asserts.Equal("1123$(fname)", testPolicy.GenerateFileName(1, "123.txt"))

	testPolicy.Type = "oss"
	testPolicy.FileNameRule = "{uid}123{originname}"
	asserts.Equal("1123${filename}", testPolicy.GenerateFileName(1, ""))

	testPolicy.Type = "upyun"
	testPolicy.FileNameRule = "{uid}123{originname}"
	asserts.Equal("1123{filename}{.suffix}", testPolicy.GenerateFileName(1, ""))
}
