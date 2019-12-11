package local

import (
	"context"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// Handler 本地策略适配器
type Handler struct {
	Policy *model.Policy
}

// Get 获取文件内容
func (handler Handler) Get(ctx context.Context, path string) (io.ReadSeeker, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		util.Log().Debug("无法打开文件：%s", err)
		return nil, err
	}

	// 开启一个协程，用于请求结束后关闭reader
	// go closeReader(ctx, file)

	return file, nil
}

// closeReader 用于在请求结束后关闭reader
// TODO 让业务代码自己关闭
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
		_ = os.Remove(value + conf.ThumbConfig.FileSuffix)
	}

	return deleteFailed, retErr
}

// Thumb 获取文件缩略图
func (handler Handler) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	file, err := handler.Get(ctx, path+conf.ThumbConfig.FileSuffix)
	if err != nil {
		return nil, err
	}

	return &response.ContentResponse{
		Redirect: false,
		Content:  file,
	}, nil
}

// Source 获取外链URL
func (handler Handler) Source(ctx context.Context, path string, url url.URL, expires int64) (string, error) {
	file, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return "", errors.New("无法获取文件记录上下文")
	}

	// 签名生成文件记录
	signedURI, err := auth.SignURI(
		fmt.Sprintf("/api/v3/file/get/%d/%s", file.ID, file.Name),
		0,
	)
	if err != nil {
		return "", serializer.NewError(serializer.CodeEncryptError, "无法对URL进行签名", err)
	}

	finalURL := url.ResolveReference(signedURI).String()
	return finalURL, nil
}
