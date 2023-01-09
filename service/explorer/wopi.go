package explorer

import (
	"errors"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/middleware"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/wopi"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type WopiService struct {
}

func (service *WopiService) Rename(c *gin.Context) error {
	fs, _, err := service.prepareFs(c)
	if err != nil {
		return err
	}

	defer fs.Recycle()

	return fs.Rename(c, []uint{}, []uint{c.MustGet("object_id").(uint)}, c.GetHeader(wopi.RenameRequestHeader))
}

func (service *WopiService) GetFile(c *gin.Context) error {
	fs, _, err := service.prepareFs(c)
	if err != nil {
		return err
	}

	defer fs.Recycle()

	resp, err := fs.Preview(c, fs.FileTarget[0].ID, true)
	if err != nil {
		return fmt.Errorf("failed to pull file content: %w", err)
	}

	// 重定向到文件源
	if resp.Redirect {
		return fmt.Errorf("redirect not supported in WOPI")
	}

	// 直接返回文件内容
	defer resp.Content.Close()

	c.Header("Cache-Control", "no-cache")
	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, fs.FileTarget[0].UpdatedAt, resp.Content)
	return nil
}

func (service *WopiService) FileInfo(c *gin.Context) (*serializer.WopiFileInfo, error) {
	fs, session, err := service.prepareFs(c)
	if err != nil {
		return nil, err
	}

	defer fs.Recycle()

	parent, err := model.GetFoldersByIDs([]uint{fs.FileTarget[0].FolderID}, fs.User.ID)
	if err != nil {
		return nil, err
	}

	if len(parent) == 0 {
		return nil, fmt.Errorf("failed to find parent folder")
	}

	parent[0].TraceRoot()
	siteUrl := model.GetSiteURL()

	// Generate url for parent folder
	parentUrl := model.GetSiteURL()
	parentUrl.Path = "/home"
	query := parentUrl.Query()
	query.Set("path", parent[0].Position)
	parentUrl.RawQuery = query.Encode()

	info := &serializer.WopiFileInfo{
		BaseFileName:           fs.FileTarget[0].Name,
		Version:                fs.FileTarget[0].Model.UpdatedAt.String(),
		BreadcrumbBrandName:    model.GetSettingByName("siteName"),
		BreadcrumbBrandUrl:     siteUrl.String(),
		FileSharingPostMessage: false,
		PostMessageOrigin:      "*",
		FileNameMaxLength:      256,
		LastModifiedTime:       fs.FileTarget[0].Model.UpdatedAt.Format(time.RFC3339),
		IsAnonymousUser:        true,
		ReadOnly:               true,
		ClosePostMessage:       true,
		Size:                   int64(fs.FileTarget[0].Size),
		OwnerId:                hashid.HashID(fs.FileTarget[0].UserID, hashid.UserID),
	}

	if session.Action == wopi.ActionEdit {
		info.FileSharingPostMessage = true
		info.IsAnonymousUser = false
		info.SupportsRename = true
		info.SupportsReviewing = true
		info.SupportsUpdate = true
		info.UserFriendlyName = fs.User.Nick
		info.UserId = hashid.HashID(fs.User.ID, hashid.UserID)
		info.UserCanRename = true
		info.UserCanReview = true
		info.UserCanWrite = true
		info.ReadOnly = false
		info.BreadcrumbFolderName = parent[0].Name
		info.BreadcrumbFolderUrl = parentUrl.String()
	}

	return info, nil
}

func (service *WopiService) prepareFs(c *gin.Context) (*filesystem.FileSystem, *wopi.SessionCache, error) {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return nil, nil, err
	}

	session := c.MustGet(middleware.WopiSessionCtx).(*wopi.SessionCache)
	if err := fs.SetTargetFileByIDs([]uint{session.FileID}); err != nil {
		fs.Recycle()
		return nil, nil, fmt.Errorf("failed to find file: %w", err)
	}

	maxSize := model.GetIntSetting("maxEditSize", 0)
	if maxSize > 0 && fs.FileTarget[0].Size > uint64(maxSize) {
		return nil, nil, errors.New("file too large")
	}

	return fs, session, nil
}
