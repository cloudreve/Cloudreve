package share

import (
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/hashid"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
	"path"
)

// ShareGetService 获取分享服务
type ShareGetService struct {
	Password string `form:"password" binding:"max=255"`
}

// ShareService 对分享进行操作的服务，
// path 为可选文件完整路径，在目录分享下有效
type ShareService struct {
	Path string `form:"path" uri:"path" binding:"max=65535"`
}

// Get 获取分享内容
func (service *ShareGetService) Get(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

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

	// 如果已经下载过或者是自己的分享，不需要付积分
	if share.UserID == user.ID || share.WasDownloadedBy(user, c) {
		share.Score = 0
	}

	return serializer.Response{
		Code: 0,
		Data: serializer.BuildShareResponse(share, unlocked),
	}
}

// CreateDownloadSession 创建下载会话
func (service *ShareService) CreateDownloadSession(c *gin.Context) serializer.Response {
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
	err = fs.SetTargetByInterface(share.GetSource())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, "源文件不存在", err)
	}

	// 重设根目录
	if share.IsDir {
		fs.Root = &fs.DirTarget[0]
	}

	// 取得下载地址
	downloadURL, err := fs.GetDownloadURL(context.Background(), service.Path, "download_timeout")
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
func (service *ShareService) PreviewContent(ctx context.Context, c *gin.Context, isText bool) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	// 用于调下层service
	if share.IsDir {
		ctx = context.WithValue(ctx, fsctx.FolderModelCtx, share.GetSource())
	} else {
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, share.GetSource())
	}
	subService := explorer.SingleFileService{
		Path: service.Path,
	}

	return subService.PreviewContent(ctx, c, isText)
}

// CreateDocPreviewSession 创建Office预览会话，返回预览地址
func (service *ShareService) CreateDocPreviewSession(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	// 用于调下层service
	ctx := context.Background()
	if share.IsDir {
		ctx = context.WithValue(ctx, fsctx.FolderModelCtx, share.GetSource())
	} else {
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, share.GetSource())
	}
	subService := explorer.SingleFileService{
		Path: service.Path,
	}

	return subService.CreateDocPreviewSession(ctx, c)
}

// SaveToMyFile 将此分享转存到自己的网盘
func (service *ShareService) SaveToMyFile(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 不能转存自己的文件
	if share.UserID == user.ID {
		return serializer.Err(serializer.CodePolicyNotAllowed, "不能转存自己的分享", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(user)
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 重设文件系统处理目标为源文件
	err = fs.SetTargetByInterface(share.GetSource())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, "源文件不存在", err)
	}

	err = fs.SaveTo(context.Background(), service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{}
}

// List 列出分享的目录下的对象
func (service *ShareService) List(c *gin.Context) serializer.Response {
	shareCtx, _ := c.Get("share")
	share := shareCtx.(*model.Share)

	if !share.IsDir {
		return serializer.ParamErr("此分享无法列目录", nil)
	}

	if !path.IsAbs(service.Path) {
		return serializer.ParamErr("路径无效", nil)
	}

	// 创建文件系统
	fs, err := filesystem.NewFileSystem(share.GetCreator())
	if err != nil {
		return serializer.Err(serializer.CodePolicyNotAllowed, err.Error(), err)
	}
	defer fs.Recycle()

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 重设根目录
	fs.Root = share.GetSource().(*model.Folder)
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
		Data: objects,
	}
}
