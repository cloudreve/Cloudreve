package request

import (
	"context"
	"errors"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ClientMock struct {
	testMock.Mock
}

func (m ClientMock) Request(method, target string, body io.Reader, opts ...Option) *Response {
	args := m.Called(method, target, body, opts)
	return args.Get(0).(*Response)
}

func TestWithTimeout(t *testing.T) {
	asserts := assert.New(t)
	options := newDefaultOption()
	WithTimeout(time.Duration(5) * time.Second).apply(options)
	asserts.Equal(time.Duration(5)*time.Second, options.timeout)
}

func TestWithHeader(t *testing.T) {
	asserts := assert.New(t)
	options := newDefaultOption()
	WithHeader(map[string][]string{"Origin": []string{"123"}}).apply(options)
	asserts.Equal(http.Header{"Origin": []string{"123"}}, options.header)
}

func TestWithCredential(t *testing.T) {
	asserts := assert.New(t)
	options := newDefaultOption()
	WithCredential(auth.HMACAuth{SecretKey: []byte("123")}, 10).apply(options)
	asserts.Equal(auth.HMACAuth{SecretKey: []byte("123")}, options.sign)
	asserts.EqualValues(10, options.signTTL)
}

func TestWithContext(t *testing.T) {
	asserts := assert.New(t)
	options := newDefaultOption()
	WithContext(context.Background()).apply(options)
	asserts.NotNil(options.ctx)
}

func TestHTTPClient_Request(t *testing.T) {
	asserts := assert.New(t)
	client := HTTPClient{}

	// 正常
	{
		resp := client.Request(
			"GET",
			"http://cloudreveisnotexist.com",
			strings.NewReader(""),
			WithTimeout(time.Duration(1)*time.Microsecond),
			WithCredential(auth.HMACAuth{SecretKey: []byte("123")}, 10),
		)
		asserts.Error(resp.Err)
		asserts.Nil(resp.Response)
	}

	// 正常 带有ctx
	{
		resp := client.Request(
			"GET",
			"http://cloudreveisnotexist.com",
			strings.NewReader(""),
			WithTimeout(time.Duration(1)*time.Microsecond),
			WithCredential(auth.HMACAuth{SecretKey: []byte("123")}, 10),
			WithContext(context.Background()),
		)
		asserts.Error(resp.Err)
		asserts.Nil(resp.Response)
	}

}

func TestResponse_GetResponse(t *testing.T) {
	asserts := assert.New(t)

	// 直接返回错误
	{
		resp := Response{
			Err: errors.New("error"),
		}
		content, err := resp.GetResponse()
		asserts.Empty(content)
		asserts.Error(err)
	}

	// 正常
	{
		resp := Response{
			Response: &http.Response{Body: ioutil.NopCloser(strings.NewReader("123"))},
		}
		content, err := resp.GetResponse()
		asserts.Equal("123", content)
		asserts.NoError(err)
	}
}

func TestResponse_CheckHTTPResponse(t *testing.T) {
	asserts := assert.New(t)

	// 直接返回错误
	{
		resp := Response{
			Err: errors.New("error"),
		}
		res := resp.CheckHTTPResponse(200)
		asserts.Error(res.Err)
	}

	// 404错误
	{
		resp := Response{
			Response: &http.Response{StatusCode: 404},
		}
		res := resp.CheckHTTPResponse(200)
		asserts.Error(res.Err)
	}

	// 通过
	{
		resp := Response{
			Response: &http.Response{StatusCode: 200},
		}
		res := resp.CheckHTTPResponse(200)
		asserts.NoError(res.Err)
	}
}

func TestResponse_GetRSCloser(t *testing.T) {
	asserts := assert.New(t)

	// 直接返回错误
	{
		resp := Response{
			Err: errors.New("error"),
		}
		res, err := resp.GetRSCloser()
		asserts.Error(err)
		asserts.Nil(res)
	}

	// 正常
	{
		resp := Response{
			Response: &http.Response{ContentLength: 3, Body: ioutil.NopCloser(strings.NewReader("123"))},
		}
		res, err := resp.GetRSCloser()
		asserts.NoError(err)
		content, err := ioutil.ReadAll(res)
		asserts.NoError(err)
		asserts.Equal("123", string(content))
		offset, err := res.Seek(0, 0)
		asserts.NoError(err)
		asserts.Equal(int64(0), offset)
		offset, err = res.Seek(0, 2)
		asserts.NoError(err)
		asserts.Equal(int64(3), offset)
		_, err = res.Seek(1, 2)
		asserts.Error(err)
		asserts.NoError(res.Close())
	}

}
