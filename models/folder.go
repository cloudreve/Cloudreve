package model

import (
	"errors"
	"path"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
)

// Folder 目录
type Folder struct {
	// 表字段
	gorm.Model
	Name     string `gorm:"unique_index:idx_only_one_name"`
	ParentID *uint  `gorm:"index:parent_id;unique_index:idx_only_one_name"`
	OwnerID  uint   `gorm:"index:owner_id"`
	PolicyID uint   // Webdav下挂载的存储策略ID

	// 数据库忽略字段
	Position        string `gorm:"-"`
	WebdavDstName   string `gorm:"-"`
	InheritPolicyID uint   `gorm:"-"` //  从父目录继承而来的policy id，默认值则使用自身的的PolicyID
}

// Create 创建目录
func (folder *Folder) Create() (uint, error) {
	if err := DB.FirstOrCreate(folder, *folder).Error; err != nil {
		folder.Model = gorm.Model{}
		err2 := DB.First(folder, *folder).Error
		return folder.ID, err2
	}

	return folder.ID, nil
}

// GetMountedFolders 列出已挂载存储策略的目录
func GetMountedFolders(uid uint) []Folder {
	var folders []Folder
	DB.Where("owner_id = ? and policy_id <> ?", uid, 0).Find(&folders)
	return folders
}

// GetChild 返回folder下名为name的子目录，不存在则返回错误
func (folder *Folder) GetChild(name string) (*Folder, error) {
	var resFolder Folder
	err := DB.
		Where("parent_id = ? AND owner_id = ? AND name = ?", folder.ID, folder.OwnerID, name).
		First(&resFolder).Error

	// 将子目录的路径及存储策略传递下去
	if err == nil {
		resFolder.Position = path.Join(folder.Position, folder.Name)
		if folder.PolicyID > 0 {
			resFolder.InheritPolicyID = folder.PolicyID
		} else if folder.InheritPolicyID > 0 {
			resFolder.InheritPolicyID = folder.InheritPolicyID
		}
	}
	return &resFolder, err
}

// TraceRoot 向上递归查找父目录
func (folder *Folder) TraceRoot() error {
	if folder.ParentID == nil {
		return nil
	}

	var parentFolder Folder
	err := DB.
		Where("id = ? AND owner_id = ?", folder.ParentID, folder.OwnerID).
		First(&parentFolder).Error

	if err == nil {
		err := parentFolder.TraceRoot()
		folder.Position = path.Join(parentFolder.Position, parentFolder.Name)
		return err
	}

	return err
}

// GetChildFolder 查找子目录
func (folder *Folder) GetChildFolder() ([]Folder, error) {
	var folders []Folder
	result := DB.Where("parent_id = ?", folder.ID).Find(&folders)

	if result.Error == nil {
		for i := 0; i < len(folders); i++ {
			folders[i].Position = path.Join(folder.Position, folder.Name)
		}
	}
	return folders, result.Error
}

// GetRecursiveChildFolder 查找所有递归子目录，包括自身
func GetRecursiveChildFolder(dirs []uint, uid uint, includeSelf bool) ([]Folder, error) {
	folders := make([]Folder, 0, len(dirs))
	var err error

	var parFolders []Folder
	result := DB.Where("owner_id = ? and id in (?)", uid, dirs).Find(&parFolders)
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

	return folders, err
}

// DeleteFolderByIDs 根据给定ID批量删除目录记录
func DeleteFolderByIDs(ids []uint) error {
	result := DB.Where("id in (?)", ids).Unscoped().Delete(&Folder{})
	return result.Error
}

// GetFoldersByIDs 根据ID和用户查找所有目录
func GetFoldersByIDs(ids []uint, uid uint) ([]Folder, error) {
	var folders []Folder
	result := DB.Where("id in (?) AND owner_id = ?", ids, uid).Find(&folders)
	return folders, result.Error
}

// MoveOrCopyFileTo 将此目录下的files移动或复制至dstFolder，
// 返回此操作新增的容量
func (folder *Folder) MoveOrCopyFileTo(files []uint, dstFolder *Folder, isCopy bool) (uint64, error) {
	// 已复制文件的总大小
	var copiedSize uint64

	if isCopy {
		// 检索出要复制的文件
		var originFiles = make([]File, 0, len(files))
		if err := DB.Where(
			"id in (?) and user_id = ? and folder_id = ?",
			files,
			folder.OwnerID,
			folder.ID,
		).Find(&originFiles).Error; err != nil {
			return 0, err
		}

		// 复制文件记录
		for _, oldFile := range originFiles {
			if !oldFile.CanCopy() {
				util.Log().Warning("Cannot copy file %q because it's being uploaded now, skipping...", oldFile.Name)
				continue
			}

			oldFile.Model = gorm.Model{}
			oldFile.FolderID = dstFolder.ID
			oldFile.UserID = dstFolder.OwnerID

			// webdav目标名重置
			if dstFolder.WebdavDstName != "" {
				oldFile.Name = dstFolder.WebdavDstName
			}

			if err := DB.Create(&oldFile).Error; err != nil {
				return copiedSize, err
			}

			copiedSize += oldFile.Size
		}

	} else {
		var updates = map[string]interface{}{
			"folder_id": dstFolder.ID,
		}
		// webdav目标名重置
		if dstFolder.WebdavDstName != "" {
			updates["name"] = dstFolder.WebdavDstName
		}

		// 更改顶级要移动文件的父目录指向
		err := DB.Model(File{}).Where(
			"id in (?) and user_id = ? and folder_id = ?",
			files,
			folder.OwnerID,
			folder.ID,
		).
			Update(updates).
			Error
		if err != nil {
			return 0, err
		}

	}

	return copiedSize, nil

}

