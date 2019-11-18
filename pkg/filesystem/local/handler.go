package local

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"os"
	"path/filepath"
)

type Handler struct {
}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string) error {
	defer file.Close()

	// 如果目标目录不存在，创建
	basePath := filepath.Dir(dst)
	if !util.Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			util.Log().Warning("无法创建目录，%s", err)
			return err
		}
	}

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		util.Log().Warning("无法创建文件，%s", err)
		return err
	}
	defer out.Close()

	// 写入文件内容
	_, err = io.Copy(out, file)
	return err
}

// Delete 删除一个或多个文件，
// 返回已删除的文件，及遇到的最后一个错误
func (handler Handler) Delete(ctx context.Context, files []string) ([]string, error) {
	deleted := make([]string, 0, len(files))
	var retErr error

	for _, value := range files {
		err := os.Remove(value)
		if err == nil {
			deleted = append(deleted, value)
			util.Log().Warning("无法删除文件，%s", err)
		} else {
			retErr = err
		}
	}

	return deleted, retErr
}
