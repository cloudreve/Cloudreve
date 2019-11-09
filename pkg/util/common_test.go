package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandStringRunes(t *testing.T) {
	asserts := assert.New(t)

	// 0 长度字符
	randStr := RandStringRunes(0)
	asserts.Len(randStr, 0)

	// 16 长度字符
	randStr = RandStringRunes(16)
	asserts.Len(randStr, 16)

	// 32 长度字符
	randStr = RandStringRunes(32)
	asserts.Len(randStr, 32)

	//相同长度字符
	sameLenStr1 := RandStringRunes(32)
	sameLenStr2 := RandStringRunes(32)
	asserts.NotEqual(sameLenStr1, sameLenStr2)
}
