package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/resource"
	"net/http"
)

func RegisterByGin(ginApp *gin.Engine) {
	// 加载HTML模板
	ginApp.SetHTMLTemplate(resource.GetTemplates())
	// 基础中间件
	ginApp.Use(middleware.SiteMaintenance)
	ginApp.Use(middleware.SiteInfo)
	ginApp.Use(middleware.GinCors)

	ginApp.GET("/reload", func(c *gin.Context) {
		if setting.IsProduction() {
			c.String(http.StatusNotFound, "404")
			return
		}
		Reload()
		ginApp.SetHTMLTemplate(resource.GetTemplates())
		c.String(200, "模板已刷新")
	})

	// 访问日志中间件
	ginApp.Use(middleware.AccessLog)

	// 接口
	auth(ginApp)
	viewRoute(ginApp)
	viewRouteV2(ginApp)
	forumRoute(ginApp)
	fileServer(ginApp)
	view(ginApp)

	// 前端资源
	// 因为存在 /*filepath 所以要放在最后面
	frontend(ginApp)

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
