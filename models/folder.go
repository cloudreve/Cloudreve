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

// GetRecursiveChildFolder 查找所有递归子目录
func GetRecursiveChildFolder(dirs []string, uid uint) ([]Folder, error) {
	folders := make([]Folder, 0, len(dirs))
	search := util.BuildRegexp(dirs, "^", "/", "|")
	result := DB.Where("(owner_id = ? and position_absolute REGEXP ?) or position_absolute in (?)", uid, search, dirs).Find(&folders)
	return folders, result.Error
}

// DeleteFolderByIDs 根据给定ID批量删除目录记录
func DeleteFolderByIDs(ids []uint) error {
	result := DB.Where("id in (?)", ids).Delete(&Folder{})
	return result.Error
}

//func (folder *Folder)GetPositionAbsolute()string{
//	return path.Join(folder.Position,folder.Name)
//}

// MoveFileTo 将此目录下的文件递归移动至dstFolder
func (folder *Folder) MoveFileTo(files []string, dstFolder *Folder) error {
	// 生成绝对路径
	fullFilePath := make([]string, len(files))
	for i := 0; i < len(files); i++ {
		fullFilePath[i] = path.Join(folder.PositionAbsolute, files[i])
	}

	// 更改顶级要移动文件的父目录指向
	err := DB.Model(File{}).Where("dir in (?) and user_id = ?", fullFilePath, folder.OwnerID).
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
	// 生成全局路径
	fullDirs := make([]string, len(dirs))
	for i := 0; i < len(dirs); i++ {
		fullDirs[i] = path.Join(folder.PositionAbsolute, dirs[i])
	}

	// 更改顶级要移动目录的父目录指向
	err := DB.Model(Folder{}).Where("position_absolute in (?) and owner_id = ?", fullDirs, folder.OwnerID).
		Update(map[string]interface{}{
			"parent_id":         dstFolder.ID,
			"position":          dstFolder.PositionAbsolute,
			"position_absolute": gorm.Expr(util.BuildConcat("?", "name", conf.DatabaseConfig.Type), util.FillSlash(dstFolder.PositionAbsolute)),
		}).
		Error
	if err != nil {
		return err
	}

	// 更新被移动的目录递归的子目录
	for _, parentDir := range fullDirs {
		toBeMoved, err := GetRecursiveChildFolder([]string{parentDir}, folder.OwnerID)
		if err != nil {
			return err
		}
		// TODO 找到更好的修改办法
		for _, subFolder := range toBeMoved {
			newPosition := path.Join(dstFolder.PositionAbsolute, strings.Replace(subFolder.Position, folder.Position, "", 1))
			DB.Model(&subFolder).Updates(map[string]interface{}{
				"position":          newPosition,
				"position_absolute": path.Join(newPosition, subFolder.Name),
			})
		}
	}

	// 更新被移动的目录递归的子文件
	for _, parentDir := range fullDirs {
		toBeMoved, err := GetRecursiveByPaths([]string{parentDir}, folder.OwnerID)
		if err != nil {
			return err
		}
		// TODO 找到更好的修改办法
		for _, subFile := range toBeMoved {
			newPosition := path.Join(dstFolder.PositionAbsolute, strings.Replace(subFile.Dir, folder.Position, "", 1))
			DB.Model(&subFile).Updates(map[string]interface{}{
				"dir": newPosition,
			})
		}
	}

	return nil

}
