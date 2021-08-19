package util

import (
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
)

var (
	OS = afero.NewOsFs()
)

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := OS.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (afero.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := OS.MkdirAll(basePath, 0700)
		if err != nil {
			Log().Warning("无法创建目录，%s", err)
			return nil, err
		}
	}

	return OS.Create(path)
}

// IsEmpty 返回给定目录是否为空目录
func IsEmpty(name string) (bool, error) {
	f, err := OS.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
