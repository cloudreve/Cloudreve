package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetLevelText(t *testing.T) {
	asserts := assert.New(t)

	asserts.Equal("Debug", GetLevelText(logrus.DebugLevel))
	asserts.Equal("Info", GetLevelText(logrus.InfoLevel))
	asserts.Equal("Warning", GetLevelText(logrus.WarnLevel))
	asserts.Equal("Error", GetLevelText(logrus.ErrorLevel))
	asserts.Equal("Fatal", GetLevelText(logrus.FatalLevel))
	asserts.Equal("Panic", GetLevelText(logrus.PanicLevel))
}

func TestFormatter_Format(t *testing.T) {
	asserts := assert.New(t)

	entry := logrus.NewEntry(logrus.StandardLogger())
	entry.Level = logrus.DebugLevel
	entry.Message = "test"
	entry.Time = time.Now()

	f := &formatter{}
	b, err := f.Format(entry)
	asserts.NoError(err)
	asserts.Contains(string(b), "test")
}

func BenchmarkFormatter_Format(b *testing.B) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry.Level = logrus.DebugLevel
	entry.Message = "test"
	entry.Time = time.Now()

	f := &formatter{}
	for i := 0; i < b.N; i++ {
		f.Format(entry)
	}
}
