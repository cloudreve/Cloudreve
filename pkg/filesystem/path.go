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
	 路径/目录相关
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

// Copy 复制src目录下的文件或目录到dst
func (fs *FileSystem) Copy(ctx context.Context, dirs, files []string, src, dst string) error {
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
		subFileSizes, err := srcFolder.MoveOrCopyFolderTo(dirs, dstFolder, true)
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
func (fs *FileSystem) Move(ctx context.Context, dirs, files []string, src, dst string) error {
	// 获取目的目录
	isDstExist, dstFolder := fs.IsPathExist(dst)
	isSrcExist, srcFolder := fs.IsPathExist(src)
	// 不存在时返回空的结果
	if !isDstExist || !isSrcExist {
		return ErrPathNotExist
	}

	// 处理目录及子文件移动
	_, err := srcFolder.MoveOrCopyFolderTo(dirs, dstFolder, false)
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
func (fs *FileSystem) Delete(ctx context.Context, dirs, files []string) error {
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
		return serializer.NewError(serializer.CodeNotFullySuccess, fmt.Sprintf("有 %d 个文件未能成功删除，已删除它们的文件记录", notDeleted), nil)
	}

	return nil
}

// ListDeleteDirs 递归列出要删除目录，及目录下所有文件
func (fs *FileSystem) ListDeleteDirs(ctx context.Context, dirs []string) error {
	// 列出所有递归子目录
	folders, err := model.GetRecursiveChildFolder(dirs, fs.User.ID, true)
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
func (fs *FileSystem) ListDeleteFiles(ctx context.Context, paths []string) error {
	files, err := model.GetFileByPaths(paths, fs.User.ID)
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

	// 获取子目录
	childFolders, _ := folder.GetChildFolder()
	// 获取子文件
	childFiles, _ := folder.GetChildFile()

	// 汇总处理结果
	objects := make([]Object, 0, len(childFiles)+len(childFolders))
	// 所有对象的父目录
	var processedPath string

	for _, folder := range childFolders {
		// 路径处理钩子，
		// 所有对象父目录都是一样的，所以只处理一次
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(folder.Position)
			} else {
				processedPath = folder.Position
			}
		}

		objects = append(objects, Object{
			ID:   folder.ID,
			Name: folder.Name,
			Path: processedPath,
			Pic:  "",
			Size: 0,
			Type: "dir",
			Date: folder.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	for _, file := range childFiles {
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(file.Dir)
			} else {
				processedPath = file.Dir
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
	// 是否有同名目录
	// TODO: 有了unique约束后可以不用在此检查
	b, _ := fs.IsPathExist(fullPath)
	if b {
		return ErrFolderExisted
	}
	// 是否有同名文件
	if ok, _ := fs.IsFileExist(path.Join(base, dir)); ok {
		return ErrFileExisted
	}
	// 创建目录
	newFolder := model.Folder{
		Name:             dir,
		ParentID:         parent.ID,
		Position:         base,
		PositionAbsolute: fullPath,
		OwnerID:          fs.User.ID,
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
func (fs *FileSystem) IsFileExist(fullPath string) (bool, model.File) {
	basePath := path.Dir(fullPath)
	fileName := path.Base(fullPath)

	file, err := model.GetFileByPathAndName(basePath, fileName, fs.User.ID)

	return err == nil, file
}
