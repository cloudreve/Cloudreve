package share

import (
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
)

// ShareGetService 获取分享服务
type ShareGetService struct {
	Password string `form:"password" binding:"max=255"`
}

// Get 获取分享内容
func (service *ShareGetService) Get(c *gin.Context) serializer.Response {
	share := model.GetShareByHashID(c.Param("id"))
	if share == nil || !share.IsAvailable() {
		return serializer.Err(serializer.CodeNotFound, "分享不存在或已被取消", nil)
	}

	// 是否已解锁
	unlocked := true
	if share.Password != "" {
		sessionKey := fmt.Sprintf("share_unlock_%d", share.ID)
		unlocked = util.GetSession(c, sessionKey) != nil
		if !unlocked && service.Password != "" {
			// 如果未解锁，且指定了密码，则尝试解锁
			if service.Password == share.Password {
				unlocked = true
				util.SetSession(c, map[string]interface{}{sessionKey: true})
			}
		}
	}

	return serializer.Response{
		Code: 0,
		Data: serializer.BuildShareResponse(share, unlocked),
	}
}
