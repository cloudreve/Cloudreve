package oss

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestGetPublicKey(t *testing.T) {
	asserts := assert.New(t)
	testCases := []struct {
		Request http.Request
		ResNil  bool
		Error   bool
	}{
		// Header解码失败
		{
			Request: http.Request{
				Header: http.Header{
					"X-Oss-Pub-Key-Url": {"中文"},
				},
			},
			ResNil: true,
			Error:  true,
		},
		// 公钥URL无效
		{
			Request: http.Request{
				Header: http.Header{
					"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9wb3JuaHViLmNvbQ=="},
				},
			},
			ResNil: true,
			Error:  true,
		},
		// 请求失败
		{
			Request: http.Request{
				Header: http.Header{
					"X-Oss-Pub-Key-Url": {"aHR0cDovL2dvc3NwdWJsaWMuYWxpY2RuLmNvbS8yMzQyMzQ="},
				},
			},
			ResNil: true,
			Error:  true,
		},
		// 成功
		{
			Request: http.Request{
				Header: http.Header{
					"X-Oss-Pub-Key-Url": {"aHR0cDovL2dvc3NwdWJsaWMuYWxpY2RuLmNvbS9jYWxsYmFja19wdWJfa2V5X3YxLnBlbQ=="},
				},
			},
			ResNil: false,
			Error:  false,
		},
	}

	for i, testCase := range testCases {
		asserts.NoError(cache.Deletes([]string{"oss_public_key"}, ""))
		res, err := GetPublicKey(&testCase.Request)
		if testCase.Error {
			asserts.Error(err, "Test Case #%d", i)
		} else {
			asserts.NoError(err, "Test Case #%d", i)
		}
		if testCase.ResNil {
			asserts.Empty(res, "Test Case #%d", i)
		} else {
			asserts.NotEmpty(res, "Test Case #%d", i)
		}
	}

	// 测试缓存
	asserts.NoError(cache.Set("oss_public_key", []byte("123"), 0))
	res, err := GetPublicKey(nil)
	asserts.NoError(err)
	asserts.Equal([]byte("123"), res)
}

func TestVerifyCallbackSignature(t *testing.T) {
	asserts := assert.New(t)
	testPubKey := `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKs/JBGzwUB2aVht4crBx3oIPBLNsjGs
C0fTXv+nvlmklvkcolvpvXLTjaxUHR3W9LXxQ2EHXAJfCB+6H2YF1k8CAwEAAQ==
-----END PUBLIC KEY-----
`

	// 成功
	{
		asserts.NoError(cache.Set("oss_public_key", []byte(testPubKey), 0))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"Authorization":     {"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.NoError(VerifyCallbackSignature(&r))
	}

	// 签名错误
	{
		asserts.NoError(cache.Set("oss_public_key", []byte(testPubKey), 0))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"Authorization":     {"e3LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}

	// GetPubKey 失败
	{
		asserts.NoError(cache.Deletes([]string{"oss_public_key"}, ""))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"Authorization": {"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}

	// getRequestMD5 失败
	{
		asserts.NoError(cache.Set("oss_public_key", []byte(testPubKey), 0))
		r := http.Request{
			URL: &url.URL{Path: "%测试"},
			Header: map[string][]string{
				"Authorization":     {"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}

	// 无 Authorization 头
	{
		asserts.NoError(cache.Set("oss_public_key", []byte(testPubKey), 0))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}

	// pub block 不存在
	{
		asserts.NoError(cache.Set("oss_public_key", []byte(""), 0))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"Authorization":     {"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}

	// ParsePKIXPublicKey出错
	{
		asserts.NoError(cache.Set("oss_public_key", []byte("-----BEGIN PUBLIC KEY-----\n-----END PUBLIC KEY-----"), 0))
		r := http.Request{
			URL: &url.URL{Path: "/api/v3/callback/oss/TnXx5E5VyfJUyM1UdkdDu1rtnJ34EbmH"},
			Header: map[string][]string{
				"Authorization":     {"e5LwzwTkP9AFAItT4YzvdJOHd0Y0wqTMWhsV/h5SG90JYGAmMd+8LQyj96R+9qUfJWjMt6suuUh7LaOryR87Dw=="},
				"X-Oss-Pub-Key-Url": {"aHR0cHM6Ly9nb3NzcHVibGljLmFsaWNkbi5jb20vY2FsbGJhY2tfcHViX2tleV92MS5wZW0="},
			},
			Body: ioutil.NopCloser(strings.NewReader(`{"name":"2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","source_name":"1/1_hFRtDLgM_2f7b2ccf30e9270ea920f1ab8a4037a546a2f0d5.jpg","size":114020,"pic_info":"810,539"}`)),
		}
		asserts.Error(VerifyCallbackSignature(&r))
	}
}
