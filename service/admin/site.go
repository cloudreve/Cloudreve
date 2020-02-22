package admin

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"time"
)

// NoParamService 无需参数的服务
type NoParamService struct {
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
