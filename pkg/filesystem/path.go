package filesystem

import (
	model "github.com/HFO4/cloudreve/models"
	"path"
)

/* =================
	 路径/目录相关
   =================
*/

// IsPathExist 返回给定目录是否存在
// 如果存在就返回目录
func (fs *FileSystem) IsPathExist(path string) (bool, model.Folder) {
	folder, err := model.GetFolderByPath(path, fs.User.ID)
	return err == nil, folder
}

// IsFileExist 返回给定路径的文件是否存在
func (fs *FileSystem) IsFileExist(fullPath string) bool {
	basePath := path.Dir(fullPath)
	fileName := path.Base(fullPath)

	_, err := model.GetFileByPathAndName(basePath, fileName, fs.User.ID)

	return err == nil
}
