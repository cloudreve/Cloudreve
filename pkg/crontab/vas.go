package crontab

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/email"
	"github.com/HFO4/cloudreve/pkg/util"
)

func notifyExpiredVAS() {
	checkStoragePack()
	checkUserGroup()
	util.Log().Info("定时任务 [cron_notify_user] 执行完毕")
}

// banOverusedUser 封禁超出宽容期的用户
func banOverusedUser() {
	users := model.GetTolerantExpiredUser()
	for _, user := range users {

		// 清除最后通知日期标记
		user.ClearNotified()

		// 检查容量是否超额
		if user.Storage > user.Group.MaxStorage+user.GetAvailablePackSize() {
			// 封禁用户
			user.SetStatus(model.OveruseBaned)
		}
	}
}

// checkUserGroup 检查已过期用户组
func checkUserGroup() {
	users := model.GetGroupExpiredUsers()
	for _, user := range users {

		// 将用户回退到初始用户组
		user.GroupFallback()

		// 重新加载用户
		user, _ = model.GetUserByID(user.ID)

		// 检查容量是否超额
		if user.Storage > user.Group.MaxStorage+user.GetAvailablePackSize() {
			// 如果超额，则通知用户
			sendNotification(&user, "用户组过期")
			// 更新最后通知日期
			user.Notified()
		}
	}
}

// checkStoragePack 检查已过期的容量包
func checkStoragePack() {
	packs := model.GetExpiredStoragePack()
	for _, pack := range packs {
		// 删除过期的容量包
		pack.Delete()

		//找到所属用户
		user, err := model.GetUserByID(pack.UserID)
		if err != nil {
			util.Log().Warning("[定时任务] 无法获取用户 [UID=%d] 信息, %s", pack.UserID, err)
			continue
		}

		// 检查容量是否超额
		if user.Storage > user.Group.MaxStorage+user.GetAvailablePackSize() {
			// 如果超额，则通知用户
			sendNotification(&user, "容量包过期")

			// 更新最后通知日期
			user.Notified()

		}
	}
}

func sendNotification(user *model.User, reason string) {
	title, body := email.NewOveruseNotification(user.Nick, reason)
	if err := email.Send(user.Email, title, body); err != nil {
		util.Log().Warning("无法发送通知邮件, %s", err)
	}
}
