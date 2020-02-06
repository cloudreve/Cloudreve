package aria2

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/aria2"
	"github.com/HFO4/cloudreve/pkg/serializer"
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

// Delete 取消下载任务
func (service *DownloadTaskService) Delete(c *gin.Context) serializer.Response {
	userCtx, _ := c.Get("user")
	user := userCtx.(*model.User)

	// 查找下载记录
	download, err := model.GetDownloadByGid(c.Param("gid"), user.ID)
	if err != nil {
		return serializer.Err(serializer.CodeNotFound, "下载记录不存在", err)
	}

	if download.Status != aria2.Downloading && download.Status != aria2.Paused {
		return serializer.Err(serializer.CodeNoPermissionErr, "此下载任务无法取消", err)
	}

	// 取消任务
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
	if err := aria2.Instance.Select(download, service.Indexes); err != nil {
		return serializer.Err(serializer.CodeNotSet, "操作失败", err)
	}

	return serializer.Response{}

}
