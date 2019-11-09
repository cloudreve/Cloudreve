package routers

import (
	"cloudreve/routers/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 顶层路由分组
	v3 := r.Group("/Api/V3")
	{
		// 测试用路由
		v3.GET("Ping", controllers.Ping)
		// 用户登录
		v3.POST("User/Login", controllers.UserLogin)

	}
	return r
}
