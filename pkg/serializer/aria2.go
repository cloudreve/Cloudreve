package serializer

import (
	"path"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
)

// DownloadListResponse 下载列表响应条目
type DownloadListResponse struct {
	UpdateTime     int64          `json:"update"`
	UpdateInterval int            `json:"interval"`
	Name           string         `json:"name"`
	Status         int            `json:"status"`
	Dst            string         `json:"dst"`
	Total          uint64         `json:"total"`
	Downloaded     uint64         `json:"downloaded"`
	Speed          int            `json:"speed"`
	Info           rpc.StatusInfo `json:"info"`
}

// FinishedListResponse 已完成任务条目
type FinishedListResponse struct {
	Name       string         `json:"name"`
	GID        string         `json:"gid"`
	Status     int            `json:"status"`
	Dst        string         `json:"dst"`
	Error      string         `json:"error"`
	Total      uint64         `json:"total"`
	Files      []rpc.FileInfo `json:"files"`
	TaskStatus int            `json:"task_status"`
	TaskError  string         `json:"task_error"`
	CreateTime string         `json:"create"`
	UpdateTime string         `json:"update"`
}

// BuildFinishedListResponse 构建已完成任务条目
func BuildFinishedListResponse(tasks []model.Download) Response {
	resp := make([]FinishedListResponse, 0, len(tasks))

	for i := 0; i < len(tasks); i++ {
		fileName := tasks[i].StatusInfo.BitTorrent.Info.Name
		if len(tasks[i].StatusInfo.Files) == 1 {
			fileName = path.Base(tasks[i].StatusInfo.Files[0].Path)
		}

		// 过滤敏感信息
		for i2 := 0; i2 < len(tasks[i].StatusInfo.Files); i2++ {
			tasks[i].StatusInfo.Files[i2].Path = path.Base(tasks[i].StatusInfo.Files[i2].Path)
		}

		download := FinishedListResponse{
			Name:       fileName,
			GID:        tasks[i].GID,
			Status:     tasks[i].Status,
			Error:      tasks[i].Error,
			Dst:        tasks[i].Dst,
			Total:      tasks[i].TotalSize,
			Files:      tasks[i].StatusInfo.Files,
			TaskStatus: -1,
			UpdateTime: tasks[i].UpdatedAt.Format("2006-01-02 15:04:05"),
			CreateTime: tasks[i].CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if tasks[i].Task != nil {
			download.TaskError = tasks[i].Task.Error
			download.TaskStatus = tasks[i].Task.Status
		}

		resp = append(resp, download)
	}

	return Response{Data: resp}
}

// BuildDownloadingResponse 构建正在下载的列表响应
func BuildDownloadingResponse(tasks []model.Download) Response {
	resp := make([]DownloadListResponse, 0, len(tasks))
	interval := model.GetIntSetting("aria2_interval", 10)

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
			UpdateTime:     tasks[i].UpdatedAt.Unix(),
			UpdateInterval: interval,
			Name:           fileName,
			Status:         tasks[i].Status,
			Dst:            tasks[i].Dst,
			Total:          tasks[i].TotalSize,
			Downloaded:     tasks[i].DownloadedSize,
			Speed:          tasks[i].Speed,
			Info:           tasks[i].StatusInfo,
		})
	}

	return Response{Data: resp}
}
