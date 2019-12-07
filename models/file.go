package model

import (
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"path"
)

// File 文件
type File struct {
	// 表字段
	gorm.Model
	Name       string `gorm:"unique_index:idx_only_one"`
	SourceName string
	UserID     uint `gorm:"index:user_id;unique_index:idx_only_one"`
	Size       uint64
	PicInfo    string
	FolderID   uint `gorm:"index:folder_id;unique_index:idx_only_one"`
	PolicyID   uint
	Dir        string `gorm:"size:65536"`

	// 关联模型
	Policy Policy `gorm:"PRELOAD:false,association_autoupdate:false"`
}

// Create 创建文件记录
func (file *File) Create() (uint, error) {
	if err := DB.Create(file).Error; err != nil {
		util.Log().Warning("无法插入文件记录, %s", err)
		return 0, err
	}
	return file.ID, nil
}

// GetFileByPathAndName 给定路径(s)、文件名、用户ID，查找文件
func GetFileByPathAndName(path string, name string, uid uint) (File, error) {
	var file File
	result := DB.Where("user_id = ? AND dir = ? AND name=?", uid, path, name).First(&file)
	return file, result.Error
}

// GetChildFile 查找目录下子文件
func (folder *Folder) GetChildFile() ([]File, error) {
	var files []File
	result := DB.Where("folder_id = ?", folder.ID).Find(&files)
	return files, result.Error
}

// GetChildFilesOfFolders 批量检索目录子文件
func GetChildFilesOfFolders(folders *[]Folder) ([]File, error) {
	// 将所有待删除目录ID抽离，以便检索文件
	folderIDs := make([]uint, 0, len(*folders))
	for _, value := range *folders {
		folderIDs = append(folderIDs, value.ID)
	}

	// 检索文件
	var files []File
	result := DB.Where("folder_id in (?)", folderIDs).Find(&files)
	return files, result.Error
}

// GetPolicy 获取文件所属策略
func (file *File) GetPolicy() *Policy {
	if file.Policy.Model.ID == 0 {
		file.Policy, _ = GetPolicyByID(file.PolicyID)
	}
	return &file.Policy
}

// GetFileByPaths 根据给定的文件路径(s)查找文件
func GetFileByPaths(paths []string, uid uint) ([]File, error) {
	var files []File
	tx := DB
	for _, value := range paths {
		base := path.Base(value)
		dir := path.Dir(value)
		tx = tx.Or("dir = ? and name = ? and user_id = ?", dir, base, uid)
	}
	result := tx.Find(&files)
	return files, result.Error
}

// RemoveFilesWithSoftLinks 去除给定的文件列表中有软链接的文件
func RemoveFilesWithSoftLinks(files []File) ([]File, error) {
	// 结果值
	filteredFiles := make([]File, 0)

	// 查询软链接的文件
	var filesWithSoftLinks []File
	tx := DB
	for _, value := range files {
		tx = tx.Or("source_name = ? and policy_id = ? and id != ?", value.SourceName, value.PolicyID, value.ID)
	}
	result := tx.Find(&filesWithSoftLinks)
	if result.Error != nil {
		return nil, result.Error
	}

	// 过滤具有软连接的文件
	// TODO: 优化复杂度
	if len(filesWithSoftLinks) == 0 {
		filteredFiles = files
	} else {
		for i := 0; i < len(files); i++ {
			finder := false
			for _, value := range filesWithSoftLinks {
				if value.PolicyID == files[i].PolicyID && value.SourceName == files[i].SourceName {
					finder = true
					break
				}
			}
			if !finder {
				filteredFiles = append(filteredFiles, files[i])
			}

		}
	}

	return filteredFiles, nil

}

// DeleteFileByIDs 根据给定ID批量删除文件记录
func DeleteFileByIDs(ids []uint) error {
	result := DB.Where("id in (?)", ids).Unscoped().Delete(&File{})
	return result.Error
}

//// GetRecursiveByPaths 根据给定的文件路径(s)递归查找文件
//func GetRecursiveByPaths(paths []string, uid uint) ([]File, error) {
//	files := make([]File, 0, len(paths))
//	search := util.BuildRegexp(paths, "^", "/", "|")
//	result := DB.Where("(user_id = ? and dir REGEXP ?) or (user_id = ？ and dir in (?))", uid, search, uid, paths).Find(&files)
//	return files, result.Error
//}

// GetFilesByParentIDs 根据父目录ID查找文件
func GetFilesByParentIDs(ids []uint, uid uint) ([]File, error) {
	files := make([]File, 0, len(ids))
	result := DB.Where("user_id = ? and folder_id in (?)", uid, ids).Find(&files)
	return files, result.Error
}

// Rename 重命名文件
func (file *File) Rename(new string) error {
	return DB.Model(&file).Update("name", new).Error
}
