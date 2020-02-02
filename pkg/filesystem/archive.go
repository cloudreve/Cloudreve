package filesystem

import (
	"archive/zip"
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path/filepath"
	"time"
)

/* ===============
     压缩/解压缩
   ===============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, folderIDs, fileIDs []uint, isArchive bool) (string, error) {
	// 查找待压缩目录
	folders, err := model.GetFoldersByIDs(folderIDs, fs.User.ID)
	if err != nil && len(folders) != 0 {
		return "", ErrDBListObjects
	}

	// 查找待压缩文件
	files, err := model.GetFilesByIDs(fileIDs, fs.User.ID)
	if err != nil && len(files) != 0 {
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
		model.GetSettingByName("temp_path"),
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
			Name:               filepath.Join(file.Position, file.Name),
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
