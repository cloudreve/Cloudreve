package local

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"os"
	"path/filepath"
)

// Handler 本地策略适配器
type Handler struct{}

// Get 获取文件内容
func (handler Handler) Get(ctx context.Context, path string) (io.ReadSeeker, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		util.Log().Debug("无法打开文件：%s", err)
		return nil, err
	}

	// 开启一个协程，用于请求结束后关闭reader
	go closeReader(ctx, file)

	return file, nil
}

// closeReader 用于在请求结束后关闭reader
func closeReader(ctx context.Context, closer io.Closer) {
	select {
	case <-ctx.Done():
		_ = closer.Close()

	}
}

// Put 将文件流保存到指定目录
func (handler Handler) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
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
// 返回未删除的文件，及遇到的最后一个错误
func (handler Handler) Delete(ctx context.Context, files []string) ([]string, error) {
	deleteFailed := make([]string, 0, len(files))
	var retErr error

	for _, value := range files {
		err := os.Remove(value)
		if err != nil {
			util.Log().Warning("无法删除文件，%s", err)
			retErr = err
			deleteFailed = append(deleteFailed, value)
		}

		// 尝试删除文件的缩略图（如果有）
		_ = os.Remove(value + "._thumb")
	}

	return deleteFailed, retErr
}
