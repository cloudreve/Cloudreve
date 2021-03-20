package slave

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

type Aria2AddService struct {
	TaskId  uint                   `json:"task_id"`
	Options map[string]interface{} `json:"options"`
}

type Aria2CancelService struct {
	TaskId uint `uri:"taskId"`
}

func (service *Aria2AddService) Add() serializer.Response {
	task, err := model.GetDownloadById(service.TaskId)
	if err != nil {
		util.Log().Warning("无法获取记录, %s", err)
		return serializer.Err(serializer.CodeNotSet, "任务创建失败, 无法获取记录", err)
	}
	aria2.Lock.RLock()
	if err := aria2.Instance.CreateTask(task, service.Options); err != nil {
		aria2.Lock.RUnlock()
		return serializer.Err(serializer.CodeNotSet, "任务创建失败", err)
	}
	aria2.Lock.RUnlock()
	return serializer.Response{}
}

func (service *Aria2CancelService) Cancel() serializer.Response {
	task, err := model.GetDownloadById(service.TaskId)
	if err != nil {
		util.Log().Warning("无法获取记录, %s", err)
		return serializer.Err(serializer.CodeNotSet, "任务创建失败, 无法获取记录", err)
	}

	// 取消任务
	aria2.Lock.RLock()
	defer aria2.Lock.RUnlock()
	if err := aria2.Instance.Cancel(task); err != nil {
		util.Log().Debug("删除远程下载任务出错, %s", err)
		return serializer.Err(serializer.CodeNotSet, "操作失败", err)
	}

	return serializer.Response{}
}
