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

func TestHashDecode(t *testing.T) {
	asserts := assert.New(t)

	// 正常
	{
		res, _ := HashEncode([]int{1, 2, 3})
		decodeRes, err := HashDecode(res)
		asserts.NoError(err)
		asserts.Equal([]int{1, 2, 3}, decodeRes)
	}

	// 出错
	{
		decodeRes, err := HashDecode("233")
		asserts.Error(err)
		asserts.Len(decodeRes, 0)
	}
}

func TestDecodeHashID(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		uid, err := DecodeHashID(HashID(1, ShareID), ShareID)
		asserts.NoError(err)
		asserts.EqualValues(1, uid)
	}

	// 类型不匹配
	{
		uid, err := DecodeHashID(HashID(1, ShareID), UserID)
		asserts.Error(err)
		asserts.EqualValues(0, uid)
	}
}
