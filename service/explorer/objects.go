package explorer

import (
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// ItemMoveService 处理多文件/目录移动
type ItemMoveService struct {
	SrcDir string        `json:"src_dir" binding:"required,min=1,max=65535"`
	Src    ItemIDService `json:"src"`
	Dst    string        `json:"dst" binding:"required,min=1,max=65535"`
}

// ItemRenameService 处理多文件/目录重命名
type ItemRenameService struct {
	Src     ItemIDService `json:"src"`
	NewName string        `json:"new_name" binding:"required,min=1,max=255"`
}

// ItemService 处理多文件/目录相关服务
type ItemService struct {
	Items []uint `json:"items"`
	Dirs  []uint `json:"dirs"`
}

// ItemIDService 处理多文件/目录相关服务，字段值为HashID，可通过Raw()方法获取原始ID
type ItemIDService struct {
	Items      []string `json:"items"`
	Dirs       []string `json:"dirs"`
	Source     *ItemService
	Force      bool `json:"force"`
	UnlinkOnly bool `json:"unlink"`
}

// ItemCompressService 文件压缩任务服务
type ItemCompressService struct {
	Src  ItemIDService `json:"src"`
	Dst  string        `json:"dst" binding:"required,min=1,max=65535"`
	Name string        `json:"name" binding:"required,min=1,max=255"`
}

// ItemRelocateService 文件转移任务服务
type ItemRelocateService struct {
	Src         ItemIDService `json:"src"`
	DstPolicyID string        `json:"dst_policy_id" binding:"required"`
}

// ItemDecompressService 文件解压缩任务服务
type ItemDecompressService struct {
	Src      string `json:"src"`
	Dst      string `json:"dst" binding:"required,min=1,max=65535"`
	Encoding string `json:"encoding"`
}

// ItemPropertyService 获取对象属性服务
type ItemPropertyService struct {
	ID        string `binding:"required"`
	TraceRoot bool   `form:"trace_root"`
	IsFolder  bool   `form:"is_folder"`
}

func init() {
	gob.Register(ItemIDService{})
}

// Raw 批量解码HashID，获取原始ID
func (service *ItemIDService) Raw() *ItemService {
	if service.Source != nil {
		return service.Source
	}

	service.Source = &ItemService{
		Dirs:  make([]uint, 0, len(service.Dirs)),
		Items: make([]uint, 0, len(service.Items)),
	}
	for _, folder := range service.Dirs {
		id, err := hashid.DecodeHashID(folder, hashid.FolderID)
		if err == nil {
			service.Source.Dirs = append(service.Source.Dirs, id)
		}
	}
	for _, file := range service.Items {
		id, err := hashid.DecodeHashID(file, hashid.FileID)
		if err == nil {
			service.Source.Items = append(service.Source.Items, id)
		}
	}

	return service.Source
}

// CreateDecompressTask 创建文件解压缩任务
func (service *ItemDecompressService) CreateDecompressTask(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveTask {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	// 存放目录是否存在
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeParentNotExist, "", nil)
	}

	// 压缩包是否存在
	exist, file := fs.IsFileExist(service.Src)
	if !exist {
		return serializer.Err(serializer.CodeFileNotFound, "", nil)
	}

	// 文件尺寸限制
	if fs.User.Group.OptionsSerialized.DecompressSize != 0 && file.Size > fs.User.Group.
		OptionsSerialized.DecompressSize {
		return serializer.Err(serializer.CodeFileTooLarge, "", nil)
	}

	// 支持的压缩格式后缀
	var (
		suffixes = []string{".zip", ".gz", ".xz", ".tar", ".rar"}
		matched  bool
	)
	for _, suffix := range suffixes {
		if strings.HasSuffix(file.Name, suffix) {
			matched = true
			break
		}
	}
	if !matched {
		return serializer.Err(serializer.CodeUnsupportedArchiveType, "", nil)
	}

	// 创建任务
	job, err := task.NewDecompressTask(fs.User, service.Src, service.Dst, service.Encoding)
	if err != nil {
		return serializer.Err(serializer.CodeCreateTaskError, "", err)
	}
	task.TaskPoll.Submit(job)

	return serializer.Response{}

}

