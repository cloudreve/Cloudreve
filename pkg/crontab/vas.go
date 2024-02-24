package crontab

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

func notifyExpiredVAS() {
	checkStoragePack()
	checkUserGroup()
	util.Log().Info("Crontab job \"cron_notify_user\" complete.")
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
			util.Log().Warning("Crontab job failed to get user info of [UID=%d]: %s", pack.UserID, err)
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
		util.Log().Warning("Failed to send notification email: %s", err)
	}
}
