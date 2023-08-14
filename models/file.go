package model

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
)

// File 文件
type File struct {
	// 表字段
	gorm.Model
	Name            string `gorm:"unique_index:idx_only_one"`
	SourceName      string `gorm:"type:text"`
	UserID          uint   `gorm:"index:user_id;unique_index:idx_only_one"`
	Size            uint64
	PicInfo         string
	FolderID        uint `gorm:"index:folder_id;unique_index:idx_only_one"`
	PolicyID        uint
	UploadSessionID *string `gorm:"index:session_id;unique_index:session_only_one"`
	Metadata        string  `gorm:"type:text"`

	// 关联模型
	Policy Policy `gorm:"PRELOAD:false,association_autoupdate:false"`

	// 数据库忽略字段
	Position           string            `gorm:"-"`
	MetadataSerialized map[string]string `gorm:"-"`
}

// Thumb related metadata
const (
	ThumbStatusNotExist     = ""
	ThumbStatusExist        = "exist"
	ThumbStatusNotAvailable = "not_available"

	ThumbStatusMetadataKey  = "thumb_status"
	ThumbSidecarMetadataKey = "thumb_sidecar"

	ChecksumMetadataKey = "webdav_checksum"
)

func init() {
	// 注册缓存用到的复杂结构
	gob.Register(File{})
}

