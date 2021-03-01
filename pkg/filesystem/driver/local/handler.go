package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

// Driver 本地策略适配器
type Driver struct {
	Policy *model.Policy
}

// List 递归列取给定物理路径下所有文件
func (handler Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	var res []response.Object

	// 取得起始路径
	root := util.RelativePath(filepath.FromSlash(path))

	// 开始遍历路径下的文件、目录
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			// 跳过根目录
			if path == root {
				return nil
			}

			if err != nil {
				util.Log().Warning("无法遍历目录 %s, %s", path, err)
				return filepath.SkipDir
			}

			// 将遍历对象的绝对路径转换为相对路径
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			res = append(res, response.Object{
				Name:         info.Name(),
				RelativePath: filepath.ToSlash(rel),
				Source:       path,
				Size:         uint64(info.Size()),
				IsDir:        info.IsDir(),
				LastModify:   info.ModTime(),
			})

			// 如果非递归，则不步入目录
			if !recursive && info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		})

	return res, err
}

// Get 获取文件内容
func (handler Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	// 打开文件
	file, err := os.Open(util.RelativePath(path))
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
func (handler Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	defer file.Close()
	dst = util.RelativePath(filepath.FromSlash(dst))

	// 如果禁止了 Overwrite，则检查是否有重名冲突
	if ctx.Value(fsctx.DisableOverwrite) != nil {
		if util.Exists(dst) {
			util.Log().Warning("物理同名文件已存在或不可用: %s", dst)
			return errors.New("物理同名文件已存在或不可用")
		}
	}

	// 如果目标目录不存在，创建
	basePath := filepath.Dir(dst)
	if !util.Exists(basePath) {
		err := os.MkdirAll(basePath, 0744)
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
func (handler Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	deleteFailed := make([]string, 0, len(files))
	var retErr error

	for _, value := range files {
		filePath := util.RelativePath(filepath.FromSlash(value))
		if util.Exists(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				util.Log().Warning("无法删除文件，%s", err)
				retErr = err
				deleteFailed = append(deleteFailed, value)
			}
		}

		// 尝试删除文件的缩略图（如果有）
		_ = os.Remove(util.RelativePath(value + conf.ThumbConfig.FileSuffix))
	}

	return deleteFailed, retErr
}

// Thumb 获取文件缩略图
func (handler Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
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
func (handler Driver) Source(
	ctx context.Context,
	path string,
	baseURL url.URL,
	ttl int64,
	isDownload bool,
	speed int,
) (string, error) {
	file, ok := ctx.Value(fsctx.FileModelCtx).(model.File)
	if !ok {
		return "", errors.New("无法获取文件记录上下文")
	}

	// 是否启用了CDN
	if handler.Policy.BaseURL != "" {
		cdnURL, err := url.Parse(handler.Policy.BaseURL)
		if err != nil {
			return "", err
		}
		baseURL = *cdnURL
	}

	var (
		signedURI *url.URL
		err       error
	)
	if isDownload {
		// 创建下载会话，将文件信息写入缓存
		downloadSessionID := util.RandStringRunes(16)
		err = cache.Set("download_"+downloadSessionID, file, int(ttl))
		if err != nil {
			return "", serializer.NewError(serializer.CodeCacheOperation, "无法创建下载会话", err)
		}

		// 签名生成文件记录
		signedURI, err = auth.SignURI(
			auth.General,
			fmt.Sprintf("/api/v3/file/download/%s", downloadSessionID),
			ttl,
		)
	} else {
		// 签名生成文件记录
		signedURI, err = auth.SignURI(
			auth.General,
			fmt.Sprintf("/api/v3/file/get/%d/%s", file.ID, file.Name),
			ttl,
		)
	}

	if err != nil {
		return "", serializer.NewError(serializer.CodeEncryptError, "无法对URL进行签名", err)
	}

	finalURL := baseURL.ResolveReference(signedURI).String()
	return finalURL, nil
}

// Token 获取上传策略和认证Token，本地策略直接返回空值
func (handler Driver) Token(ctx context.Context, ttl int64, key string) (serializer.UploadCredential, error) {
	return serializer.UploadCredential{}, nil
}
