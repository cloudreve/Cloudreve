package filesystem

import (
	"archive/zip"
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"io"
	"net/http"
	"os"
)

/* ===============
     压缩/解压缩
   ===============
*/

// Compress 创建给定目录和文件的压缩文件
func (fs *FileSystem) Compress(ctx context.Context, folderIDs, fileIDs []uint, writer http.ResponseWriter) error {
	// 查找待压缩目录
	folders, err := model.GetFoldersByIDs(folderIDs, fs.User.ID)
	if err != nil && len(folders) != 0 {
		return ErrDBListObjects
	}
	// 查找待压缩文件
	files, err := model.GetFilesByIDs(fileIDs, fs.User.ID)
	if err != nil && len(files) != 0 {
		return ErrDBListObjects
	}

	// 将顶级待处理对象的路径设为根路径
	for i := 0; i < len(folders); i++ {
		folders[i].Position = ""
	}
	for i := 0; i < len(files); i++ {
		files[i].Position = ""
	}

	// 创建压缩文件
	zipFile, err := os.Create("temp/archive.zip")
	if err != nil {
		util.Log().Warning("%s", err)
		return err
	}
	defer zipFile.Close()

	//writer.Header().Set("Content-Type", "application/zip")
	//writer.Header().Set("Content-Disposition", "attachment; filename=\"archive.zip\"")

	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	ctx, _ = context.WithCancel(context.Background())
	// 压缩各个目录及文件
	for i := 0; i < len(folders); i++ {
		fs.doCompress(ctx, nil, &folders[i], zipWriter)
	}
	for i := 0; i < len(files); i++ {
		fs.doCompress(ctx, &files[i], nil, zipWriter)
	}

	return nil
}

func (fs *FileSystem) doCompress(ctx context.Context, file *model.File, folder *model.Folder, zipWriter *zip.Writer) {
	// 如果对象是文件
	if file != nil {
		// 切换上传策略
		fs.Policy = file.GetPolicy()
		err := fs.dispatchHandler()
		if err != nil {
			util.Log().Warning("无法压缩文件%s，%s", file.Name, err)
			return
		}

		fileToZip, err := fs.Handler.Get(ctx, file.SourceName)
		if err != nil {
			util.Log().Debug("Open%s，%s", file.Name, err)
			return
		}
		if closer, ok := fileToZip.(io.Closer); ok {
			defer closer.Close()
		}

		header := &zip.FileHeader{
			Name: file.Name,
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return
		}

		_, err = io.Copy(writer, fileToZip)
		util.Log().Debug("开始压缩%s", file.Name)

		//f, err := os.Open(file.SourceName)
		//if err != nil {
		//	util.Log().Debug("Open%s，%s", file.Name, err)
		//	return
		//}
		//
		//info, err := f.Stat()
		//if err != nil {
		//	util.Log().Debug("Stat%s，%s", file.Name, err)
		//	return
		//}
		//
		//header, err := zip.FileInfoHeader(info)
		//if err != nil {
		//	util.Log().Debug("header%s，%s", file.Name, err)
		//	return
		//}
		//
		//writer, err := zipWriter.CreateHeader(header)
		//if err != nil {
		//	util.Log().Debug("压缩出错%s，%s", header.Name, err)
		//	return
		//}
		//
		//_, err = io.Copy(writer, f)
		//if err != nil {
		//	util.Log().Debug("压缩出错%s，%s", header.Name, err)
		//	return
		//}
		//
		//if true {
		//	f.Close()
		//}
		//
		//util.Log().Debug("压缩完成%s", header.Name)

		//pool.Submit(compressJob)
	}
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
