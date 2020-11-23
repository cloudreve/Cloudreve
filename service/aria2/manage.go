package aria2

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// SelectFileService 选择要下载的文件服务
type SelectFileService struct {
	Indexes []int `json:"indexes" binding:"required"`
}

// DownloadTaskService 下载任务管理服务
type DownloadTaskService struct {
	GID string `uri:"gid" binding:"required"`
}

// DownloadListService 下载列表服务
type DownloadListService struct {
	Page uint `form:"page"`
}

// Finished 获取已完成的任务
func (service *DownloadListService) Finished(c *gin.Context, user *model.User) serializer.Response {
	// 查找下载记录
	downloads := model.GetDownloadsByStatusAndUser(service.Page, user.ID, aria2.Error, aria2.Complete, aria2.Canceled, aria2.Unknown)
	return serializer.BuildFinishedListResponse(downloads)
}

// Downloading 获取正在下载中的任务
func (service *DownloadListService) Downloading(c *gin.Context, user *model.User) serializer.Response {
	// 查找下载记录
	downloads := model.GetDownloadsByStatusAndUser(service.Page, user.ID, aria2.Downloading, aria2.Paused, aria2.Ready)
	return serializer.BuildDownloadingResponse(downloads)
}

// Delete 取消或删除下载任务
func (service *DownloadTaskService) Delete(c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 查找下载记录
	download, err := model.GetDownloadByGid(c.Param("gid"), user.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "下载记录不存在", err)
	}

	if download.Status >= aria2.Error {
		// 如果任务已完成，则删除任务记录
		if err := download.Delete(); err != nil {
			return serializer.Err(serializer.CodeDBError, "任务记录删除失败", err)
		}
		return serializer.Response{}
	}

	// 取消任务
	aria2.Lock.RLock()
	defer aria2.Lock.RUnlock()
	if err := aria2.Instance.Cancel(download); err != nil {
		return serializer.Err(serializer.CodeNotSet, "操作失败", err)
	}

	return serializer.Response{}
}

// Select 选取要下载的文件
func (service *SelectFileService) Select(c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 查找下载记录
	download, err := model.GetDownloadByGid(c.Param("gid"), user.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "下载记录不存在", err)
	}

	if download.StatusInfo.BitTorrent.Mode != "multi" || (download.Status != aria2.Downloading && download.Status != aria2.Paused) {
		return serializer.Err(serializer.CodeNoPermissionErr, "此下载任务无法选取文件", err)
	}

	// 选取下载
	aria2.Lock.RLock()
	defer aria2.Lock.RUnlock()
	if err := aria2.Instance.Select(download, service.Indexes); err != nil {
		return serializer.Err(serializer.CodeNotSet, "操作失败", err)
	}

	return serializer.Response{}

}