// CreateCompressTask 创建文件压缩任务
func (service *ItemCompressService) CreateCompressTask(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveTask {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	// 补齐压缩文件扩展名（如果没有）
	if !strings.HasSuffix(service.Name, ".zip") {
		service.Name += ".zip"
	}

	// 存放目录是否存在，是否重名
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeParentNotExist, "", nil)
	}
	if exist, _ := fs.IsFileExist(path.Join(service.Dst, service.Name)); exist {
		return serializer.ParamErr("File "+service.Name+" already exist", nil)
	}

	// 检查文件名合法性
	if !fs.ValidateLegalName(context.Background(), service.Name) {
		return serializer.Err(serializer.CodeIllegalObjectName, "", nil)
	}
	if !fs.ValidateExtension(context.Background(), service.Name) {
		return serializer.Err(serializer.CodeFileTypeNotAllowed, "", nil)
	}

	// 递归列出待压缩子目录
	folders, err := model.GetRecursiveChildFolder(service.Src.Raw().Dirs, fs.User.ID, true)
	if err != nil {
		return serializer.DBErr("Failed to list folders", err)
	}

	// 列出所有待压缩文件
	files, err := model.GetChildFilesOfFolders(&folders)
	if err != nil {
		return serializer.DBErr("Failed to list files", err)
	}

	// 计算待压缩文件大小
	var totalSize uint64
	for i := 0; i < len(files); i++ {
		totalSize += files[i].Size
	}

	// 文件尺寸限制
	if fs.User.Group.OptionsSerialized.CompressSize != 0 && totalSize > fs.User.Group.
		OptionsSerialized.CompressSize {
		return serializer.Err(serializer.CodeFileTooLarge, "", nil)
	}

	// 按照平均压缩率计算用户空间是否足够
	compressRatio := 0.4
	spaceNeeded := uint64(math.Round(float64(totalSize) * compressRatio))
	if fs.User.GetRemainingCapacity() < spaceNeeded {
		return serializer.Err(serializer.CodeInsufficientCapacity, "", err)
	}

	// 创建任务
	job, err := task.NewCompressTask(fs.User, path.Join(service.Dst, service.Name), service.Src.Raw().Dirs,
		service.Src.Raw().Items)
	if err != nil {
		return serializer.Err(serializer.CodeCreateTaskError, "", err)
	}
	task.TaskPoll.Submit(job)

	return serializer.Response{}

}

// CreateRelocateTask 创建文件转移任务
func (service *ItemRelocateService) CreateRelocateTask(c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 取得存储策略的ID
	rawID, err := hashid.DecodeHashID(service.DstPolicyID, hashid.PolicyID)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotExist, "", err)
	}

	// 检查用户组权限
	if !user.Group.OptionsSerialized.Relocate {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	// 用户是否可以使用目的存储策略
	if !util.ContainsUint(user.Group.PolicyList, rawID) {
		return serializer.ParamErr("Storage policy is not available", nil)
	}

	// 查找存储策略
	if _, err := model.GetPolicyByID(rawID); err != nil {
		return serializer.ParamErr("Storage policy is not available", nil)
	}

	// 查找是否有正在进行中的转存任务
	var tasks []model.Task
	model.DB.Where("status in (?) and user_id = ? and type = ?",
		[]int{task.Queued, task.Processing}, user.ID,
		task.RelocateTaskType).Find(&tasks)
	if len(tasks) > 0 {
		return serializer.Response{
			Code: serializer.CodeConflict,
			Msg:  "There's ongoing relocate task, please wait for the previous task to finish",
		}
	}

	IDRaw := service.Src.Raw()

	// 创建任务
	job, err := task.NewRelocateTask(user, rawID, IDRaw.Dirs,
		IDRaw.Items)
	if err != nil {
		return serializer.Err(serializer.CodeCreateTaskError, "", err)
	}
	task.TaskPoll.Submit(job)

	return serializer.Response{}
}

// Archive 创建归档
func (service *ItemIDService) Archive(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveDownload {
		return serializer.Err(serializer.CodeGroupNotAllowed, "", nil)
	}

	// 创建打包下载会话
	ttl := model.GetIntSetting("archive_timeout", 30)
	downloadSessionID := util.RandStringRunes(16)
	cache.Set("archive_"+downloadSessionID, *service, ttl)
	cache.Set("archive_user_"+downloadSessionID, *fs.User, ttl)
	signURL, err := auth.SignURI(
		auth.General,
		fmt.Sprintf("/api/v3/file/archive/%s/archive.zip", downloadSessionID),
		int64(ttl),
	)

	return serializer.Response{
		Code: 0,
		Data: signURL.String(),
	}
}

// Delete 删除对象
func (service *ItemIDService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	force, unlink := false, false
	if fs.User.Group.OptionsSerialized.AdvanceDelete {
		force = service.Force
		unlink = service.UnlinkOnly
	}

	// 删除对象
	items := service.Raw()
	err = fs.Delete(ctx, items.Dirs, items.Items, force, unlink)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}

}

