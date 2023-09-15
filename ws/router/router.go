package router

import (
	"github.com/gin-gonic/gin"
	"ws/api"
	"ws/api/middleware"
	"ws/service"
)

func InitRouter() {
	r := gin.Default()

	//r.Use(middleware.CORS())
	r.Use(middleware.JWTAuthMiddleware())
	rGroup := r.Group("/")
	{
		rGroup.POST("login", api.Login)            //登录
		rGroup.POST("register", api.Register)      //注册
		rGroup.POST("search", api.SearchForFriend) //搜寻所有用户
		rGroup.GET("ws", api.StartSocket)
		rGroup.POST("addfriend", api.AddFriend)
		service.PicURL(r)
		service.VideoURL(r)
		rGroup.POST("deletefriend", api.DeleteFriend)
	}

	r.Run(":8080")
}
