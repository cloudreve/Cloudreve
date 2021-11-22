package scripts

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/fatih/color"
)

type ResetAdminPassword int

// Run 运行脚本从社区版升级至 Pro 版
func (script ResetAdminPassword) Run(ctx context.Context) {
	// 查找用户
	user, err := model.GetUserByID(1)
	if err != nil {
		util.Log().Panic("初始管理员用户不存在, %s", err)
	}

	// 生成密码
	password := util.RandStringRunes(8)

	// 更改为新密码
	user.SetPassword(password)
	if err := user.Update(map[string]interface{}{"password": user.Password}); err != nil {
		util.Log().Panic("密码更改失败, %s", err)
	}

	c := color.New(color.FgWhite).Add(color.BgBlack).Add(color.Bold)
	util.Log().Info("初始管理员密码已更改为：" + c.Sprint(password))
}
