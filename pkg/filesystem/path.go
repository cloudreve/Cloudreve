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

// CreateDirectory 在`base`路径下创建名为`dir`的目录
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
