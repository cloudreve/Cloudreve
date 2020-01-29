package share

import (
	"context"
	"errors"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/service/explorer"
	"github.com/gin-gonic/gin"
)

// ShareGetService 获取分享服务
type ShareGetService struct {
	Password string `form:"password" binding:"max=255"`
}

// SingleFileService 对单文件进行操作的服务，path为可选文件完整路径
type SingleFileService struct {
	Path string `form:"path" binding:"max=65535"`
}

// Get 获取分享内容
func (service *ShareGetService) Get(c *gin.Context) serializer.Response {
	user := currentUser(c)
	share := model.GetShareByHashID(c.Param("id"))
	if share == nil || !share.IsAvailable() {
		return serializer.Err(serializer.CodeNotFound, "分享不存在或已被取消", nil)
	}

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

	// 如果已经下载过，不需要付积分
	if share.WasDownloadedBy(user, c) {
		share.Score = 0
	}

	return serializer.Response{
		Code: 0,
		Data: serializer.BuildShareResponse(share, unlocked),
	}
}

// CreateDownloadSession 创建下载会话
func (service *SingleFileService) CreateDownloadSession(c *gin.Context) serializer.Response {
	user := currentUser(c)
	share := model.GetShareByHashID(c.Param("id"))
	if share == nil || !share.IsAvailable() {
		return serializer.Err(serializer.CodeNotFound, "分享不存在或已被取消", nil)
	}

	// 检查用户是否可以下载此分享的文件
	err := CheckBeforeGetShare(share, user, c)
	if err != nil {
		return serializer.Err(serializer.CodeNoPermissionErr, err.Error(), nil)
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

	// 取得下载地址
	downloadURL, err := fs.GetDownloadURL(context.Background(), "", "download_timeout")
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
func (service *SingleFileService) PreviewContent(ctx context.Context, c *gin.Context, isText bool) serializer.Response {
	user := currentUser(c)
	share := model.GetShareByHashID(c.Param("id"))
	if share == nil || !share.IsAvailable() {
		return serializer.Err(serializer.CodeNotFound, "分享不存在或已被取消", nil)
	}

	if !share.PreviewEnabled {
		return serializer.Err(serializer.CodeNoPermissionErr, "此分享无法预览", nil)
	}

	// 检查用户是否可以下载此分享的文件
	err := CheckBeforeGetShare(share, user, c)
	if err != nil {
		return serializer.Err(serializer.CodeNoPermissionErr, err.Error(), nil)
	}

	// 用于调下层service
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, share.GetSource())
	subService := explorer.SingleFileService{
		Path: "",
	}

	return subService.PreviewContent(ctx, c, isText)
}

// CreateDocPreviewSession 创建Office预览会话，返回预览地址
func (service *SingleFileService) CreateDocPreviewSession(c *gin.Context) serializer.Response {
	user := currentUser(c)
	share := model.GetShareByHashID(c.Param("id"))
	if share == nil || !share.IsAvailable() {
		return serializer.Err(serializer.CodeNotFound, "分享不存在或已被取消", nil)
	}

	if !share.PreviewEnabled {
		return serializer.Err(serializer.CodeNoPermissionErr, "此分享无法预览", nil)
	}

	// 检查用户是否可以下载此分享的文件
	err := CheckBeforeGetShare(share, user, c)
	if err != nil {
		return serializer.Err(serializer.CodeNoPermissionErr, err.Error(), nil)
	}

	// 用于调下层service
	ctx := context.WithValue(context.Background(), fsctx.FileModelCtx, share.GetSource())
	subService := explorer.SingleFileService{
		Path: "",
	}

	return subService.CreateDocPreviewSession(ctx, c)
}

// CheckBeforeGetShare 获取分享内容/下载前进行的一系列检查
func CheckBeforeGetShare(share *model.Share, user *model.User, c *gin.Context) error {
	// 检查用户是否可以下载此分享的文件
	err := share.CanBeDownloadBy(user)
	if err != nil {
		return err
	}

	// 分享是否已解锁
	if share.Password != "" {
		sessionKey := fmt.Sprintf("share_unlock_%d", share.ID)
		unlocked := util.GetSession(c, sessionKey) != nil
		if !unlocked {
			return errors.New("无权访问此分享")
		}
	}

	// 对积分、下载次数进行更新
	err = share.DownloadBy(user, c)
	if err != nil {
		return err
	}

	return nil
}
