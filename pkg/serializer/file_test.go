package serializer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeUploadPolicy(t *testing.T) {
	asserts := assert.New(t)

	testCases := []struct {
		input       string
		expectError bool
		expectNil   bool
		expectRes   *UploadPolicy
	}{
		{
			"错误的base64字符",
			true,
			true,
			&UploadPolicy{},
		},
		{
			"6ZSZ6K+v55qESlNPTuWtl+espg==",
			true,
			true,
			&UploadPolicy{},
		},
		{
			"e30=",
			false,
			false,
			&UploadPolicy{},
		},
		{
			"eyJjYWxsYmFja19rZXkiOiJ0ZXN0In0=",
			false,
			false,
			&UploadPolicy{CallbackKey: "test"},
		},
	}

	for _, testCase := range testCases {
		res, err := DecodeUploadPolicy(testCase.input)
		if testCase.expectError {
			asserts.Error(err)
		}
		if testCase.expectNil {
			asserts.Nil(res)
		}
		if !testCase.expectNil {
			asserts.Equal(testCase.expectRes, res)
		}
	}
}
