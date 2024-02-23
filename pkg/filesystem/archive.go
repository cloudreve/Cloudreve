package filesystem

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
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
	"github.com/mholt/archiver/v4"
)

/* ===============
     压缩/解压缩
   ===============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, writer io.Writer, folderIDs, fileIDs []uint, isArchive bool) error {
	// 查找待压缩目录
	folders, err := model.GetFoldersByIDs(folderIDs, fs.User.ID)
	if err != nil && len(folderIDs) != 0 {
		return ErrDBListObjects
	}

	// 查找待压缩文件
	files, err := model.GetFilesByIDs(fileIDs, fs.User.ID)
	if err != nil && len(fileIDs) != 0 {
		return ErrDBListObjects
	}

	// 如果上下文限制了父目录，则进行检查
	if parent, ok := ctx.Value(fsctx.LimitParentCtx).(*model.Folder); ok {
		// 检查目录
		for _, folder := range folders {
			if *folder.ParentID != parent.ID {
				return ErrObjectNotExist
			}
		}

		// 检查文件
		for _, file := range files {
			if file.FolderID != parent.ID {
				return ErrObjectNotExist
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

	// 创建压缩文件Writer
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	ctx = reqContext

	// 压缩各个目录及文件
	for i := 0; i < len(folders); i++ {
		select {
		case <-reqContext.Done():
			// 取消压缩请求
			return ErrClientCanceled
		default:
			fs.doCompress(reqContext, nil, &folders[i], zipWriter, isArchive)
		}

	}
	for i := 0; i < len(files); i++ {
		select {
		case <-reqContext.Done():
			// 取消压缩请求
			return ErrClientCanceled
		default:
			fs.doCompress(reqContext, &files[i], nil, zipWriter, isArchive)
		}
	}

	return nil
}

func (fs *FileSystem) doCompress(ctx context.Context, file *model.File, folder *model.Folder, zipWriter *zip.Writer, isArchive bool) {
	// 如果对象是文件
	if file != nil {
		// 切换上传策略
		fs.Policy = file.GetPolicy()
		err := fs.DispatchHandler()
		if err != nil {
			util.Log().Warning("Failed to compress file %q: %s", file.Name, err)
			return
		}

		// 获取文件内容
		fileToZip, err := fs.Handler.Get(
			context.WithValue(ctx, fsctx.FileModelCtx, *file),
			file.SourceName,
		)
		if err != nil {
			util.Log().Debug("Failed to open %q: %s", file.Name, err)
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
func (fs *FileSystem) Decompress(ctx context.Context, src, dst, encoding string) error {
	err := fs.ResetFileIfNotExist(ctx, src)
	if err != nil {
		return err
	}

	tempZipFilePath := ""
	defer func() {
		// 结束时删除临时压缩文件
		if tempZipFilePath != "" {
			if err := os.Remove(tempZipFilePath); err != nil {
				util.Log().Warning("Failed to delete temp archive file %q: %s", tempZipFilePath, err)
			}
		}
	}()

	// 下载压缩文件到临时目录
	fileStream, err := fs.Handler.Get(ctx, fs.FileTarget[0].SourceName)
	if err != nil {
		return err
	}

	defer fileStream.Close()

	tempZipFilePath = filepath.Join(
		util.RelativePath(model.GetSettingByName("temp_path")),
		"decompress",
		fmt.Sprintf("archive_%d.zip", time.Now().UnixNano()),
	)

	zipFile, err := util.CreatNestedFile(tempZipFilePath)
	if err != nil {
		util.Log().Warning("Failed to create temp archive file %q: %s", tempZipFilePath, err)
		tempZipFilePath = ""
		return err
	}
	defer zipFile.Close()

	// 下载前先判断是否是可解压的格式
	format, readStream, err := archiver.Identify(fs.FileTarget[0].SourceName, fileStream)
	if err != nil {
		util.Log().Warning("Failed to detect compressed format of file %q: %s", fs.FileTarget[0].SourceName, err)
		return err
	}

	extractor, ok := format.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("file not an extractor %s", fs.FileTarget[0].SourceName)
	}

	// 只有zip格式可以多个文件同时上传
	var isZip bool
	switch extractor.(type) {
	case archiver.Zip:
		extractor = archiver.Zip{TextEncoding: encoding}
		isZip = true
	}

	// 除了zip必须下载到本地，其余的可以边下载边解压
	reader := readStream
	if isZip {
		_, err = io.Copy(zipFile, readStream)
		if err != nil {
			util.Log().Warning("Failed to write temp archive file %q: %s", tempZipFilePath, err)
			return err
		}

		fileStream.Close()

		// 设置文件偏移量
		zipFile.Seek(0, io.SeekStart)
		reader = zipFile
	}

	var wg sync.WaitGroup
	parallel := model.GetIntSetting("max_parallel_transfer", 4)
	worker := make(chan int, parallel)
	for i := 0; i < parallel; i++ {
		worker <- i
	}

	// 上传文件函数
	uploadFunc := func(fileStream io.ReadCloser, size int64, savePath, rawPath string) {
		defer func() {
			if isZip {
				worker <- 1
				wg.Done()
			}
			if err := recover(); err != nil {
				util.Log().Warning("Error while uploading files inside of archive file.")
				fmt.Println(err)
			}
		}()

		err := fs.UploadFromStream(ctx, &fsctx.FileStream{
			File:        fileStream,
			Size:        uint64(size),
			Name:        path.Base(savePath),
			VirtualPath: path.Dir(savePath),
		}, true)
		fileStream.Close()
		if err != nil {
			util.Log().Debug("Failed to upload file %q in archive file: %s, skipping...", rawPath, err)
		}
	}

	// 解压缩文件，回调函数如果出错会停止解压的下一步进行，全部return nil
	err = extractor.Extract(ctx, reader, nil, func(ctx context.Context, f archiver.File) error {
		rawPath := util.FormSlash(f.NameInArchive)
		savePath := path.Join(dst, rawPath)
		// 路径是否合法
		if !strings.HasPrefix(savePath, util.FillSlash(path.Clean(dst))) {
			util.Log().Warning("%s: illegal file path", f.NameInArchive)
			return nil
		}

		// 如果是目录
		if f.FileInfo.IsDir() {
			fs.CreateDirectory(ctx, savePath)
			return nil
		}

		// 上传文件
		fileStream, err := f.Open()
		if err != nil {
			util.Log().Warning("Failed to open file %q in archive file: %s, skipping...", rawPath, err)
			return nil
		}

		if !isZip {
			uploadFunc(fileStream, f.FileInfo.Size(), savePath, rawPath)
		} else {
			<-worker
			wg.Add(1)
			go uploadFunc(fileStream, f.FileInfo.Size(), savePath, rawPath)
		}
		return nil
	})
	wg.Wait()
	return err

}
