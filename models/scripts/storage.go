package scripts

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

type UserStorageCalibration int

func init() {
	register("CalibrateUserStorage", UserStorageCalibration(0))
}

type storageResult struct {
	Total uint64
}

// Run 运行脚本校准所有用户容量
func (script UserStorageCalibration) Run(ctx context.Context) {
	// 列出所有用户
	var res []model.User
	model.DB.Model(&model.User{}).Find(&res)

	// 逐个检查容量
	for _, user := range res {
		// 计算正确的容量
		var total storageResult
		model.DB.Model(&model.File{}).Where("user_id = ?", user.ID).Select("sum(size) as total").Scan(&total)
		// 更新用户的容量
		if user.Storage != total.Total {
			util.Log().Info("将用户 [%s] 的容量由 %d 校准为 %d", user.Email,
				user.Storage, total.Total)
			model.DB.Model(&user).Update("storage", total.Total)
		}
	}
}
