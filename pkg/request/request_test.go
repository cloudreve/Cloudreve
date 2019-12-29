package request

import (
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ClientMock struct {
	testMock.Mock
}

func (m ClientMock) Request(method, target string, body io.Reader, opts ...Option) Response {
	args := m.Called(method, target, body, opts)
	return args.Get(0).(Response)
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

}
