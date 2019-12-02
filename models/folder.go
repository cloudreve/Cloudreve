package model

import (
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/jinzhu/gorm"
	"path"
	"strings"
)

// Folder 目录
type Folder struct {
	// 表字段
	gorm.Model
	Name             string `gorm:"unique_index:idx_only_one_name"`
	ParentID         uint   `gorm:"index:parent_id;unique_index:idx_only_one_name"`
	Position         string `gorm:"size:65536"`
	OwnerID          uint   `gorm:"index:owner_id"`
	PositionAbsolute string `gorm:"size:65536"`
}

// Create 创建目录
func (folder *Folder) Create() (uint, error) {
	if err := DB.Create(folder).Error; err != nil {
		util.Log().Warning("无法插入目录记录, %s", err)
		return 0, err
	}
	return folder.ID, nil
}

// GetFolderByPath 根据绝对路径和UID查找目录
func GetFolderByPath(path string, uid uint) (Folder, error) {
	var folder Folder
	result := DB.Where("owner_id = ? AND position_absolute = ?", uid, path).First(&folder)
	return folder, result.Error
}

// GetChildFolder 查找子目录
func (folder *Folder) GetChildFolder() ([]Folder, error) {
	var folders []Folder
	result := DB.Where("parent_id = ?", folder.ID).Find(&folders)
	return folders, result.Error
}

// GetRecursiveChildFolder 查找所有递归子目录，包括自身
func GetRecursiveChildFolder(dirs []string, uid uint, includeSelf bool) ([]Folder, error) {
	folders := make([]Folder, 0, len(dirs))
	var err error
	if conf.DatabaseConfig.Type == "mysql" {

		// MySQL 下使用正则查询
		search := util.BuildRegexp(dirs, "^", "/", "|")
		result := DB.Where("(owner_id = ? and position_absolute REGEXP ?) or (owner_id = ? and position_absolute in (?))", uid, search, uid, dirs).Find(&folders)
		err = result.Error

	} else {

		// SQLite 下使用递归查询
		var parFolders []Folder
		result := DB.Where("owner_id = ? and position_absolute in (?)", uid, dirs).Find(&parFolders)
		if result.Error != nil {
			return folders, err
		}

		// 整理父目录的ID
		var parentIDs = make([]uint, 0, len(parFolders))
		for _, folder := range parFolders {
			parentIDs = append(parentIDs, folder.ID)
		}

		if includeSelf {
			// 合并至最终结果
			folders = append(folders, parFolders...)
		}
		parFolders = []Folder{}

		// 递归查询子目录,最大递归65535次
		for i := 0; i < 65535; i++ {

			result = DB.Where("owner_id = ? and parent_id in (?)", uid, parentIDs).Find(&parFolders)

			// 查询结束条件
			if len(parFolders) == 0 {
				break
			}

			// 整理父目录的ID
			parentIDs = make([]uint, 0, len(parFolders))
			for _, folder := range parFolders {
				parentIDs = append(parentIDs, folder.ID)
			}

			// 合并至最终结果
			folders = append(folders, parFolders...)
			parFolders = []Folder{}

		}

	}

	return folders, err
}

// DeleteFolderByIDs 根据给定ID批量删除目录记录
func DeleteFolderByIDs(ids []uint) error {
	result := DB.Where("id in (?)", ids).Unscoped().Delete(&Folder{})
	return result.Error
}

//func (folder *Folder)GetPositionAbsolute()string{
//	return path.Join(folder.Position,folder.Name)
//}

// MoveFileTo 将此目录下的文件移动至dstFolder
func (folder *Folder) MoveFileTo(files []string, dstFolder *Folder) error {
	// 更改顶级要移动文件的父目录指向
	err := DB.Model(File{}).Where("name in (?) and user_id = ? and dir = ?", files, folder.OwnerID, folder.PositionAbsolute).
		Update(map[string]interface{}{
			"folder_id": dstFolder.ID,
			"dir":       dstFolder.PositionAbsolute,
		}).
		Error
	if err != nil {
		return err
	}
	return nil

}

// MoveFolderTo 将此目录下的目录移动至dstFolder
func (folder *Folder) MoveFolderTo(dirs []string, dstFolder *Folder) error {
	// 生成绝对路径
	fullDirs := make([]string, len(dirs))
	for i := 0; i < len(dirs); i++ {
		fullDirs[i] = path.Join(
			folder.PositionAbsolute,
			path.Base(dirs[i]),
		)
	}

	var subFolders = make([][]Folder, len(fullDirs))

	// 更新被移动的目录递归的子目录和文件
	for key, parentDir := range fullDirs {
		// 检索被移动的目录的所有子目录
		toBeMoved, err := GetRecursiveChildFolder([]string{parentDir}, folder.OwnerID, true)
		if err != nil {
			return err
		}
		subFolders[key] = toBeMoved
	}

	// 更改顶级要移动目录的父目录指向
	err := DB.Model(Folder{}).Where("position_absolute in (?) and owner_id = ?", fullDirs, folder.OwnerID).
		Update(map[string]interface{}{
			"parent_id":         dstFolder.ID,
			"position":          dstFolder.PositionAbsolute,
			"position_absolute": gorm.Expr(util.BuildConcat("?", "name", conf.DatabaseConfig.Type), util.FillSlash(dstFolder.PositionAbsolute)),
		}).Error
	if err != nil {
		return err
	}

	// 更新被移动的目录递归的子目录和文件
	for parKey, toBeMoved := range subFolders {
		ignorePath := fullDirs[parKey]
		// TODO 找到更好的修改办法

		for _, subFolder := range toBeMoved {
			// 每个分组的第一个目录已经变更指向，直接跳过
			if subFolder.PositionAbsolute != ignorePath {
				newPosition := path.Join(dstFolder.PositionAbsolute, strings.Replace(subFolder.Position, folder.PositionAbsolute, "", 1))
				DB.Model(&subFolder).Updates(map[string]interface{}{
					"position":          newPosition,
					"position_absolute": path.Join(newPosition, subFolder.Name),
				})
			}
		}

		// 抽离所有子目录的ID
		var subFolderIDs = make([]uint, len(toBeMoved))
		for key, subFolder := range toBeMoved {
			subFolderIDs[key] = subFolder.ID
		}

		// 获取子目录下的所有子文件
		toBeMovedFile, err := GetFilesByParentIDs(subFolderIDs, folder.OwnerID)
		if err != nil {
			return err
		}
		for _, subFile := range toBeMovedFile {
			newPosition := path.Join(dstFolder.PositionAbsolute, strings.Replace(subFile.Dir, folder.PositionAbsolute, "", 1))
			DB.Model(&subFile).Updates(map[string]interface{}{
				"dir": newPosition,
			})
		}

	}

	return nil

}
