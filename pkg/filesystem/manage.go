package filesystem

import (
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"path"
)

/* =================
	 文件/目录管理
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

// Rename 重命名对象
func (fs *FileSystem) Rename(ctx context.Context, src, new string) (err error) {
	// 验证新名字
	if !fs.ValidateLegalName(ctx, new) || !fs.ValidateExtension(ctx, new) {
		return ErrIllegalObjectName
	}

	// 如果源对象是文件
	fileExist, file := fs.IsFileExist(src)
	if fileExist {
		err = file.Rename(new)
		return err
	}

	// 源对象是目录
	folderExist, folder := fs.IsPathExist(src)
	if folderExist {
		err = folder.Rename(new)
		return err
	}

	return ErrPathNotExist
}

// Copy 复制src目录下的文件或目录到dst，
// 暂时只支持单文件
func (fs *FileSystem) Copy(ctx context.Context, dirs, files []uint, src, dst string) error {
	// 获取目的目录
	isDstExist, dstFolder := fs.IsPathExist(dst)
	isSrcExist, srcFolder := fs.IsPathExist(src)
	// 不存在时返回空的结果
	if !isDstExist || !isSrcExist {
		return ErrPathNotExist
	}

	// 记录复制的文件的总容量
	var newUsedStorage uint64

	// 复制目录
	if len(dirs) > 0 {
		subFileSizes, err := srcFolder.CopyFolderTo(dirs[0], dstFolder)
		if err != nil {
			return serializer.NewError(serializer.CodeDBError, "操作失败，可能有重名冲突", err)
		}
		newUsedStorage += subFileSizes
	}

	// 复制文件
	if len(files) > 0 {
		subFileSizes, err := srcFolder.MoveOrCopyFileTo(files, dstFolder, true)
		if err != nil {
			return serializer.NewError(serializer.CodeDBError, "操作失败，可能有重名冲突", err)
		}
		newUsedStorage += subFileSizes
	}

	// 扣除容量
	fs.User.IncreaseStorageWithoutCheck(newUsedStorage)

	return nil
}

// Move 移动文件和目录
func (fs *FileSystem) Move(ctx context.Context, dirs, files []uint, src, dst string) error {
	// 获取目的目录
	isDstExist, dstFolder := fs.IsPathExist(dst)
	isSrcExist, srcFolder := fs.IsPathExist(src)
	// 不存在时返回空的结果
	if !isDstExist || !isSrcExist {
		return ErrPathNotExist
	}

	// 处理目录及子文件移动
	err := srcFolder.MoveFolderTo(dirs, dstFolder)
	if err != nil {
		return serializer.NewError(serializer.CodeDBError, "操作失败，可能有重名冲突", err)
	}

	// 处理文件移动
	_, err = srcFolder.MoveOrCopyFileTo(files, dstFolder, false)
	if err != nil {
		return serializer.NewError(serializer.CodeDBError, "操作失败，可能有重名冲突", err)
	}

	// 移动文件

	return err
}

// Delete 递归删除对象
func (fs *FileSystem) Delete(ctx context.Context, dirs, files []uint) error {
	// 已删除的总容量,map用于去重
	var deletedStorage = make(map[uint]uint64)
	// 已删除的文件ID
	var deletedFileIDs = make([]uint, 0, len(fs.FileTarget))
	// 删除失败的文件的父目录ID

	// 所有文件的ID
	var allFileIDs = make([]uint, 0, len(fs.FileTarget))

	// 列出要删除的目录
	if len(dirs) > 0 {
		err := fs.ListDeleteDirs(ctx, dirs)
		if err != nil {
			return err
		}
	}

	// 列出要删除的文件
	if len(files) > 0 {
		err := fs.ListDeleteFiles(ctx, files)
		if err != nil {
			return err
		}
	}

	// 去除待删除文件中包含软连接的部分
	filesToBeDelete, err := model.RemoveFilesWithSoftLinks(fs.FileTarget)
	if err != nil {
		return ErrDBListObjects.WithError(err)
	}

	// 根据存储策略将文件分组
	policyGroup := fs.GroupFileByPolicy(ctx, filesToBeDelete)

	// 按照存储策略分组删除对象
	failed := fs.deleteGroupedFile(ctx, policyGroup)

	for i := 0; i < len(fs.FileTarget); i++ {
		if util.ContainsString(failed[fs.FileTarget[i].PolicyID], fs.FileTarget[i].SourceName) {
			// TODO 删除失败时不删除文件记录及父目录
		} else {
			deletedFileIDs = append(deletedFileIDs, fs.FileTarget[i].ID)
		}
		deletedStorage[fs.FileTarget[i].ID] = fs.FileTarget[i].Size
		allFileIDs = append(allFileIDs, fs.FileTarget[i].ID)
	}

	// 删除文件记录
	err = model.DeleteFileByIDs(allFileIDs)
	if err != nil {
		return ErrDBDeleteObjects.WithError(err)
	}

	// 归还容量
	var total uint64
	for _, value := range deletedStorage {
		total += value
	}
	fs.User.DeductionStorage(total)

	// 删除目录
	var allFolderIDs = make([]uint, 0, len(fs.DirTarget))
	for _, value := range fs.DirTarget {
		allFolderIDs = append(allFolderIDs, value.ID)
	}
	err = model.DeleteFolderByIDs(allFolderIDs)
	if err != nil {
		return ErrDBDeleteObjects.WithError(err)
	}

	if notDeleted := len(fs.FileTarget) - len(deletedFileIDs); notDeleted > 0 {
		return serializer.NewError(
			serializer.CodeNotFullySuccess,
			fmt.Sprintf("有 %d 个文件未能成功删除，已删除它们的文件记录", notDeleted),
			nil,
		)
	}

	return nil
}

// ListDeleteDirs 递归列出要删除目录，及目录下所有文件
func (fs *FileSystem) ListDeleteDirs(ctx context.Context, ids []uint) error {
	// 列出所有递归子目录
	folders, err := model.GetRecursiveChildFolder(ids, fs.User.ID, true)
	if err != nil {
		return ErrDBListObjects.WithError(err)
	}
	fs.SetTargetDir(&folders)

	// 检索目录下的子文件
	files, err := model.GetChildFilesOfFolders(&folders)
	if err != nil {
		return ErrDBListObjects.WithError(err)
	}
	fs.SetTargetFile(&files)

	return nil
}

// ListDeleteFiles 根据给定的路径列出要删除的文件
func (fs *FileSystem) ListDeleteFiles(ctx context.Context, ids []uint) error {
	files, err := model.GetFilesByIDs(ids, fs.User.ID)
	if err != nil {
		return ErrDBListObjects.WithError(err)
	}
	fs.SetTargetFile(&files)
	return nil
}

// List 列出路径下的内容,
// pathProcessor为最终对象路径的处理钩子。
// 有些情况下（如在分享页面列对象）时，
// 路径需要截取掉被分享目录路径之前的部分。
func (fs *FileSystem) List(ctx context.Context, dirPath string, pathProcessor func(string) string) ([]Object, error) {
	// 获取父目录
	isExist, folder := fs.IsPathExist(dirPath)
	// 不存在时返回空的结果
	if !isExist {
		return []Object{}, nil
	}

	var parentPath = path.Join(folder.PositionTemp, folder.Name)
	var childFolders []model.Folder
	var childFiles []model.File

	// 获取子目录
	childFolders, _ = folder.GetChildFolder()

	// 获取子文件
	childFiles, _ = folder.GetChildFiles()

	// 汇总处理结果
	objects := make([]Object, 0, len(childFiles)+len(childFolders))
	// 所有对象的父目录
	var processedPath string

	for _, subFolder := range childFolders {
		// 路径处理钩子，
		// 所有对象父目录都是一样的，所以只处理一次
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(parentPath)
			} else {
				processedPath = parentPath
			}
		}

		objects = append(objects, Object{
			ID:   subFolder.ID,
			Name: subFolder.Name,
			Path: processedPath,
			Pic:  "",
			Size: 0,
			Type: "dir",
			Date: subFolder.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	for _, file := range childFiles {
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(parentPath)
			} else {
				processedPath = parentPath
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
	if ok, _ := fs.IsChildFileExist(parent, dir); ok {
		return ErrFileExisted
	}

	// 创建目录
	newFolder := model.Folder{
		Name:     dir,
		ParentID: parent.ID,
		OwnerID:  fs.User.ID,
	}
	_, err := newFolder.Create()

	if err != nil {
		return ErrFolderExisted
	}
	return nil
}
