package scripts

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
)

type UpgradeToPro int

// Run 运行脚本从社区版升级至 Pro 版
func (script UpgradeToPro) Run(ctx context.Context) {
	// folder.PolicyID 字段设为 0
	model.DB.Model(model.Folder{}).UpdateColumn("policy_id", 0)
	// shares.Score 字段设为0
	model.DB.Model(model.Share{}).UpdateColumn("score", 0)
	// user 表相关初始字段
	model.DB.Model(model.User{}).Updates(map[string]interface{}{
		"score":             0,
		"previous_group_id": 0,
		"open_id":           "",
	})
}