// Move 移动对象
func (service *ItemMoveService) Move(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 移动对象
	items := service.Src.Raw()
	err = fs.Move(ctx, items.Dirs, items.Items, service.SrcDir, service.Dst)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}

}

// Copy 复制对象
func (service *ItemMoveService) Copy(ctx context.Context, c *gin.Context) serializer.Response {
	// 复制操作只能对一个目录或文件对象进行操作
	if len(service.Src.Items)+len(service.Src.Dirs) > 1 {
		return filesystem.ErrOneObjectOnly
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 复制对象
	err = fs.Copy(ctx, service.Src.Raw().Dirs, service.Src.Raw().Items, service.SrcDir, service.Dst)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}

}

// Rename 重命名对象
func (service *ItemRenameService) Rename(ctx context.Context, c *gin.Context) serializer.Response {
	// 重命名作只能对一个目录或文件对象进行操作
	if len(service.Src.Items)+len(service.Src.Dirs) > 1 {
		return filesystem.ErrOneObjectOnly
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重命名对象
	err = fs.Rename(ctx, service.Src.Raw().Dirs, service.Src.Raw().Items, service.NewName)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}
}

// GetProperty 获取对象的属性
func (service *ItemPropertyService) GetProperty(ctx context.Context, c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	var props serializer.ObjectProps
	props.QueryDate = time.Now()

	// 如果是文件对象
	if !service.IsFolder {
		res, err := hashid.DecodeHashID(service.ID, hashid.FileID)
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, "", err)
		}

		file, err := model.GetFilesByIDs([]uint{res}, user.ID)
		if err != nil {
			return serializer.DBErr("Failed to query file records", err)
		}

		props.CreatedAt = file[0].CreatedAt
		props.UpdatedAt = file[0].UpdatedAt
		props.Policy = file[0].GetPolicy().Name
		props.Size = file[0].Size

		// 查找父目录
		if service.TraceRoot {
			parent, err := model.GetFoldersByIDs([]uint{file[0].FolderID}, user.ID)
			if err != nil {
				return serializer.DBErr("Parent folder record not exist", err)
			}

			if err := parent[0].TraceRoot(); err != nil {
				return serializer.DBErr("Failed to trace root folder", err)
			}

			props.Path = path.Join(parent[0].Position, parent[0].Name)
		}
	} else {
		res, err := hashid.DecodeHashID(service.ID, hashid.FolderID)
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, "", err)
		}

		folder, err := model.GetFoldersByIDs([]uint{res}, user.ID)
		if err != nil {
			return serializer.DBErr("Failed to query folder records", err)
		}

		policy := user.GetPolicyID(&folder[0])
		if folder[0].PolicyID > 0 {
			props.Policy = policy.Name
		}
		props.CreatedAt = folder[0].CreatedAt
		props.UpdatedAt = folder[0].UpdatedAt

		// 如果对象是目录, 先尝试返回缓存结果
		if cacheRes, ok := cache.Get(fmt.Sprintf("folder_props_%d", res)); ok {
			res := cacheRes.(serializer.ObjectProps)
			res.CreatedAt = props.CreatedAt
			res.UpdatedAt = props.UpdatedAt
			res.Policy = props.Policy
			return serializer.Response{Data: res}
		}

		// 统计子目录
		childFolders, err := model.GetRecursiveChildFolder([]uint{folder[0].ID},
			user.ID, true)
		if err != nil {
			return serializer.DBErr("Failed to list child folders", err)
		}
		props.ChildFolderNum = len(childFolders) - 1

		// 统计子文件
		files, err := model.GetChildFilesOfFolders(&childFolders)
		if err != nil {
			return serializer.DBErr("Failed to list child files", err)
		}

		// 统计子文件个数和大小
		props.ChildFileNum = len(files)
		for i := 0; i < len(files); i++ {
			props.Size += files[i].Size
		}

		// 查找父目录
		if service.TraceRoot {
			if err := folder[0].TraceRoot(); err != nil {
				return serializer.DBErr("Failed to list child folders", err)
			}

			props.Path = folder[0].Position
		}

		// 如果列取对象是目录，则缓存结果
		cache.Set(fmt.Sprintf("folder_props_%d", res), props,
			model.GetIntSetting("folder_props_timeout", 300))
	}

	return serializer.Response{
		Code: 0,
		Data: props,
	}
}
