package util

import (
	"os"
	"path/filepath"
)

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			Log().Warning("无法创建目录，%s", err)
			return nil, err
		}
	}

	return os.Create(path)
}
