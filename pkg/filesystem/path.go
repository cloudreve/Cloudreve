package filesystem

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"path"
)

/* =================
	 路径/目录相关
   =================
*/

// Object 文件或者目录
type Object struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Pic  string `json:"pic"`
	Size uint64 `json:"size"`
	Type string `json:"type"`
	Date string `json:"date"`
}

// List 列出路径下的内容,
// pathProcessor为最终对象路径的处理钩子。
// 有些情况下（如在分享页面列对象）时，
// 路径需要截取掉被分享目录路径之前的部分。
func (fs *FileSystem) List(ctx context.Context, path string, pathProcessor func(string) string) ([]Object, error) {
	// 获取父目录
	isExist, folder := fs.IsPathExist(path)
	// 不存在时返回空的结果
	if !isExist {
		return nil, nil
	}

	// 获取子目录
	childFolders, _ := folder.GetChildFolder()
	// 获取子文件
	childFiles, _ := folder.GetChildFile()

	// 汇总处理结果
	objects := make([]Object, 0, len(childFiles)+len(childFolders))
	// 所有对象的父目录
	var processedPath string

	for _, folder := range childFolders {
		// 路径处理钩子，
		// 所有对象父目录都是一样的，所以只处理一次
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(folder.Position)
			} else {
				processedPath = folder.Position
			}
		}

		objects = append(objects, Object{
			ID:   folder.ID,
			Name: folder.Name,
			Path: processedPath,
			Pic:  "",
			Size: 0,
			Type: "dir",
			Date: folder.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	for _, file := range childFiles {
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(file.Dir)
			} else {
				processedPath = file.Dir
			}
		}

		objects = append(objects, Object{
			ID:   file.ID,
			Name: file.Name,
			Path: processedPath,
			Pic:  file.PicInfo,
			Size: file.Size,
			Type: "file",
			Date: file.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return objects, nil
}

// CreateDirectory 根据给定的完整创建目录，不会递归创建
func (fs *FileSystem) CreateDirectory(ctx context.Context, fullPath string) error {
	// 获取要创建目录的父路径和目录名
	fullPath = path.Clean(fullPath)
	base := path.Dir(fullPath)
	dir := path.Base(fullPath)

	// 检查目录名是否合法
	if !fs.ValidateLegalName(ctx, dir) {
		return ErrIllegalObjectName
	}

	// 父目录是否存在
	isExist, parent := fs.IsPathExist(base)
	if !isExist {
		return ErrPathNotExist
	}

	// 是否有同名文件
	if fs.IsFileExist(path.Join(base, dir)) {
		return ErrFileExisted
	}

	// 创建目录
	newFolder := model.Folder{
		Name:             dir,
		ParentID:         parent.ID,
		Position:         base,
		OwnerID:          fs.User.ID,
		PositionAbsolute: fullPath,
	}
	_, err := newFolder.Create()

	return err
}

// IsPathExist 返回给定目录是否存在
// 如果存在就返回目录
func (fs *FileSystem) IsPathExist(path string) (bool, *model.Folder) {
	folder, err := model.GetFolderByPath(path, fs.User.ID)
	return err == nil, &folder
}

// IsFileExist 返回给定路径的文件是否存在
func (fs *FileSystem) IsFileExist(fullPath string) bool {
	basePath := path.Dir(fullPath)
	fileName := path.Base(fullPath)

	_, err := model.GetFileByPathAndName(basePath, fileName, fs.User.ID)

	return err == nil
}
