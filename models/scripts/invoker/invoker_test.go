package invoker

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestScript int

func (script TestScript) Run(ctx context.Context) {

}

func TestRunDBScript(t *testing.T) {
	asserts := assert.New(t)
	Register("test", TestScript(0))

	// 不存在
	{
		asserts.Error(RunDBScript("else", context.Background()))
	}

	// 存在
	{
		asserts.NoError(RunDBScript("test", context.Background()))
	}
}

func TestListPrefix(t *testing.T) {
	asserts := assert.New(t)
	Register("U1", TestScript(0))
	Register("U2", TestScript(0))
	Register("U3", TestScript(0))
	Register("P1", TestScript(0))

	res := ListPrefix("U")
	asserts.Len(res, 3)
}
