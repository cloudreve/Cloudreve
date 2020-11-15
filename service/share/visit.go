package share

import (
	"context"
	"fmt"
	"net/http"
	"path"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/cloudreve/Cloudreve/v3/service/explorer"
	"github.com/gin-gonic/gin"
)

// ShareUserGetService 获取用户的分享服务
type ShareUserGetService struct {
	Type string `form:"type" binding:"required,eq=hot|eq=default"`
	Page uint   `form:"page" binding:"required,min=1"`
}

// ShareGetService 获取分享服务
type ShareGetService struct {
	Password string `form:"password" binding:"max=255"`
}

// Service 对分享进行操作的服务，
// path 为可选文件完整路径，在目录分享下有效
type Service struct {
	Path string `form:"path" uri:"path" binding:"max=65535"`
}

// ArchiveService 分享归档下载服务
type ArchiveService struct {
	Path  string   `json:"path" binding:"required,max=65535"`
	Items []string `json:"items"`
	Dirs  []string `json:"dirs"`
}

// ShareListService 列出分享
type ShareListService struct {
	Page     uint   `form:"page" binding:"required,min=1"`
	OrderBy  string `form:"order_by" binding:"required,eq=created_at|eq=downloads|eq=views"`
	Order    string `form:"order" binding:"required,eq=DESC|eq=ASC"`
	Keywords string `form:"keywords"`
}

// Get 获取给定用户的分享
func (service *ShareUserGetService) Get(c *gin.Context) serializer.Response {
	// 取得用户
	userID, _ := c.Get("object_id")
	user, err := model.GetActiveUserByID(userID.(uint))
	if err != nil || user.OptionsSerialized.ProfileOff {
		return serializer.Err(serializer.CodeNotFound, "用户不存在", err)
	}

	// 列出分享
	hotNum := model.GetIntSetting("hot_share_num", 10)
	if service.Type == "default" {
		hotNum = 10
	}
	orderBy := "created_at desc"
	if service.Type == "hot" {
		orderBy = "views desc"
	}
	shares, total := model.ListShares(user.ID, int(service.Page), hotNum, orderBy, true)
	// 列出分享对应的文件
	for i := 0; i < len(shares); i++ {
		shares[i].Source()
	}

	res := serializer.BuildShareList(shares, total)
	res.Data.(map[string]interface{})["user"] = struct {
		ID    string `json:"id"`
		Nick  string `json:"nick"`
		Group string `json:"group"`
		Date  string `json:"date"`
	}{
		hashid.HashID(user.ID, hashid.UserID),
		user.Nick,
		user.Group.Name,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return res
}

// Search 搜索公共分享
func (service *ShareListService) Search(c *gin.Context) serializer.Response {
	// 列出分享
	shares, total := model.SearchShares(int(service.Page), 18, service.OrderBy+" "+
		service.Order, service.Keywords)
	// 列出分享对应的文件
	for i := 0; i < len(shares); i++ {
		shares[i].Source()
	}

	return serializer.BuildShareList(shares, total)
}

// List 列出用户分享
func (service *ShareListService) List(c *gin.Context, user *model.User) serializer.Response {
	// 列出分享
	shares, total := model.ListShares(user.ID, int(service.Page), 18, service.OrderBy+" "+
		service.Order, false)
	// 列出分享对应的文件
	for i := 0; i < len(shares); i++ {
		shares[i].Source()
	}

	return serializer.BuildShareList(shares, total)
}

// Get 获取分享内容
func (service *ShareGetService) Get(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	// 是否已解锁
	unlocked := true
	if share.Password != "" {
		sessionKey := fmt.Sprintf("share_unlock_%d", share.ID)
		unlocked = util.GetSession(c, sessionKey) != nil
		if !unlocked && service.Password != "" {
			// 如果未解锁，且指定了密码，则尝试解锁
			if service.Password == share.Password {
				unlocked = true
				util.SetSession(c, map[string]interface{}{sessionKey: true})
			}
		}
	}

	if unlocked {
		share.Viewed()
	}

	return serializer.Response{
		Code: 0,
		Data: serializer.BuildShareResponse(share, unlocked),
	}
}

// CreateDownloadSession 创建下载会话
func (service *Service) CreateDownloadSession(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重设文件系统处理目标为源文件
	err = fs.SetTargetByInterface(share.Source())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, "源文件不存在", err)
	}

	ctx := context.Background()

	// 重设根目录
	if share.IsDir {
		fs.Root = &fs.DirTarget[0]

		// 找到目标文件
		err = fs.ResetFileIfNotExist(ctx, service.Path)
		if err != nil {
			return serializer.Err(serializer.CodeNotSet, err.Error(), err)
		}
	}

	// 取得下载地址
	downloadURL, err := fs.GetDownloadURL(ctx, 0, "download_timeout")
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: downloadURL,
	}
}

