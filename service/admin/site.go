package admin

import (
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// NoParamService 无需参数的服务
type NoParamService struct {
}

// BatchSettingChangeService 设定批量更改服务
type BatchSettingChangeService struct {
	Options []SettingChangeService `json:"options"`
}

// SettingChangeService  设定更改服务
type SettingChangeService struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

// BatchSettingGet 设定批量获取服务
type BatchSettingGet struct {
	Keys []string `json:"keys"`
}

// MailTestService 邮件测试服务
type MailTestService struct {
	Email string `json:"to" binding:"email"`
}

// Send 发送测试邮件
func (service *MailTestService) Send() serializer.Response {
	if err := email.Send(service.Email, "Cloudreve发信测试", "这是一封测试邮件，用于测试 Cloudreve 发信设置。"); err != nil {
		return serializer.Err(serializer.CodeInternalSetting, "发信失败, "+err.Error(), nil)
	}
	return serializer.Response{}
}

// Get 获取设定值
func (service *BatchSettingGet) Get() serializer.Response {
	options := model.GetSettingByNames(service.Keys...)
	return serializer.Response{Data: options}
}

// Change 批量更改站点设定
func (service *BatchSettingChangeService) Change() serializer.Response {
	cacheClean := make([]string, 0, len(service.Options))

	for _, setting := range service.Options {

		if err := model.DB.Model(&model.Setting{}).Where("name = ?", setting.Key).Update("value", setting.Value).Error; err != nil {
			cache.Deletes(cacheClean, "setting_")
			return serializer.DBErr("设置 "+setting.Key+" 更新失败", err)
		}

		cacheClean = append(cacheClean, setting.Key)
	}

	cache.Deletes(cacheClean, "setting_")

	return serializer.Response{}
}

// Summary 获取站点统计概况
func (service *NoParamService) Summary() serializer.Response {
	// 统计每日概况
	total := 12
	files := make([]int, total)
	users := make([]int, total)
	shares := make([]int, total)
	date := make([]string, total)

	toRound := time.Now()
	timeBase := time.Date(toRound.Year(), toRound.Month(), toRound.Day()+1, 0, 0, 0, 0, toRound.Location())
	for day := range files {
		start := timeBase.Add(-time.Duration(total-day) * time.Hour * 24)
		end := timeBase.Add(-time.Duration(total-day-1) * time.Hour * 24)
		date[day] = start.Format("1月2日")
		model.DB.Model(&model.User{}).Where("created_at BETWEEN ? AND ?", start, end).Count(&users[day])
		model.DB.Model(&model.File{}).Where("created_at BETWEEN ? AND ?", start, end).Count(&files[day])
		model.DB.Model(&model.Share{}).Where("created_at BETWEEN ? AND ?", start, end).Count(&shares[day])
	}

	// 统计总数
	fileTotal := 0
	userTotal := 0
	publicShareTotal := 0
	secretShareTotal := 0
	model.DB.Model(&model.User{}).Count(&userTotal)
	model.DB.Model(&model.File{}).Count(&fileTotal)
	model.DB.Model(&model.Share{}).Where("password = ?", "").Count(&publicShareTotal)
	model.DB.Model(&model.Share{}).Where("password <> ?", "").Count(&secretShareTotal)

	// 获取版本信息
	versions := map[string]string{
		"backend": conf.BackendVersion,
		"db":      conf.RequiredDBVersion,
		"commit":  conf.LastCommit,
		"is_pro":  conf.IsPro,
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"date":             date,
			"files":            files,
			"users":            users,
			"shares":           shares,
			"version":          versions,
			"siteURL":          model.GetSettingByName("siteURL"),
			"fileTotal":        fileTotal,
			"userTotal":        userTotal,
			"publicShareTotal": publicShareTotal,
			"secretShareTotal": secretShareTotal,
		},
	}
}
