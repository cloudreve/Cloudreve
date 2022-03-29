package requestmock

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/stretchr/testify/mock"
	"io"
)

type RequestMock struct {
	mock.Mock
}

func (r RequestMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	return r.Called(method, target, body, opts).Get(0).(*request.Response)
}
