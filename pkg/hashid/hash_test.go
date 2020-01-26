package hashid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashEncode(t *testing.T) {
	asserts := assert.New(t)

	{
		res, err := HashEncode([]int{1, 2, 3})
		asserts.NoError(err)
		asserts.NotEmpty(res)
	}

	{
		res, err := HashEncode([]int{})
		asserts.Error(err)
		asserts.Empty(res)
	}

}

func TestHashID(t *testing.T) {
	asserts := assert.New(t)

	{
		res := HashID(1, ShareID)
		asserts.NotEmpty(res)
	}
}
