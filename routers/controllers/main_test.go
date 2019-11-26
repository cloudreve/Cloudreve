package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v8"
	"testing"
)

var mock sqlmock.Sqlmock
var memDB *gorm.DB

// TestMain 初始化数据库Mock
func TestMain(m *testing.M) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化sqlmock
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic("An error was not expected when opening a stub database connection")
	}

	// 初始话内存数据库
	model.Init()
	memDB = model.DB

	model.DB, _ = gorm.Open("mysql", db)
	defer db.Close()

	switchToMemDB()

	m.Run()
}

func switchToMemDB() {
	model.DB = memDB
}

// 测试 ErrorResponse
func TestErrorResponse(t *testing.T) {
	asserts := assert.New(t)

	type Test struct {
		UserName string `validate:"required,min=10,max=20"`
	}

	testCase1 := Test{}
	validate := validator.New(&validator.Config{TagName: "validate"})
	errs := validate.Struct(testCase1)
	res := ErrorResponse(errs)
	asserts.Equal(serializer.ParamErr(ParamErrorMsg("UserName", "required"), errs), res)

	testCase2 := Test{
		UserName: "a",
	}
	validate2 := validator.New(&validator.Config{TagName: "validate"})
	errs = validate2.Struct(testCase2)
	res = ErrorResponse(errs)
	asserts.Equal(serializer.ParamErr(ParamErrorMsg("UserName", "min"), errs), res)

	type JsonTest struct {
		UserName string `json:"UserName"`
	}
	testCase3 := JsonTest{}
	errs = json.Unmarshal([]byte("{\"UserName\":1}"), &testCase3)
	res = ErrorResponse(errs)
	asserts.Equal(serializer.ParamErr("JSON类型不匹配", errs), res)

	errs = errors.New("Unknown error")
	res = ErrorResponse(errs)
	asserts.Equal(serializer.ParamErr("参数错误", errs), res)
}

// 测试 ParamErrorMs
func TestParamErrorMsg(t *testing.T) {
	asserts := assert.New(t)
	testCase := []struct {
		field  string
		tag    string
		expect string
	}{
		{
			"UserName",
			"required",
			"邮箱不能为空",
		},
		{
			"Password",
			"min",
			"密码太短",
		},
		{
			"Password",
			"max",
			"密码太长",
		},
		{
			"something unexpected",
			"max",
			"",
		},
		{
			"",
			"",
			"",
		},
	}
	for _, value := range testCase {
		asserts.Equal(value.expect, ParamErrorMsg(value.field, value.tag), "case %v", value)
	}
}
