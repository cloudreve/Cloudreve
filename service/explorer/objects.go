package explorer

import (
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/cache"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/task"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
	"math"
	"net/url"
	"path"
	"time"
)

// ItemMoveService 处理多文件/目录移动
type ItemMoveService struct {
	SrcDir string      `json:"src_dir" binding:"required,min=1,max=65535"`
	Src    ItemService `json:"src" binding:"exists"`
	Dst    string      `json:"dst" binding:"required,min=1,max=65535"`
}

// ItemRenameService 处理多文件/目录重命名
type ItemRenameService struct {
	Src     ItemService `json:"src" binding:"exists"`
	NewName string      `json:"new_name" binding:"required,min=1,max=255"`
}

// ItemService 处理多文件/目录相关服务
type ItemService struct {
	Items []uint `json:"items" binding:"exists"`
	Dirs  []uint `json:"dirs" binding:"exists"`
}

// ItemCompressService 文件压缩任务服务
type ItemCompressService struct {
	Src  ItemService `json:"src" binding:"exists"`
	Dst  string      `json:"dst" binding:"required,min=1,max=65535"`
	Name string      `json:"name" binding:"required,min=1,max=255"`
}

// CreateCompressTask 创建文件压缩任务
func (service *ItemCompressService) CreateCompressTask(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveTask {
		return serializer.Err(serializer.CodeGroupNotAllowed, "当前用户组无法进行此操作", nil)
	}

	// 存放目录是否存在，是否重名
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeNotFound, "存放路径不存在", nil)
	}
	if exist, _ := fs.IsFileExist(path.Join(service.Dst, service.Name)); exist {
		return serializer.ParamErr("名为 "+service.Name+" 的文件已存在", nil)
	}

	// 检查文件名合法性
	if !fs.ValidateLegalName(context.Background(), service.Name) {
		return serializer.ParamErr("文件名非法", nil)
	}
	if !fs.ValidateExtension(context.Background(), service.Name) {
		return serializer.ParamErr("不允许存储此扩展名的文件", nil)
	}

	// 递归列出待压缩子目录
	folders, err := model.GetRecursiveChildFolder(service.Src.Dirs, fs.User.ID, true)
	if err != nil {
		return serializer.Err(serializer.CodeDBError, "无法列出子目录", err)
	}

	// 列出所有待压缩文件
	files, err := model.GetChildFilesOfFolders(&folders)
	if err != nil {
		return serializer.Err(serializer.CodeDBError, "无法列出子文件", err)
	}

	// 计算待压缩文件大小
	var totalSize uint64
	for i := 0; i < len(files); i++ {
		totalSize += files[i].Size
	}

	// 按照平均压缩率计算用户空间是否足够
	compressRatio := 0.4
	spaceNeeded := uint64(math.Round(float64(totalSize) * compressRatio))
	if fs.User.GetRemainingCapacity() < spaceNeeded {
		return serializer.Err(serializer.CodeParamErr, "剩余空间不足", err)
	}

	// 创建任务
	job, err := task.NewCompressTask(fs.User, path.Join(service.Dst, service.Name), service.Src.Dirs,
		service.Src.Items)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}

	task.TaskPoll.Submit(job)
	return serializer.Response{}

}

// Archive 创建归档
func (service *ItemService) Archive(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveDownload {
		return serializer.Err(serializer.CodeGroupNotAllowed, "当前用户组无法进行此操作", nil)
	}

	// 开始压缩
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	zipFile, err := fs.Compress(ctx, service.Dirs, service.Items, true)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法创建压缩文件", err)
	}

	// 生成一次性压缩文件下载地址
	siteURL, err := url.Parse(model.GetSettingByName("siteURL"))
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法解析站点URL", err)
	}
	zipID := util.RandStringRunes(16)
	ttl := model.GetIntSetting("archive_timeout", 30)
	signedURI, err := auth.SignURI(
		auth.General,
		fmt.Sprintf("/api/v3/file/archive/%s/archive.zip", zipID),
		time.Now().Unix()+int64(ttl),
	)
	finalURL := siteURL.ResolveReference(signedURI).String()

	// 将压缩文件记录存入缓存
	err = cache.Set("archive_"+zipID, zipFile, ttl)
	if err != nil {
		return serializer.Err(serializer.CodeIOFailed, "无法写入缓存", err)
	}

	return serializer.Response{
		Code: 0,
		Data: finalURL,
	}
}

// Delete 删除对象
func (service *ItemService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 删除对象
	err = fs.Delete(ctx, service.Dirs, service.Items)
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
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 移动对象
	err = fs.Move(ctx, service.Src.Dirs, service.Src.Items, service.SrcDir, service.Dst)
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
		return serializer.ParamErr("只能复制一个对象", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 复制对象
	err = fs.Copy(ctx, service.Src.Dirs, service.Src.Items, service.SrcDir, service.Dst)
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
		return serializer.ParamErr("只能操作一个对象", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重命名对象
	err = fs.Rename(ctx, service.Src.Dirs, service.Src.Items, service.NewName)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
	}
}
