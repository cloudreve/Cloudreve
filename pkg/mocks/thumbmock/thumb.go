package thumbmock

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/stretchr/testify/mock"
	"io"
)

type GeneratorMock struct {
	mock.Mock
}

func (g GeneratorMock) Generate(ctx context.Context, file io.Reader, src string, name string, options map[string]string) (*thumb.Result, error) {
	res := g.Called(ctx, file, src, name, options)
	return res.Get(0).(*thumb.Result), res.Error(1)
}

func (g GeneratorMock) Priority() int {
	return 0
}

func (g GeneratorMock) EnableFlag() string {
	return "thumb_vips_enabled"
}