// CopyFolderTo 将此目录及其子目录及文件递归复制至dstFolder
// 返回此操作新增的容量
func (folder *Folder) CopyFolderTo(folderID uint, dstFolder *Folder) (size uint64, err error) {
	// 列出所有子目录
	subFolders, err := GetRecursiveChildFolder([]uint{folderID}, folder.OwnerID, true)
	if err != nil {
		return 0, err
	}

	// 抽离所有子目录的ID
	var subFolderIDs = make([]uint, len(subFolders))
	for key, value := range subFolders {
		subFolderIDs[key] = value.ID
	}

	// 复制子目录
	var newIDCache = make(map[uint]uint)
	for _, folder := range subFolders {
		// 新的父目录指向
		var newID uint
		// 顶级目录直接指向新的目的目录
		if folder.ID == folderID {
			newID = dstFolder.ID
			// webdav目标名重置
			if dstFolder.WebdavDstName != "" {
				folder.Name = dstFolder.WebdavDstName
			}
		} else if IDCache, ok := newIDCache[*folder.ParentID]; ok {
			newID = IDCache
		} else {
			util.Log().Warning("Failed to get parent folder %q", *folder.ParentID)
			return size, errors.New("Failed to get parent folder")
		}

		// 插入新的目录记录
		oldID := folder.ID
		folder.Model = gorm.Model{}
		folder.ParentID = &newID
		folder.OwnerID = dstFolder.OwnerID
		if err = DB.Create(&folder).Error; err != nil {
			return size, err
		}
		// 记录新的ID以便其子目录使用
		newIDCache[oldID] = folder.ID

	}

	// 复制文件
	var originFiles = make([]File, 0, len(subFolderIDs))
	if err := DB.Where(
		"user_id = ? and folder_id in (?)",
		folder.OwnerID,
		subFolderIDs,
	).Find(&originFiles).Error; err != nil {
		return 0, err
	}

	// 复制文件记录
	for _, oldFile := range originFiles {
		if !oldFile.CanCopy() {
			util.Log().Warning("Cannot copy file %q because it's being uploaded now, skipping...", oldFile.Name)
			continue
		}

		oldFile.Model = gorm.Model{}
		oldFile.FolderID = newIDCache[oldFile.FolderID]
		oldFile.UserID = dstFolder.OwnerID
		if err := DB.Create(&oldFile).Error; err != nil {
			return size, err
		}

		size += oldFile.Size
	}

	return size, nil

}

// MoveFolderTo 将folder目录下的dirs子目录复制或移动到dstFolder，
// 返回此过程中增加的容量
func (folder *Folder) MoveFolderTo(dirs []uint, dstFolder *Folder) error {

	// 如果目标位置为待移动的目录，会导致 parent 为自己
	// 造成死循环且无法被除搜索以外的组件展示
	if folder.OwnerID == dstFolder.OwnerID && util.ContainsUint(dirs, dstFolder.ID) {
		return errors.New("cannot move a folder into itself")
	}

	var updates = map[string]interface{}{
		"parent_id": dstFolder.ID,
	}
	// webdav目标名重置
	if dstFolder.WebdavDstName != "" {
		updates["name"] = dstFolder.WebdavDstName
	}

	// 更改顶级要移动目录的父目录指向
	err := DB.Model(Folder{}).Where(
		"id in (?) and owner_id = ? and parent_id = ?",
		dirs,
		folder.OwnerID,
		folder.ID,
	).Update(updates).Error

	return err

}

// Rename 重命名目录
func (folder *Folder) Rename(new string) error {
	return DB.Model(&folder).UpdateColumn("name", new).Error
}

// Mount 目录挂载
func (folder *Folder) Mount(new uint) error {
	return DB.Model(&folder).Update("policy_id", new).Error
}

/*
	实现 FileInfo.FileInfo 接口
	TODO 测试
*/

func (folder *Folder) GetName() string {
	return folder.Name
}

func (folder *Folder) GetSize() uint64 {
	return 0
}
func (folder *Folder) ModTime() time.Time {
	return folder.UpdatedAt
}
func (folder *Folder) IsDir() bool {
	return true
}
func (folder *Folder) GetPosition() string {
	return folder.Position
}
