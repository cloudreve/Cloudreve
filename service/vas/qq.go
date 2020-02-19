package vas

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/qq"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/gin-gonic/gin"
)

// QQCallbackService QQ互联回调处理服务
type QQCallbackService struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// Callback 处理QQ互联回调
func (service *QQCallbackService) Callback(c *gin.Context, user *model.User) serializer.Response {
	state := util.GetSession(c, "qq_login_secret")
	if stateStr, ok := state.(string); !ok || stateStr != service.State {
		return serializer.Err(serializer.CodeSignExpired, "请求过期，请重试", nil)
	}

	credential, err := qq.Callback(service.Code)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "无法获取登录状态", err)
	}

	// 如果已登录，则绑定已有用户
	if user != nil {
		if user.OpenID != "" {
			return serializer.Err(serializer.CodeCallbackError, "您已绑定了QQ账号，请先解除绑定", nil)
		}
		if err := user.Update(map[string]interface{}{"open_id": credential.OpenID}); err != nil {
			return serializer.DBErr("绑定失败", err)
		}
		return serializer.Response{
			Data: "/setting",
		}
	}

	return serializer.Response{}
}
