package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLogger(t *testing.T) {
	asserts := assert.New(t)
	logger := NewLogger()
	asserts.NotNil(logger)
}

func TestSetLevel(t *testing.T) {
	asserts := assert.New(t)
	SetLevel(logrus.DebugLevel)
	asserts.Equal(logrus.DebugLevel, GlobalLogger.GetLevel())
}
