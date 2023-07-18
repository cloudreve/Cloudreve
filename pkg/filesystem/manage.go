package filesystem

import (
	"context"
	"fmt"
	"path"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

/* =================
	 文件/目录管理
   =================
*/

// Rename 重命名对象
func (fs *FileSystem) Rename(ctx context.Context, dir, file []uint, new string) (err error) {
	// 验证新名字
	if !fs.ValidateLegalName(ctx, new) || (len(file) > 0 && !fs.ValidateExtension(ctx, new)) {
		return ErrIllegalObjectName
	}

	// 如果源对象是文件
	if len(file) > 0 {
		fileObject, err := model.GetFilesByIDs([]uint{file[0]}, fs.User.ID)
		if err != nil || len(fileObject) == 0 {
			return ErrPathNotExist
		}

		err = fileObject[0].Rename(new)
		if err != nil {
			return ErrFileExisted
		}
		return nil
	}

	if len(dir) > 0 {
		folderObject, err := model.GetFoldersByIDs([]uint{dir[0]}, fs.User.ID)
		if err != nil || len(folderObject) == 0 {
			return ErrPathNotExist
		}

		err = folderObject[0].Rename(new)
		if err != nil {
			return ErrFileExisted
		}
		return nil
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

	// 设置webdav目标名
	if dstName, ok := ctx.Value(fsctx.WebdavDstName).(string); ok {
		dstFolder.WebdavDstName = dstName
	}

	// 复制目录
	if len(dirs) > 0 {
		subFileSizes, err := srcFolder.CopyFolderTo(dirs[0], dstFolder)
		if err != nil {
			return ErrObjectNotExist.WithError(err)
		}
		newUsedStorage += subFileSizes
	}

	// 复制文件
	if len(files) > 0 {
		subFileSizes, err := srcFolder.MoveOrCopyFileTo(files, dstFolder, true)
		if err != nil {
			return ErrObjectNotExist.WithError(err)
		}
		newUsedStorage += subFileSizes
	}

	// 扣除容量
	fs.User.IncreaseStorageWithoutCheck(newUsedStorage)

	return nil
}

// Move 移动文件和目录, 将id列表dirs和files从src移动至dst
func (fs *FileSystem) Move(ctx context.Context, dirs, files []uint, src, dst string) error {
	// 获取目的目录
	isDstExist, dstFolder := fs.IsPathExist(dst)
	isSrcExist, srcFolder := fs.IsPathExist(src)
	// 不存在时返回空的结果
	if !isDstExist || !isSrcExist {
		return ErrPathNotExist
	}

	// 设置webdav目标名
	if dstName, ok := ctx.Value(fsctx.WebdavDstName).(string); ok {
		dstFolder.WebdavDstName = dstName
	}

	// 处理目录及子文件移动
	err := srcFolder.MoveFolderTo(dirs, dstFolder)
	if err != nil {
		return ErrFileExisted.WithError(err)
	}

	// 处理文件移动
	_, err = srcFolder.MoveOrCopyFileTo(files, dstFolder, false)
	if err != nil {
		return ErrFileExisted.WithError(err)
	}

	// 移动文件

	return err
}

// Delete 递归删除对象, force 为 true 时强制删除文件记录，忽略物理删除是否成功;
// unlink 为 true 时只删除虚拟文件系统的文件记录，不删除物理文件。
func (fs *FileSystem) Delete(ctx context.Context, dirs, files []uint, force, unlink bool) error {
	// 已删除的文件ID
	var deletedFiles = make([]*model.File, 0, len(fs.FileTarget))
	// 删除失败的文件的父目录ID

	// 所有文件的ID
	var allFiles = make([]*model.File, 0, len(fs.FileTarget))

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
	failed := make(map[uint][]string)
	if !unlink {
		failed = fs.deleteGroupedFile(ctx, policyGroup)
	}

	// 整理删除结果
	for i := 0; i < len(fs.FileTarget); i++ {
		if !util.ContainsString(failed[fs.FileTarget[i].PolicyID], fs.FileTarget[i].SourceName) {
			// 已成功删除的文件
			deletedFiles = append(deletedFiles, &fs.FileTarget[i])
		}

		// 全部文件
		allFiles = append(allFiles, &fs.FileTarget[i])
	}

	// 如果强制删除，则将全部文件视为删除成功
	if force {
		deletedFiles = allFiles
	}

	// 删除文件记录
	err = model.DeleteFiles(deletedFiles, fs.User.ID)
	if err != nil {
		return ErrDBDeleteObjects.WithError(err)
	}

	// 删除文件记录对应的分享记录
	// TODO 先取消分享再删除文件
	deletedFileIDs := make([]uint, len(deletedFiles))
	for k, file := range deletedFiles {
		deletedFileIDs[k] = file.ID
	}

	model.DeleteShareBySourceIDs(deletedFileIDs, false)

	// 如果文件全部删除成功，继续删除目录
	if len(deletedFiles) == len(allFiles) {
		var allFolderIDs = make([]uint, 0, len(fs.DirTarget))
		for _, value := range fs.DirTarget {
			allFolderIDs = append(allFolderIDs, value.ID)
		}
		err = model.DeleteFolderByIDs(allFolderIDs)
		if err != nil {
			return ErrDBDeleteObjects.WithError(err)
		}

		// 删除目录记录对应的分享记录
		model.DeleteShareBySourceIDs(allFolderIDs, true)
	}

	if notDeleted := len(fs.FileTarget) - len(deletedFiles); notDeleted > 0 {
		return serializer.NewError(
			serializer.CodeNotFullySuccess,
			fmt.Sprintf("Failed to delete %d file(s).", notDeleted),
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

	// 忽略根目录
	for i := 0; i < len(folders); i++ {
		if folders[i].ParentID == nil {
			folders = append(folders[:i], folders[i+1:]...)
			break
		}
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
func (fs *FileSystem) List(ctx context.Context, dirPath string, pathProcessor func(string) string) ([]serializer.Object, error) {
	// 获取父目录
	isExist, folder := fs.IsPathExist(dirPath)
	if !isExist {
		return nil, ErrPathNotExist
	}
	fs.SetTargetDir(&[]model.Folder{*folder})

	var parentPath = path.Join(folder.Position, folder.Name)
	var childFolders []model.Folder
	var childFiles []model.File

	// 获取子目录
	childFolders, _ = folder.GetChildFolder()

	// 获取子文件
	childFiles, _ = folder.GetChildFiles()

	return fs.listObjects(ctx, parentPath, childFiles, childFolders, pathProcessor), nil
}

// ListPhysical 列出存储策略中的外部目录
// TODO:测试
func (fs *FileSystem) ListPhysical(ctx context.Context, dirPath string) ([]serializer.Object, error) {
	if err := fs.DispatchHandler(); fs.Policy == nil || err != nil {
		return nil, ErrUnknownPolicyType
	}

	// 存储策略不支持列取时，返回空结果
	if !fs.Policy.CanStructureBeListed() {
		return nil, nil
	}

	// 列取路径
	objects, err := fs.Handler.List(ctx, dirPath, false)
	if err != nil {
		return nil, err
	}

	var (
		folders []model.Folder
	)
	for _, object := range objects {
		if object.IsDir {
			folders = append(folders, model.Folder{
				Name: object.Name,
			})
		}
	}

	return fs.listObjects(ctx, dirPath, nil, folders, nil), nil
}

func (fs *FileSystem) listObjects(ctx context.Context, parent string, files []model.File, folders []model.Folder, pathProcessor func(string) string) []serializer.Object {
	// 分享文件的ID
	shareKey := ""
	if key, ok := ctx.Value(fsctx.ShareKeyCtx).(string); ok {
		shareKey = key
	}

	// 汇总处理结果
	objects := make([]serializer.Object, 0, len(files)+len(folders))

	// 所有对象的父目录
	var processedPath string

	for _, subFolder := range folders {
		// 路径处理钩子，
		// 所有对象父目录都是一样的，所以只处理一次
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(parent)
			} else {
				processedPath = parent
			}
		}

		objects = append(objects, serializer.Object{
			ID:         hashid.HashID(subFolder.ID, hashid.FolderID),
			Name:       subFolder.Name,
			Path:       processedPath,
			Size:       0,
			Type:       "dir",
			Date:       subFolder.UpdatedAt,
			CreateDate: subFolder.CreatedAt,
		})
	}

	for _, file := range files {
		if processedPath == "" {
			if pathProcessor != nil {
				processedPath = pathProcessor(parent)
			} else {
				processedPath = parent
			}
		}

		if file.UploadSessionID == nil {
			newFile := serializer.Object{
				ID:            hashid.HashID(file.ID, hashid.FileID),
				Name:          file.Name,
				Path:          processedPath,
				Thumb:         file.ShouldLoadThumb(),
				Size:          file.Size,
				Type:          "file",
				Date:          file.UpdatedAt,
				SourceEnabled: file.GetPolicy().IsOriginLinkEnable,
				CreateDate:    file.CreatedAt,
			}
			if shareKey != "" {
				newFile.Key = shareKey
			}
			objects = append(objects, newFile)
		}
	}

	return objects
}

// CreateDirectory 根据给定的完整创建目录，支持递归创建。如果目录已存在，则直接
// 返回已存在的目录。
func (fs *FileSystem) CreateDirectory(ctx context.Context, fullPath string) (*model.Folder, error) {
	if fullPath == "." || fullPath == "" {
		return nil, ErrRootProtected
	}

	if fullPath == "/" {
		if fs.Root != nil {
			return fs.Root, nil
		}
		return fs.User.Root()
	}

	// 获取要创建目录的父路径和目录名
	fullPath = path.Clean(fullPath)
	base := path.Dir(fullPath)
	dir := path.Base(fullPath)

	// 去掉结尾空格
	dir = strings.TrimRight(dir, " ")

	// 检查目录名是否合法
	if !fs.ValidateLegalName(ctx, dir) {
		return nil, ErrIllegalObjectName
	}

	// 父目录是否存在
	isExist, parent := fs.IsPathExist(base)
	if !isExist {
		newParent, err := fs.CreateDirectory(ctx, base)
		if err != nil {
			return nil, err
		}
		parent = newParent
	}

	// 是否有同名文件
	if ok, _ := fs.IsChildFileExist(parent, dir); ok {
		return nil, ErrFileExisted
	}

	// 创建目录
	newFolder := model.Folder{
		Name:     dir,
		ParentID: &parent.ID,
		OwnerID:  fs.User.ID,
	}
	_, err := newFolder.Create()

	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return &newFolder, nil
}

// SaveTo 将别人分享的文件转存到目标路径下
func (fs *FileSystem) SaveTo(ctx context.Context, path string) error {
	// 获取父目录
	isExist, folder := fs.IsPathExist(path)
	if !isExist {
		return ErrPathNotExist
	}

	var (
		totalSize uint64
		err       error
	)

	if len(fs.DirTarget) > 0 {
		totalSize, err = fs.DirTarget[0].CopyFolderTo(fs.DirTarget[0].ID, folder)
	} else {
		parent := model.Folder{
			OwnerID: fs.FileTarget[0].UserID,
		}
		parent.ID = fs.FileTarget[0].FolderID
		totalSize, err = parent.MoveOrCopyFileTo([]uint{fs.FileTarget[0].ID}, folder, true)
	}

	// 扣除用户容量
	fs.User.IncreaseStorageWithoutCheck(totalSize)
	if err != nil {
		return ErrFileExisted.WithError(err)
	}

	return nil
}
