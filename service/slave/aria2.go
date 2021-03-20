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
