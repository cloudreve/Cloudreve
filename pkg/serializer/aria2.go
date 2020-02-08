package serializer

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/zyxar/argo/rpc"
	"path"
)

// DownloadListResponse 下载列表响应条目
type DownloadListResponse struct {
	UpdateTime int64          `json:"update"`
	Name       string         `json:"name"`
	Status     int            `json:"status"`
	UserID     uint           `json:"uid"`
	Error      string         `json:"error"`
	Dst        string         `json:"dst"`
	Total      uint64         `json:"total"`
	Downloaded uint64         `json:"downloaded"`
	Speed      int            `json:"speed"`
	Info       rpc.StatusInfo `json:"info"`
}

// BuildDownloadingResponse 构建正在下载的列表响应
func BuildDownloadingResponse(tasks []model.Download) Response {
	resp := make([]DownloadListResponse, 0, len(tasks))

	for i := 0; i < len(tasks); i++ {
		fileName := ""
		if len(tasks[i].StatusInfo.Files) > 0 {
			fileName = path.Base(tasks[i].StatusInfo.Files[0].Path)
		}

		// 过滤敏感信息
		tasks[i].StatusInfo.Dir = ""
		for i2 := 0; i2 < len(tasks[i].StatusInfo.Files); i2++ {
			tasks[i].StatusInfo.Files[i2].Path = path.Base(tasks[i].StatusInfo.Files[i2].Path)
		}

		resp = append(resp, DownloadListResponse{
			UpdateTime: tasks[i].UpdatedAt.Unix(),
			Name:       fileName,
			Status:     tasks[i].Status,
			UserID:     tasks[i].UserID,
			Error:      tasks[i].Error,
			Dst:        tasks[i].Dst,
			Total:      tasks[i].TotalSize,
			Downloaded: tasks[i].DownloadedSize,
			Speed:      tasks[i].Speed,
			Info:       tasks[i].StatusInfo,
		})
	}

	return Response{Data: resp}
}
