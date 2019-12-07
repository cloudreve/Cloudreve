package filesystem

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"path"
)

/* =================
	 路径/目录相关
   =================
*/

// IsPathExist 返回给定目录是否存在
// 如果存在就返回目录
func (fs *FileSystem) IsPathExist(path string) (bool, *model.Folder) {
	pathList := util.SplitPath(path)
	if len(pathList) == 0 {
		return false, nil
	}

	// 递归步入目录
	var currentFolder *model.Folder
	for _, folderName := range pathList {
		var err error

		// 根目录
		if folderName == "/" {
			currentFolder, err = fs.User.Root()
			if err != nil {
				return false, nil
			}
		} else {
			currentFolder, err = currentFolder.GetChild(folderName)
			if err != nil {
				return false, nil
			}
		}
	}

	return true, currentFolder
}

// IsFileExist 返回给定路径的文件是否存在
func (fs *FileSystem) IsFileExist(fullPath string) (bool, *model.File) {
	basePath := path.Dir(fullPath)
	fileName := path.Base(fullPath)

	// 获得父目录
	exist, parent := fs.IsPathExist(basePath)
	if !exist {
		return false, nil
	}

	file, err := parent.GetChildFile(fileName)

	return err == nil, file
}

// IsChildFileExist 确定folder目录下是否有名为name的文件
func (fs *FileSystem) IsChildFileExist(folder *model.Folder, name string) (bool, *model.File) {
	file, err := folder.GetChildFile(name)
	return err == nil, file
}
