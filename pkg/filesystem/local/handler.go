package local

import (
	"context"
	"fmt"
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
		fmt.Println("创建", basePath)
		err := os.MkdirAll(basePath, 0666)
		if err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return err
}
