package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEnvStr(t *testing.T) {
	asserts := assert.New(t)

	{
		asserts.Equal("default", EnvStr("not_exist", "default"))
	}
	{
		err := os.Setenv("exist", "value")
		asserts.NoError(err)
		asserts.Equal("value", EnvStr("exist", "default"))

		err = os.Unsetenv("exist")
		asserts.NoError(err)
		asserts.Equal("default", EnvStr("exist", "default"))
	}
}

func TestEnvInt(t *testing.T) {
	asserts := assert.New(t)

	{
		asserts.Equal(1, EnvInt("not_exist", 1))
	}
	{
		err := os.Setenv("exist", "2")
		asserts.NoError(err)
		asserts.Equal(2, EnvInt("exist", 1))

		err = os.Unsetenv("exist")
		asserts.NoError(err)
		asserts.Equal(1, EnvInt("exist", 1))
	}
	{
		err := os.Setenv("exist", "not_number")
		asserts.NoError(err)
		asserts.Equal(1, EnvInt("exist", 1))
	}
}

func TestEnvArr(t *testing.T) {
	asserts := assert.New(t)

	{
		asserts.Equal([]string{"default"}, EnvArr("not_exist", []string{"default"}))
	}
	{
		err := os.Setenv("exist", "value1,value2")
		asserts.NoError(err)
		asserts.Equal([]string{"value1", "value2"}, EnvArr("exist", []string{"default"}))

		err = os.Unsetenv("exist")
		asserts.NoError(err)
		asserts.Equal([]string{"default"}, EnvArr("exist", []string{"default"}))
	}
}
