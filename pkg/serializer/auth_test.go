package serializer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRequestSignString(t *testing.T) {
	asserts := assert.New(t)

	sign := NewRequestSignString("1", "2", "3")
	asserts.NotEmpty(sign)
}
