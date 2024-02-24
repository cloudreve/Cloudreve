package admin

import (
	"encoding/gob"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/email"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/thumb"
	"github.com/cloudreve/Cloudreve/v3/pkg/vol"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
}

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
	if err := email.Send(service.Email, "Cloudreve Email delivery test", "This is a test Email, to test Cloudreve Email delivery settings"); err != nil {
		return serializer.Err(serializer.CodeFailedSendEmail, err.Error(), nil)
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
	tx := model.DB.Begin()

	for _, setting := range service.Options {

		if err := tx.Model(&model.Setting{}).Where("name = ?", setting.Key).Update("value", setting.Value).Error; err != nil {
			cache.Deletes(cacheClean, "setting_")
			tx.Rollback()
			return serializer.Err(serializer.CodeUpdateSetting, "Setting "+setting.Key+" failed to update", err)
		}

		cacheClean = append(cacheClean, setting.Key)
	}

	if err := tx.Commit().Error; err != nil {
		return serializer.DBErr("Failed to update setting", err)
	}

	cache.Deletes(cacheClean, "setting_")

	return serializer.Response{}
}

// Summary 获取站点统计概况
func (service *NoParamService) Summary() serializer.Response {
	// 获取版本信息
	versions := map[string]string{
		"backend": conf.BackendVersion,
		"db":      conf.RequiredDBVersion,
		"commit":  conf.LastCommit,
		"is_plus": conf.IsPlus,
	}

	if res, ok := cache.Get("admin_summary"); ok {
		resMap := res.(map[string]interface{})
		resMap["version"] = versions
		resMap["siteURL"] = model.GetSettingByName("siteURL")
		return serializer.Response{Data: resMap}
	}

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

	resp := map[string]interface{}{
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
	}

	cache.Set("admin_summary", resp, 86400)
	return serializer.Response{
		Data: resp,
	}
}

// ThumbGeneratorTestService 缩略图生成测试服务
type ThumbGeneratorTestService struct {
	Name       string `json:"name" binding:"required"`
	Executable string `json:"executable" binding:"required"`
}

// Test 通过获取生成器版本来测试
func (s *ThumbGeneratorTestService) Test(c *gin.Context) serializer.Response {
	version, err := thumb.TestGenerator(c, s.Name, s.Executable)
	if err != nil {
		return serializer.Err(serializer.CodeParamErr, err.Error(), err)
	}

	return serializer.Response{
		Data: version,
	}
}

// VOL 授权管理服务
type VolService struct {
}

// Sync 同步 VOL 授权
func (s *VolService) Sync() serializer.Response {
	volClient := vol.New(vol.ClientSecret)
	content, signature, err := volClient.Sync()
	if err != nil {
		return serializer.Err(serializer.CodeInternalSetting, err.Error(), err)
	}

	subService := &BatchSettingChangeService{
		Options: []SettingChangeService{
			{
				Key:   "vol_content",
				Value: content,
			},
			{
				Key:   "vol_signature",
				Value: signature,
			},
		},
	}

	res := subService.Change()
	if res.Code != 0 {
		return res
	}

	return serializer.Response{Data: content}
}
