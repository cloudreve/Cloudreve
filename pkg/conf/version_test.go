package conf

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteVersionLock(t *testing.T) {
	asserts := assert.New(t)

	// 清理残余文件
	if _, err := os.Stat("version.lock"); !os.IsNotExist(err) {
		err = os.Remove("version.lock")
		asserts.NoError(err)
	}

	err := WriteVersionLock()
	defer func() { err = os.Remove("version.lock") }()
	writtenVersion, err := ioutil.ReadFile("version.lock")

	// 写入的版本应与当前版本相同
	asserts.NoError(err)
	asserts.Equal(string(writtenVersion), BackendVersion)

}
