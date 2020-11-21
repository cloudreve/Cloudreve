package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/middleware"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/stretchr/testify/assert"
)

func TestListDirectoryRoute(t *testing.T) {
	switchToMemDB()
	asserts := assert.New(t)
	router := InitMasterRouter()
	w := httptest.NewRecorder()

	// 成功
	req, _ := http.NewRequest(
		"GET",
		"/api/v3/directory/",
		nil,
	)
	middleware.SessionMock = map[string]interface{}{"user_id": 1}
	router.ServeHTTP(w, req)
	asserts.Equal(200, w.Code)
	resJSON := &serializer.Response{}
	err := json.Unmarshal(w.Body.Bytes(), resJSON)
	asserts.NoError(err)
	asserts.Equal(0, resJSON.Code)

	w.Body.Reset()

}

func TestLocalFileUpload(t *testing.T) {
	switchToMemDB()
	asserts := assert.New(t)
	router := InitMasterRouter()
	w := httptest.NewRecorder()
	middleware.SessionMock = map[string]interface{}{"user_id": 1}

	testCases := []struct {
		GetRequest func() *http.Request
		ExpectCode int
		RollBack   func()
	}{
		// 文件大小指定错误
		{
			GetRequest: func() *http.Request {
				req, _ := http.NewRequest(
					"POST",
					"/api/v3/file/upload",
					nil,
				)
				req.Header.Add("Content-Length", "ddf")
				return req
			},
			ExpectCode: 40001,
		},
		// 返回错误
		{
			GetRequest: func() *http.Request {
				req, _ := http.NewRequest(
					"POST",
					"/api/v3/file/upload",
					strings.NewReader("2333"),
				)
				req.Header.Add("Content-Length", "4")
				req.Header.Add("X-FileName", "大地的%sfsf")
				return req
			},
			ExpectCode: 40002,
		},
		// 成功
		{
			GetRequest: func() *http.Request {
				req, _ := http.NewRequest(
					"POST",
					"/api/v3/file/upload",
					strings.NewReader("2333"),
				)
				req.Header.Add("Content-Length", "4")
				req.Header.Add("X-FileName", "TestFileUploadRoute.txt")
				req.Header.Add("X-Path", "/")
				return req
			},
			ExpectCode: 0,
		},
	}

	for key, testCase := range testCases {
		req := testCase.GetRequest()
		router.ServeHTTP(w, req)
		asserts.Equal(200, w.Code)
		resJSON := &serializer.Response{}
		err := json.Unmarshal(w.Body.Bytes(), resJSON)
		asserts.NoError(err, "测试用例%d", key)
		asserts.Equal(testCase.ExpectCode, resJSON.Code, "测试用例%d", key)
		if testCase.RollBack != nil {
			testCase.RollBack()
		}
		w.Body.Reset()
	}

}

func TestObjectDelete(t *testing.T) {
	asserts := assert.New(t)
	router := InitMasterRouter()
	w := httptest.NewRecorder()
	middleware.SessionMock = map[string]interface{}{"user_id": 1}

	testCases := []struct {
		Mock       []string
		GetRequest func() *http.Request
		ExpectCode int
		RollBack   []string
	}{
		// 路径不存在，返回无错误
		{
			GetRequest: func() *http.Request {
				body := explorer.ItemService{
					Items: []uint{1},
				}
				bodyStr, _ := json.Marshal(body)
				req, _ := http.NewRequest(
					"DELETE",
					"/api/v3/object",
					bytes.NewReader(bodyStr),
				)
				return req
			},
			ExpectCode: 0,
		},
		// 文件删除失败，返回203
		{
			Mock: []string{"INSERT INTO `files` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `source_name`, `user_id`, `size`, `pic_info`, `folder_id`, `policy_id`) VALUES(5, '2019-11-30 07:08:33', '2019-11-30 07:08:33', NULL, 'pigeon.zip', '65azil3B_pigeon.zip', 1, 1667217, '', 1, 1);"},
			GetRequest: func() *http.Request {
				body := explorer.ItemService{
					Items: []uint{5},
				}
				bodyStr, _ := json.Marshal(body)
				req, _ := http.NewRequest(
					"DELETE",
					"/api/v3/object",
					bytes.NewReader(bodyStr),
				)
				return req
			},
			RollBack:   []string{"DELETE FROM `v3_files` WHERE `id`=5"},
			ExpectCode: 203,
		},
	}

	for key, testCase := range testCases {
		for _, value := range testCase.Mock {
			model.DB.Exec(value)
		}
		req := testCase.GetRequest()
		router.ServeHTTP(w, req)
		asserts.Equal(200, w.Code)
		resJSON := &serializer.Response{}
		err := json.Unmarshal(w.Body.Bytes(), resJSON)
		asserts.NoError(err, "测试用例%d", key)
		asserts.Equal(testCase.ExpectCode, resJSON.Code, "测试用例%d", key)

		for _, value := range testCase.RollBack {
			model.DB.Exec(value)
		}

		w.Body.Reset()
	}
}
