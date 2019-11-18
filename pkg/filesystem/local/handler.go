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
			return err
		}
	}

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入文件内容
	_, err = io.Copy(out, file)
	return err
}
