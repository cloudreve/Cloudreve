package conf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/stretchr/testify/assert"
)

// 测试Init日志路径错误
func TestInitPanic(t *testing.T) {
	asserts := assert.New(t)

	// 日志路径不存在时
	asserts.NotPanics(func() {
		Init("not/exist/path/conf.ini")
	})

	asserts.True(util.Exists("not/exist/path/conf.ini"))

}

// TestInitDelimiterNotFound 日志路径存在但 Key 格式错误时
func TestInitDelimiterNotFound(t *testing.T) {
	asserts := assert.New(t)
	testCase := `[Database]
Type = mysql
User = root
Password233root
Host = 127.0.0.1:3306
Name = v3
TablePrefix = v3_`
	err := ioutil.WriteFile("testConf.ini", []byte(testCase), 0644)
	defer func() { err = os.Remove("testConf.ini") }()
	if err != nil {
		panic(err)
	}
	asserts.Panics(func() {
		Init("testConf.ini")
	})
}

// TestInitNoPanic 日志路径存在且合法时
func TestInitNoPanic(t *testing.T) {
	asserts := assert.New(t)
	testCase := `
[System]
Listen = 3000
HashIDSalt = 1

[Database]
Type = mysql
User = root
Password = root
Host = 127.0.0.1:3306
Name = v3
TablePrefix = v3_`
	err := ioutil.WriteFile("testConf.ini", []byte(testCase), 0644)
	defer func() { err = os.Remove("testConf.ini") }()
	if err != nil {
		panic(err)
	}
	asserts.NotPanics(func() {
		Init("testConf.ini")
	})
}

func TestMapSection(t *testing.T) {
	asserts := assert.New(t)

	//正常情况
	testCase := `
[System]
Listen = 3000
HashIDSalt = 1

[Database]
Type = mysql
User = root
Password:root
Host = 127.0.0.1:3306
Name = v3
TablePrefix = v3_`
	err := ioutil.WriteFile("testConf.ini", []byte(testCase), 0644)
	defer func() { err = os.Remove("testConf.ini") }()
	if err != nil {
		panic(err)
	}
	Init("testConf.ini")
	err = mapSection("Database", DatabaseConfig)
	asserts.NoError(err)

}
