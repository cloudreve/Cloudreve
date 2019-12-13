package filesystem

import (
	"archive/zip"
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"path/filepath"
	"time"
)

/* ===============
     压缩/解压缩
   ===============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, folderIDs, fileIDs []uint) (string, error) {
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

	// 将顶级待处理对象的路径设为根路径
	for i := 0; i < len(folders); i++ {
		folders[i].Position = ""
	}
	for i := 0; i < len(files); i++ {
		files[i].Position = ""
	}

	// 创建临时压缩文件
	zipFilePath := filepath.Join(model.GetSettingByName("temp_path"), fmt.Sprintf("archive_%d.zip", time.Now().UnixNano()))
	zipFile, err := util.CreatNestedFile(zipFilePath)
	defer zipFile.Close()
	if err != nil {
		util.Log().Warning("%s", err)
		return "", err
	}

	// 创建压缩文件Writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	ctx, _ = context.WithCancel(context.Background())
	// 压缩各个目录及文件
	for i := 0; i < len(folders); i++ {
		fs.doCompress(ctx, nil, &folders[i], zipWriter, true)
	}
	for i := 0; i < len(files); i++ {
		fs.doCompress(ctx, &files[i], nil, zipWriter, true)
	}

	return zipFilePath, nil
}

func (fs *FileSystem) doCompress(ctx context.Context, file *model.File, folder *model.Folder, zipWriter *zip.Writer, isArchive bool) {
	// 如果对象是文件
	if file != nil {
		// 切换上传策略
		fs.Policy = file.GetPolicy()
		err := fs.dispatchHandler()
		if err != nil {
			util.Log().Warning("无法压缩文件%s，%s", file.Name, err)
			return
		}

		// 获取文件内容
		fileToZip, err := fs.Handler.Get(ctx, file.SourceName)
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
