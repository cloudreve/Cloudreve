package model

import (
	"errors"
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

// MoveOrCopyFileTo 将此目录下的files移动或复制至dstFolder，
// 返回此操作新增的容量
func (folder *Folder) MoveOrCopyFileTo(files []string, dstFolder *Folder, isCopy bool) (uint64, error) {
	// 已复制文件的总大小
	var copiedSize uint64

	if isCopy {
		// 检索出要复制的文件
		var originFiles = make([]File, 0, len(files))
		if err := DB.Where(
			"name in (?) and user_id = ? and dir = ?",
			files,
			folder.OwnerID,
			folder.PositionAbsolute,
		).Find(&originFiles).Error; err != nil {
			return 0, err
		}

		// 复制文件记录
		for _, oldFile := range originFiles {
			oldFile.Model = gorm.Model{}
			oldFile.FolderID = dstFolder.ID
			oldFile.Dir = dstFolder.PositionAbsolute

			if err := DB.Create(&oldFile).Error; err != nil {
				return copiedSize, err
			}

			copiedSize += oldFile.Size
		}

	} else {
		// 更改顶级要移动文件的父目录指向
		err := DB.Model(File{}).Where(
			"name in (?) and user_id = ? and dir = ?",
			files,
			folder.OwnerID,
			folder.PositionAbsolute,
		).
			Update(map[string]interface{}{
				"folder_id": dstFolder.ID,
				"dir":       dstFolder.PositionAbsolute,
			}).
			Error
		if err != nil {
			return 0, err
		}

	}

	return copiedSize, nil

}

// MoveOrCopyFolderTo 将folder目录下的dirs子目录复制或移动到dstFolder，
// 返回此过程中增加的容量
func (folder *Folder) MoveOrCopyFolderTo(dirs []string, dstFolder *Folder, isCopy bool) (uint64, error) {
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
			return 0, err
		}
		subFolders[key] = toBeMoved
	}

	// 记录复制要用到的父目录源路径和新的ID
	var copyCache = make(map[string]uint)
	// 记录已复制文件的容量
	var newUsedStorage uint64

	var err error
	if isCopy {
		// 复制
		// TODO:支持多目录
		origin := Folder{}
		if DB.Where(
			"position_absolute in (?) and owner_id = ?",
			fullDirs,
			folder.OwnerID,
		).Find(&origin).Error != nil {
			return 0, errors.New("找不到原始目录")
		}

		oldPosition := origin.PositionAbsolute

		// 更新复制后的相关属性
		origin.PositionAbsolute = util.FillSlash(dstFolder.PositionAbsolute) + origin.Name
		origin.Position = dstFolder.PositionAbsolute
		origin.ParentID = dstFolder.ID

		// 清空主键
		origin.Model = gorm.Model{}

		if err := DB.Create(&origin).Error; err != nil {
			return 0, err
		}

		// 记录新的主键
		copyCache[oldPosition] = origin.Model.ID

	} else {
		// 移动
		// 更改顶级要移动目录的父目录指向
		err = DB.Model(Folder{}).
			Where("position_absolute in (?) and owner_id = ?",
				fullDirs,
				folder.OwnerID,
			).
			Update(map[string]interface{}{
				"parent_id": dstFolder.ID,
				"position":  dstFolder.PositionAbsolute,
				"position_absolute": gorm.Expr(
					util.BuildConcat("?",
						"name",
						conf.DatabaseConfig.Type,
					),
					util.FillSlash(dstFolder.PositionAbsolute),
				),
			}).Error
	}

	if err != nil {
		return 0, err
	}

	// 更新被移动的目录递归的子目录和文件
	for parKey, toBeMoved := range subFolders {
		ignorePath := fullDirs[parKey]
		// TODO 找到更好的修改办法

		// 抽离所有子目录的ID
		var subFolderIDs = make([]uint, len(toBeMoved))
		for key, subFolder := range toBeMoved {
			subFolderIDs[key] = subFolder.ID
		}

		if isCopy {
			index := 0
			for len(toBeMoved) != 0 {
				innerIndex := index % len(toBeMoved)
				index++
				// 限制循环次数
				if index > 65535 {
					return 0, errors.New("循环超出限制")
				}

				// 如果是顶级父目录，直接删除，不需要复制
				if toBeMoved[innerIndex].PositionAbsolute == ignorePath {
					toBeMoved = append(toBeMoved[:innerIndex], toBeMoved[innerIndex+1:]...)
					continue
				}

				// 如果缓存中存在父目录ID，执行复制,并删除
				if newID, ok := copyCache[toBeMoved[innerIndex].Position]; ok {
					// 记录目录原来的路径
					oldPosition := toBeMoved[innerIndex].PositionAbsolute

					// 设置目录i虚拟的路径
					newPosition := path.Join(
						dstFolder.PositionAbsolute, strings.Replace(
							toBeMoved[innerIndex].Position,
							folder.PositionAbsolute, "", 1),
					)
					toBeMoved[innerIndex].Position = newPosition
					toBeMoved[innerIndex].PositionAbsolute = path.Join(
						newPosition,
						toBeMoved[innerIndex].Name,
					)
					toBeMoved[innerIndex].ParentID = newID
					toBeMoved[innerIndex].Model = gorm.Model{}
					if err := DB.Create(&toBeMoved[innerIndex]).Error; err != nil {
						return 0, err
					}

					// 将当前目录老路径和新ID保存，以便后续待处理目录文件使用
					copyCache[oldPosition] = toBeMoved[innerIndex].Model.ID
					toBeMoved = append(toBeMoved[:innerIndex], toBeMoved[innerIndex+1:]...)
				}
			}

		} else {
			for _, subFolder := range toBeMoved {
				// 每个分组的第一个目录已经变更指向，直接跳过
				if subFolder.PositionAbsolute != ignorePath {
					newPosition := path.Join(dstFolder.PositionAbsolute,
						strings.Replace(subFolder.Position,
							folder.PositionAbsolute,
							"",
							1,
						),
					)
					// 移动
					DB.Model(&subFolder).Updates(map[string]interface{}{
						"position":          newPosition,
						"position_absolute": path.Join(newPosition, subFolder.Name),
					})
				}
			}
		}

		// 获取子目录下的所有子文件
		toBeMovedFile, err := GetFilesByParentIDs(subFolderIDs, folder.OwnerID)
		if err != nil {
			return 0, err
		}

		// 开始复制或移动子文件
		for _, subFile := range toBeMovedFile {
			newPosition := path.Join(dstFolder.PositionAbsolute,
				strings.Replace(
					subFile.Dir,
					folder.PositionAbsolute,
					"",
					1,
				),
			)
			if isCopy {
				// 复制
				if newID, ok := copyCache[subFile.Dir]; ok {
					subFile.FolderID = newID
				} else {
					util.Log().Debug("无法找到文件的父目录ID，原始路径：%s", subFile.Dir)
				}
				subFile.Dir = newPosition
				subFile.Model = gorm.Model{}

				// 复制文件记录
				if err := DB.Create(&subFile).Error; err != nil {
					util.Log().Warning("无法复制子文件：%s", err)
				} else {
					// 记录此文件容量
					newUsedStorage += subFile.Size
				}
			} else {
				DB.Model(&subFile).Updates(map[string]interface{}{
					"dir": newPosition,
				})
			}

		}

	}

	return newUsedStorage, nil

}

// Rename 重命名目录
func (folder *Folder) Rename(new string) error {
	if err := DB.Model(&folder).Update("name", new).Error; err != nil {
		return err
	}
	return nil
}
