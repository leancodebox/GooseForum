package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
)

func api(ginApp *gin.Engine) {
	apiGroup := ginApp.Group("api")
	apiGroup.GET("about", ginUpNP(controllers.About))

	// 权限组合判断/上面的代码为 fullpath的获取，下面的代码为所有routes的获取，
	// 两者配合就可以实现基于接口权限设置的底层资源配置，和拦截
	//ginApp.GET("fullpath/:name/:id", func(c *gin.Context) {
	//	c.JSON(200, map[string]any{
	//		"f": c.FullPath(),
	//	})
	//})
	//for _, item := range ginApp.Routes() {
	//	fmt.Println(item.Path)
	//}
}