// PreviewContent 预览文件，需要登录会话, isText - 是否为文本文件，文本文件会
// 强制经由服务端中转
func (service *Service) PreviewContent(ctx context.Context, c *gin.Context, isText bool) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	// 用于调下层service
	if share.IsDir {
		ctx = context.WithValue(ctx, fsctx.FolderModelCtx, share.Source())
		ctx = context.WithValue(ctx, fsctx.PathCtx, service.Path)
	} else {
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, share.Source())
	}
	subService := explorer.FileIDService{}

	return subService.PreviewContent(ctx, c, isText)
}

// CreateDocPreviewSession 创建Office预览会话，返回预览地址
func (service *Service) CreateDocPreviewSession(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	// 用于调下层service
	ctx := context.Background()
	if share.IsDir {
		ctx = context.WithValue(ctx, fsctx.FolderModelCtx, share.Source())
		ctx = context.WithValue(ctx, fsctx.PathCtx, service.Path)
	} else {
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, share.Source())
	}
	subService := explorer.FileIDService{}

	return subService.CreateDocPreviewSession(ctx, c)
}

// List 列出分享的目录下的对象
func (service *Service) List(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	if !share.IsDir {
		return serializer.ParamErr("此分享无法列目录", nil)
	}

	if !path.IsAbs(service.Path) {
		return serializer.ParamErr("路径无效", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(share.Creator())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 重设根目录
	fs.Root = share.Source().(*model.Folder)
	fs.Root.Name = "/"

	// 分享Key上下文
	ctx = context.WithValue(ctx, fsctx.ShareKeyCtx, hashid.HashID(share.ID, hashid.ShareID))

	// 获取子项目
	objects, err := fs.List(ctx, service.Path, nil)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFolderFailed, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: map[string]interface{}{
			"parent":  "0000",
			"objects": objects,
		},
	}
}

// Thumb 获取被分享文件的缩略图
func (service *Service) Thumb(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	if !share.IsDir {
		return serializer.ParamErr("此分享无缩略图", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(share.Creator())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重设根目录
	fs.Root = share.Source().(*model.Folder)

	// 找到缩略图的父目录
	exist, parent := fs.IsPathExist(service.Path)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "路径不存在", nil)
	}

	ctx := context.WithValue(context.Background(), fsctx.LimitParentCtx, parent)

	// 获取文件ID
	fileID, err := hashid.DecodeHashID(c.Param("file"), hashid.FileID)
	if err != nil {
		return serializer.ParamErr("无法解析文件ID", err)
	}

	// 获取缩略图
	resp, err := fs.GetThumb(ctx, uint(fileID))
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法获取缩略图", err)
	}

	if resp.Redirect {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", resp.MaxAge))
		c.Redirect(http.StatusMovedPermanently, resp.URL)
		return serializer.Response{Code: -1}
	}

	defer resp.Content.Close()
	http.ServeContent(c.Writer, c.Request, "thumb.png", fs.FileTarget[0].UpdatedAt, resp.Content)

	return serializer.Response{Code: -1}

}

// Archive 创建批量下载归档
func (service *ArchiveService) Archive(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 是否有权限
	if !user.Group.OptionsSerialized.ArchiveDownload {
		return serializer.Err(serializer.CodeNoPermissionErr, "您的用户组无权进行此操作", nil)
	}

	if !share.IsDir {
		return serializer.ParamErr("此分享无法进行打包", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重设根目录
	fs.Root = share.Source().(*model.Folder)

	// 找到要打包文件的父目录
	exist, parent := fs.IsPathExist(service.Path)
	if !exist {
		return serializer.Err(serializer.CodeNotFound, "路径不存在", nil)
	}

	// 限制操作范围为父目录下
	ctx := context.WithValue(context.Background(), fsctx.LimitParentCtx, parent)

	// 用于调下层service
	tempUser := share.Creator()
	tempUser.Group.OptionsSerialized.ArchiveDownload = true
	c.Set("user", tempUser)

	subService := explorer.ItemIDService{
		Dirs:  service.Dirs,
		Items: service.Items,
	}

	return subService.Archive(ctx, c)
}
