package model

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
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
