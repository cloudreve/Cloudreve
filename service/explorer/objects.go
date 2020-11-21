package explorer

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"path"
	"strings"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
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
	Items  []string `json:"items"`
	Dirs   []string `json:"dirs"`
	Source *ItemService
}

// ItemCompressService 文件压缩任务服务
type ItemCompressService struct {
	Src  ItemIDService `json:"src"`
	Dst  string        `json:"dst" binding:"required,min=1,max=65535"`
	Name string        `json:"name" binding:"required,min=1,max=255"`
}

// ItemDecompressService 文件解压缩任务服务
type ItemDecompressService struct {
	Src string `json:"src"`
	Dst string `json:"dst" binding:"required,min=1,max=65535"`
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
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 检查用户组权限
	if !fs.User.Group.OptionsSerialized.ArchiveTask {
		return serializer.Err(serializer.CodeGroupNotAllowed, "当前用户组无法进行此操作", nil)
	}

	// 存放目录是否存在
	if exist, _ := fs.IsPathExist(service.Dst); !exist {
		return serializer.Err(serializer.CodeNotFound, "存放路径不存在", nil)
	}

	// 压缩包是否存在
	exist, file := fs.IsFileExist(service.Src)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "文件不存在", nil)
	}

	// 文件尺寸限制
	if fs.User.Group.OptionsSerialized.DecompressSize != 0 && file.Size > fs.User.Group.
		OptionsSerialized.DecompressSize {
		return serializer.Err(serializer.CodeParamErr, "文件太大", nil)
	}

	// 必须是zip压缩包
	if !strings.HasSuffix(file.Name, ".zip") {
		return serializer.Err(serializer.CodeParamErr, "只能解压 ZIP 格式的压缩文件", nil)
	}

	// 创建任务
	job, err := task.NewDecompressTask(fs.User, service.Src, service.Dst)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}
	task.TaskPoll.Submit(job)

	return serializer.Response{}

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

	// 补齐压缩文件扩展名（如果没有）
	if !strings.HasSuffix(service.Name, ".zip") {
		service.Name += ".zip"
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
	folders, err := model.GetRecursiveChildFolder(service.Src.Raw().Dirs, fs.User.ID, true)
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

	// 文件尺寸限制
	if fs.User.Group.OptionsSerialized.DecompressSize != 0 && totalSize > fs.User.Group.
		OptionsSerialized.CompressSize {
		return serializer.Err(serializer.CodeParamErr, "文件太大", nil)
	}

	// 按照平均压缩率计算用户空间是否足够
	compressRatio := 0.4
	spaceNeeded := uint64(math.Round(float64(totalSize) * compressRatio))
	if fs.User.GetRemainingCapacity() < spaceNeeded {
		return serializer.Err(serializer.CodeParamErr, "剩余空间不足", err)
	}

	// 创建任务
	job, err := task.NewCompressTask(fs.User, path.Join(service.Dst, service.Name), service.Src.Raw().Dirs,
		service.Src.Raw().Items)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}
	task.TaskPoll.Submit(job)

	return serializer.Response{}

}

// Archive 创建归档
func (service *ItemIDService) Archive(ctx context.Context, c *gin.Context) serializer.Response {
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
	items := service.Raw()
	zipFile, err := fs.Compress(ctx, items.Dirs, items.Items, true)
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
func (service *ItemIDService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 删除对象
	items := service.Raw()
	err = fs.Delete(ctx, items.Dirs, items.Items, false)
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
		return serializer.ParamErr("只能复制一个对象", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
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
		return serializer.ParamErr("只能操作一个对象", nil)
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
