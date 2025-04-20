package util

import (
	"io"
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
			return nil, err
		}
	}

	return os.Create(path)
}

// CreatNestedFolder creates a folder with the given path, if the directory does not exist,
// it will be created recursively.
func CreatNestedFolder(path string) error {
	if !Exists(path) {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsEmpty 返回给定目录是否为空目录
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
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

type CallbackReader struct {
	reader   io.Reader
	callback func(int64)
}

func NewCallbackReader(reader io.Reader, callback func(int64)) *CallbackReader {
	return &CallbackReader{
		reader:   reader,
		callback: callback,
	}
}

func (r *CallbackReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.callback(int64(n))
	return
}
