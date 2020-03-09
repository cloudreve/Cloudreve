package admin

import (
	"context"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
	"strings"
)

// FileService 文件ID服务
type FileService struct {
	ID uint `uri:"id" json:"id" binding:"required"`
}

// FileBatchService 文件批量操作服务
type FileBatchService struct {
	ID []uint `json:"id" binding:"min=1"`
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
			fs.Delete(context.Background(), []uint{}, ids)
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
