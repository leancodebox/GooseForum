package routes

import (
	"github.com/leancodebox/GooseForum/resource"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
)

func RegisterByGin(ginApp *gin.Engine) {
	// 加载HTML模板
	ginApp.SetHTMLTemplate(resource.GetTemplates())
	// 基础中间件
	ginApp.Use(middleware.SiteMaintenance)
	ginApp.Use(middleware.SiteInfo)
	ginApp.Use(middleware.GinCors)

	// 前端资源
	frontend(ginApp)

	// 访问日志中间件
	ginApp.Use(middleware.GinLogger)

	// 接口
	auth(ginApp)
	bbs(ginApp)
	fileServer(ginApp)

	ginApp.NoRoute(controllers.NotFound)

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

func SetupRegisterByGin(ginApp *gin.Engine) {
	// 基础中间件
	ginApp.Use(middleware.SiteMaintenance)
	ginApp.Use(middleware.SiteInfo)
	ginApp.Use(middleware.GinCors)

	// 前端资源
	frontend(ginApp)

	// 访问日志中间件
	ginApp.Use(middleware.GinLogger)

	// 接口
	setup(ginApp)

	ginApp.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/actor/setup.html")
		return
	})
}
