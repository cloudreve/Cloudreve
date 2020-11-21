package admin

import (
	"context"
	"strings"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

// FileService 文件ID服务
type FileService struct {
	ID uint `uri:"id" json:"id" binding:"required"`
}

// FileBatchService 文件批量操作服务
type FileBatchService struct {
	ID    []uint `json:"id" binding:"min=1"`
	Force bool   `json:"force"`
}

// ListFolderService 列目录结构
type ListFolderService struct {
	Path string `uri:"path" binding:"required,max=65535"`
	ID   uint   `uri:"id" binding:"required"`
	Type string `uri:"type" binding:"eq=policy|eq=user"`
}

// List 列出指定路径下的目录
func (service *ListFolderService) List(c *gin.Context) serializer.Response {
	if service.Type == "policy" {
		// 列取存储策略中的目录
		policy, err := model.GetPolicyByID(service.ID)
		if err != nil {
			return serializer.Err(serializer.CodeNotFound, "存储策略不存在", err)
		}

		// 创建文件系统
		fs, err := filesystem.NewAnonymousFileSystem()
		if err != nil {
			return serializer.Err(serializer.CodeInternalSetting, "无法创建文件系统", err)
		}
		defer fs.Recycle()

		// 列取存储策略中的文件
		fs.Policy = &policy
		res, err := fs.ListPhysical(c.Request.Context(), service.Path)
		if err != nil {
			return serializer.Err(serializer.CodeIOFailed, "无法列取目录", err)
		}

		return serializer.Response{
			Data: map[string]interface{}{
				"objects": res,
			},
		}

	}

	// 列取用户空间目录
	// 查找用户
	user, err := model.GetUserByID(service.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(&user)
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "无法创建文件系统", err)
	}
	defer fs.Recycle()

	// 列取目录
	res, err := fs.List(c.Request.Context(), service.Path, nil)
	if err != nil {
		return serializer.Err(serializer.CodeIOFailed, "无法列取目录", err)
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"objects": res,
		},
	}
}

// Delete 删除文件
func (service *FileBatchService) Delete(c *gin.Context) serializer.Response {
	files, err := model.GetFilesByIDs(service.ID, 0)
	if err != nil {
		return serializer.DBErr("无法列出待删除文件", err)
	}

	// 根据用户分组
	userFile := make(map[uint][]model.File)
	for i := 0; i < len(files); i++ {
		if _, ok := userFile[files[i].UserID]; !ok {
			userFile[files[i].UserID] = []model.File{}
		}
		userFile[files[i].UserID] = append(userFile[files[i].UserID], files[i])
	}

	// 异步执行删除
	go func(files map[uint][]model.File) {
		for uid, file := range files {
			user, err := model.GetUserByID(uid)
			if err != nil {
				continue
			}

			fs, err := filesystem.NewFileSystem(&user)
			if err != nil {
				fs.Recycle()
				continue
			}

			// 汇总文件ID
			ids := make([]uint, 0, len(file))
			for i := 0; i < len(file); i++ {
				ids = append(ids, file[i].ID)
			}

			// 执行删除
			fs.Delete(context.Background(), []uint{}, ids, service.Force)
			fs.Recycle()
		}
	}(userFile)

	// 分组执行删除
	return serializer.Response{}

}

// Get 预览文件
func (service *FileService) Get(c *gin.Context) serializer.Response {
	file, err := model.GetFilesByIDs([]uint{service.ID}, 0)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "文件不存在", err)
	}

	ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, &file[0])
	var subService explorer.FileIDService
	res := subService.PreviewContent(ctx, c, false)

	return res
}

// Files 列出文件
func (service *AdminListService) Files() serializer.Response {
	var res []model.File
	total := 0

	tx := model.DB.Model(&model.File{})
	if service.OrderBy != "" {
		tx = tx.Order(service.OrderBy)
	}

	for k, v := range service.Conditions {
		tx = tx.Where(k+" = ?", v)
	}

	if len(service.Searches) > 0 {
		search := ""
		for k, v := range service.Searches {
			search += k + " like '%" + v + "%' OR "
		}
		search = strings.TrimSuffix(search, " OR ")
		tx = tx.Where(search)
	}

	// 计算总数用于分页
	tx.Count(&total)

	// 查询记录
	tx.Limit(service.PageSize).Offset((service.Page - 1) * service.PageSize).Find(&res)

	// 查询对应用户
	users := make(map[uint]model.User)
	for _, file := range res {
		users[file.UserID] = model.User{}
	}

	userIDs := make([]uint, 0, len(users))
	for k := range users {
		userIDs = append(userIDs, k)
	}

	var userList []model.User
	model.DB.Where("id in (?)", userIDs).Find(&userList)

	for _, v := range userList {
		users[v.ID] = v
	}

	return serializer.Response{Data: map[string]interface{}{
		"total": total,
		"items": res,
		"users": users,
	}}
}