// Create 创建文件记录
func (file *File) Create() error {
	tx := DB.Begin()

	if err := tx.Create(file).Error; err != nil {
		util.Log().Warning("Failed to insert file record: %s", err)
		tx.Rollback()
		return err
	}

	user := &User{}
	user.ID = file.UserID
	if err := user.ChangeStorage(tx, "+", file.Size); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AfterFind 找到文件后的钩子
func (file *File) AfterFind() (err error) {
	// 反序列化文件元数据
	if file.Metadata != "" {
		err = json.Unmarshal([]byte(file.Metadata), &file.MetadataSerialized)
	} else {
		file.MetadataSerialized = make(map[string]string)
	}

	return
}

// BeforeSave Save策略前的钩子
func (file *File) BeforeSave() (err error) {
	if len(file.MetadataSerialized) > 0 {
		metaValue, err := json.Marshal(&file.MetadataSerialized)
		file.Metadata = string(metaValue)
		return err
	}

	return nil
}

// GetChildFile 查找目录下名为name的子文件
func (folder *Folder) GetChildFile(name string) (*File, error) {
	var file File
	result := DB.Where("folder_id = ? AND name = ?", folder.ID, name).Find(&file)

	if result.Error == nil {
		file.Position = path.Join(folder.Position, folder.Name)
	}
	return &file, result.Error
}

// GetChildFiles 查找目录下子文件
func (folder *Folder) GetChildFiles() ([]File, error) {
	var files []File
	result := DB.Where("folder_id = ?", folder.ID).Find(&files)

	if result.Error == nil {
		for i := 0; i < len(files); i++ {
			files[i].Position = path.Join(folder.Position, folder.Name)
		}
	}
	return files, result.Error
}

// GetFilesByIDs 根据文件ID批量获取文件,
// UID为0表示忽略用户，只根据文件ID检索
func GetFilesByIDs(ids []uint, uid uint) ([]File, error) {
	return GetFilesByIDsFromTX(DB, ids, uid)
}

func GetFilesByIDsFromTX(tx *gorm.DB, ids []uint, uid uint) ([]File, error) {
	var files []File
	var result *gorm.DB
	if uid == 0 {
		result = tx.Where("id in (?)", ids).Find(&files)
	} else {
		result = tx.Where("id in (?) AND user_id = ?", ids, uid).Find(&files)
	}
	return files, result.Error
}

// GetFilesByKeywords 根据关键字搜索文件,
// UID为0表示忽略用户，只根据文件ID检索. 如果 parents 非空， 则只限制在 parent 包含的目录下搜索
func GetFilesByKeywords(uid uint, parents []uint, keywords ...interface{}) ([]File, error) {
	var (
		files      []File
		result     = DB
		conditions string
	)

	// 生成查询条件
	for i := 0; i < len(keywords); i++ {
		conditions += "name like ?"
		if i != len(keywords)-1 {
			conditions += " or "
		}
	}

	if uid != 0 {
		result = result.Where("user_id = ?", uid)
	}

	if len(parents) > 0 {
		result = result.Where("folder_id in (?)", parents)
	}

	result = result.Where("("+conditions+")", keywords...).Find(&files)

	return files, result.Error
}

// GetChildFilesOfFolders 批量检索目录子文件
func GetChildFilesOfFolders(folders *[]Folder) ([]File, error) {
	// 将所有待检索目录ID抽离，以便检索文件
	folderIDs := make([]uint, 0, len(*folders))
	for _, value := range *folders {
		folderIDs = append(folderIDs, value.ID)
	}

	// 检索文件
	var files []File
	result := DB.Where("folder_id in (?)", folderIDs).Find(&files)
	return files, result.Error
}

// GetUploadPlaceholderFiles 获取所有上传占位文件
// UID为0表示忽略用户
func GetUploadPlaceholderFiles(uid uint) []*File {
	query := DB
	if uid != 0 {
		query = query.Where("user_id = ?", uid)
	}

	var files []*File
	query.Where("upload_session_id is not NULL").Find(&files)
	return files
}

// GetPolicy 获取文件所属策略
func (file *File) GetPolicy() *Policy {
	if file.Policy.Model.ID == 0 {
		file.Policy, _ = GetPolicyByID(file.PolicyID)
	}
	return &file.Policy
}

// RemoveFilesWithSoftLinks 去除给定的文件列表中有软链接的文件
func RemoveFilesWithSoftLinks(files []File) ([]File, error) {
	// 结果值
	filteredFiles := make([]File, 0)

	if len(files) == 0 {
		return filteredFiles, nil
	}

	// 查询软链接的文件
	filesWithSoftLinks := make([]File, 0)
	for _, file := range files {
		var softLinkFile File
		res := DB.
			Where("source_name = ? and policy_id = ? and id != ?", file.SourceName, file.PolicyID, file.ID).
			First(&softLinkFile)
		if res.Error == nil {
			filesWithSoftLinks = append(filesWithSoftLinks, softLinkFile)
		}
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

// DeleteFiles 批量删除文件记录并归还容量
func DeleteFiles(files []*File, uid uint) error {
	tx := DB.Begin()
	user := &User{}
	user.ID = uid
	var size uint64
	for _, file := range files {
		if uid > 0 && file.UserID != uid {
			tx.Rollback()
			return errors.New("user id not consistent")
		}

		result := tx.Unscoped().Where("size = ?", file.Size).Delete(file)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			return errors.New("file size is dirty")
		}

		size += file.Size
	}

	if uid > 0 {
		if err := user.ChangeStorage(tx, "-", size); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetFilesByParentIDs 根据父目录ID查找文件
func GetFilesByParentIDs(ids []uint, uid uint) ([]File, error) {
	files := make([]File, 0, len(ids))
	result := DB.Where("user_id = ? and folder_id in (?)", uid, ids).Find(&files)
	return files, result.Error
}

// GetFilesByUploadSession 查找上传会话对应的文件
func GetFilesByUploadSession(sessionID string, uid uint) (*File, error) {
	file := File{}
	result := DB.Where("user_id = ? and upload_session_id = ?", uid, sessionID).Find(&file)
	return &file, result.Error
}

// Rename 重命名文件
func (file *File) Rename(new string) error {
	if file.MetadataSerialized[ThumbStatusMetadataKey] == ThumbStatusNotAvailable {
		if !strings.EqualFold(filepath.Ext(new), filepath.Ext(file.Name)) {
			// Reset thumb status for new ext name.
			if err := file.resetThumb(); err != nil {
				return err
			}
		}
	}

	return DB.Model(&file).Set("gorm:association_autoupdate", false).Updates(map[string]interface{}{
		"name":     new,
		"metadata": file.Metadata,
	}).Error
}

// UpdatePicInfo 更新文件的图像信息
func (file *File) UpdatePicInfo(value string) error {
	return DB.Model(&file).Set("gorm:association_autoupdate", false).UpdateColumns(File{PicInfo: value}).Error
}

// UpdateMetadata 新增或修改文件的元信息
func (file *File) UpdateMetadata(data map[string]string) error {
	if file.MetadataSerialized == nil {
		file.MetadataSerialized = make(map[string]string)
	}

	for k, v := range data {
		file.MetadataSerialized[k] = v
	}
	metaValue, err := json.Marshal(&file.MetadataSerialized)
	if err != nil {
		return err
	}

	return DB.Model(&file).Set("gorm:association_autoupdate", false).UpdateColumns(File{Metadata: string(metaValue)}).Error
}

// UpdateSize 更新文件的大小信息
// TODO: 全局锁
func (file *File) UpdateSize(value uint64) error {
	tx := DB.Begin()
	var sizeDelta uint64
	operator := "+"
	user := User{}
	user.ID = file.UserID
	if value > file.Size {
		sizeDelta = value - file.Size
	} else {
		operator = "-"
		sizeDelta = file.Size - value
	}

	if err := file.resetThumb(); err != nil {
		tx.Rollback()
		return err
	}

	if res := tx.Model(&file).
		Where("size = ?", file.Size).
		Set("gorm:association_autoupdate", false).
		Updates(map[string]interface{}{
			"size":     value,
			"metadata": file.Metadata,
		}); res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	if err := user.ChangeStorage(tx, operator, sizeDelta); err != nil {
		tx.Rollback()
		return err
	}

	file.Size = value
	return tx.Commit().Error
}

// UpdateSourceName 更新文件的源文件名
func (file *File) UpdateSourceName(value string) error {
	if err := file.resetThumb(); err != nil {
		return err
	}

	return DB.Model(&file).Set("gorm:association_autoupdate", false).Updates(map[string]interface{}{
		"source_name": value,
		"metadata":    file.Metadata,
	}).Error
}

func (file *File) PopChunkToFile(lastModified *time.Time, picInfo string) error {
	file.UploadSessionID = nil
	if lastModified != nil {
		file.UpdatedAt = *lastModified
	}

	return DB.Model(file).UpdateColumns(map[string]interface{}{
		"upload_session_id": file.UploadSessionID,
		"updated_at":        file.UpdatedAt,
		"pic_info":          picInfo,
	}).Error
}

// CanCopy 返回文件是否可被复制
func (file *File) CanCopy() bool {
	return file.UploadSessionID == nil
}

// CreateOrGetSourceLink creates a SourceLink model. If the given model exists, the existing
// model will be returned.
func (file *File) CreateOrGetSourceLink() (*SourceLink, error) {
	res := &SourceLink{}
	err := DB.Set("gorm:auto_preload", true).Where("file_id = ?", file.ID).Find(&res).Error
	if err == nil && res.ID > 0 {
		return res, nil
	}

	res.FileID = file.ID
	res.Name = file.Name
	if err := DB.Save(res).Error; err != nil {
		return nil, fmt.Errorf("failed to insert SourceLink: %w", err)
	}

	res.File = *file
	return res, nil
}

func (file *File) resetThumb() error {
	if _, ok := file.MetadataSerialized[ThumbStatusMetadataKey]; !ok {
		return nil
	}

	delete(file.MetadataSerialized, ThumbStatusMetadataKey)
	metaValue, err := json.Marshal(&file.MetadataSerialized)
	file.Metadata = string(metaValue)
	return err
}

/*
	实现 webdav.FileInfo 接口
*/

func (file *File) GetName() string {
	return file.Name
}

func (file *File) GetSize() uint64 {
	return file.Size
}
func (file *File) ModTime() time.Time {
	return file.UpdatedAt
}

func (file *File) IsDir() bool {
	return false
}

func (file *File) GetPosition() string {
	return file.Position
}

// ShouldLoadThumb returns if file explorer should try to load thumbnail for this file.
// `True` does not guarantee the load request will success in next step, but the client
// should try to load and fallback to default placeholder in case error returned.
func (file *File) ShouldLoadThumb() bool {
	return file.MetadataSerialized[ThumbStatusMetadataKey] != ThumbStatusNotAvailable && file.Size != 0
}

// return sidecar thumb file name
func (file *File) ThumbFile() string {
	return file.SourceName + GetSettingByNameWithDefault("thumb_file_suffix", "._thumb")
}
