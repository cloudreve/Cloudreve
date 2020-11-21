package filesystem

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

/* ===============
     压缩/解压缩
   ===============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, folderIDs, fileIDs []uint, isArchive bool) (string, error) {
	// 查找待压缩目录
	folders, err := model.GetFoldersByIDs(folderIDs, fs.User.ID)
	if err != nil && len(folderIDs) != 0 {
		return "", ErrDBListObjects
	}

	// 查找待压缩文件
	files, err := model.GetFilesByIDs(fileIDs, fs.User.ID)
	if err != nil && len(fileIDs) != 0 {
		return "", ErrDBListObjects
	}

	// 如果上下文限制了父目录，则进行检查
	if parent, ok := ctx.Value(fsctx.LimitParentCtx).(*model.Folder); ok {
		// 检查目录
		for _, folder := range folders {
			if *folder.ParentID != parent.ID {
				return "", ErrObjectNotExist
			}
		}

		// 检查文件
		for _, file := range files {
			if file.FolderID != parent.ID {
				return "", ErrObjectNotExist
			}
		}
	}

	// 尝试获取请求上下文，以便于后续检查用户取消任务
	reqContext := ctx
	ginCtx, ok := ctx.Value(fsctx.GinCtx).(*gin.Context)
	if ok {
		reqContext = ginCtx.Request.Context()
	}

	// 将顶级待处理对象的路径设为根路径
	for i := 0; i < len(folders); i++ {
		folders[i].Position = ""
	}
	for i := 0; i < len(files); i++ {
		files[i].Position = ""
	}

	// 创建临时压缩文件
	saveFolder := "archive"
	if !isArchive {
		saveFolder = "compress"
	}
	zipFilePath := filepath.Join(
		util.RelativePath(model.GetSettingByName("temp_path")),
		saveFolder,
		fmt.Sprintf("archive_%d.zip", time.Now().UnixNano()),
	)
	zipFile, err := util.CreatNestedFile(zipFilePath)
	if err != nil {
		util.Log().Warning("%s", err)
		return "", err
	}
	defer zipFile.Close()

	// 创建压缩文件Writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	ctx = reqContext

	// 压缩各个目录及文件
	for i := 0; i < len(folders); i++ {
		select {
		case <-reqContext.Done():
			// 取消压缩请求
			fs.cancelCompress(ctx, zipWriter, zipFile, zipFilePath)
			return "", ErrClientCanceled
		default:
			fs.doCompress(ctx, nil, &folders[i], zipWriter, isArchive)
		}

	}
	for i := 0; i < len(files); i++ {
		select {
		case <-reqContext.Done():
			// 取消压缩请求
			fs.cancelCompress(ctx, zipWriter, zipFile, zipFilePath)
			return "", ErrClientCanceled
		default:
			fs.doCompress(ctx, &files[i], nil, zipWriter, isArchive)
		}
	}

	return zipFilePath, nil
}

// cancelCompress 取消压缩进程
func (fs *FileSystem) cancelCompress(ctx context.Context, zipWriter *zip.Writer, file *os.File, path string) {
	util.Log().Debug("客户端取消压缩请求")
	zipWriter.Close()
	file.Close()
	_ = os.Remove(path)
}

func (fs *FileSystem) doCompress(ctx context.Context, file *model.File, folder *model.Folder, zipWriter *zip.Writer, isArchive bool) {
	// 如果对象是文件
	if file != nil {
		// 切换上传策略
		fs.Policy = file.GetPolicy()
		err := fs.DispatchHandler()
		if err != nil {
			util.Log().Warning("无法压缩文件%s，%s", file.Name, err)
			return
		}

		// 获取文件内容
		fileToZip, err := fs.Handler.Get(
			context.WithValue(ctx, fsctx.FileModelCtx, *file),
			file.SourceName,
		)
		if err != nil {
			util.Log().Debug("Open%s，%s", file.Name, err)
			return
		}
		if closer, ok := fileToZip.(io.Closer); ok {
			defer closer.Close()
		}

		// 创建压缩文件头
		header := &zip.FileHeader{
			Name:               filepath.FromSlash(path.Join(file.Position, file.Name)),
			Modified:           file.UpdatedAt,
			UncompressedSize64: file.Size,
		}

		// 指定是压缩还是归档
		if isArchive {
			header.Method = zip.Store
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return
		}

		_, err = io.Copy(writer, fileToZip)
	} else if folder != nil {
		// 对象是目录
		// 获取子文件
		subFiles, err := folder.GetChildFiles()
		if err == nil && len(subFiles) > 0 {
			for i := 0; i < len(subFiles); i++ {
				fs.doCompress(ctx, &subFiles[i], nil, zipWriter, isArchive)
			}

		}
		// 获取子目录，继续递归遍历
		subFolders, err := folder.GetChildFolder()
		if err == nil && len(subFolders) > 0 {
			for i := 0; i < len(subFolders); i++ {
				fs.doCompress(ctx, nil, &subFolders[i], zipWriter, isArchive)
			}
		}
	}
}

// Decompress 解压缩给定压缩文件到dst目录
func (fs *FileSystem) Decompress(ctx context.Context, src, dst string) error {
	err := fs.ResetFileIfNotExist(ctx, src)
	if err != nil {
		return err
	}

	tempZipFilePath := ""
	defer func() {
		// 结束时删除临时压缩文件
		if tempZipFilePath != "" {
			if err := os.Remove(tempZipFilePath); err != nil {
				util.Log().Warning("无法删除临时压缩文件 %s , %s", tempZipFilePath, err)
			}
		}
	}()

	// 下载压缩文件到临时目录
	fileStream, err := fs.Handler.Get(ctx, fs.FileTarget[0].SourceName)
	if err != nil {
		return err
	}

	tempZipFilePath = filepath.Join(
		util.RelativePath(model.GetSettingByName("temp_path")),
		"decompress",
		fmt.Sprintf("archive_%d.zip", time.Now().UnixNano()),
	)

	zipFile, err := util.CreatNestedFile(tempZipFilePath)
	if err != nil {
		util.Log().Warning("无法创建临时压缩文件 %s , %s", tempZipFilePath, err)
		tempZipFilePath = ""
		return err
	}
	defer zipFile.Close()

	_, err = io.Copy(zipFile, fileStream)
	if err != nil {
		util.Log().Warning("无法写入临时压缩文件 %s , %s", tempZipFilePath, err)
		return err
	}

	zipFile.Close()

	// 解压缩文件
	r, err := zip.OpenReader(tempZipFilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 重设存储策略
	fs.Policy = &fs.User.Policy
	err = fs.DispatchHandler()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	parallel := model.GetIntSetting("max_parallel_transfer", 4)
	worker := make(chan int, parallel)
	for i := 0; i < parallel; i++ {
		worker <- i
	}

	for _, f := range r.File {
		fileName := f.Name
		// 处理非UTF-8编码
		if f.NonUTF8 {
			i := bytes.NewReader([]byte(fileName))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := ioutil.ReadAll(decoder)
			fileName = string(content)
		}

		rawPath := util.FormSlash(fileName)
		savePath := path.Join(dst, rawPath)
		// 路径是否合法
		if !strings.HasPrefix(savePath, util.FillSlash(path.Clean(dst))) {
			return fmt.Errorf("%s: illegal file path", f.Name)
		}

		// 如果是目录
		if f.FileInfo().IsDir() {
			fs.CreateDirectory(ctx, savePath)
			continue
		}

		// 上传文件
		fileStream, err := f.Open()
		if err != nil {
			util.Log().Warning("无法打开压缩包内文件%s , %s , 跳过", rawPath, err)
			continue
		}

		select {
		case <-worker:
			wg.Add(1)
			go func(fileStream io.ReadCloser, size int64) {
				defer func() {
					worker <- 1
					wg.Done()
					if err := recover(); err != nil {
						util.Log().Warning("上传压缩包内文件时出错")
						fmt.Println(err)
					}
				}()

				err = fs.UploadFromStream(ctx, fileStream, savePath, uint64(size))
				fileStream.Close()
				if err != nil {
					util.Log().Debug("无法上传压缩包内的文件%s , %s , 跳过", rawPath, err)
				}
			}(fileStream, f.FileInfo().Size())
		}

	}
	wg.Wait()
	return nil

}
