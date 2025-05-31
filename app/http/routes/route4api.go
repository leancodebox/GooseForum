package routes

import (
	"bytes"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/assert"
	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/resource"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

var ht *template.Template
var htOnce sync.Once

func getHt() {
	htOnce.Do(func() {
		nuxtFs, _ := fs.Sub(assert.GetActorFs(), "frontend/nuxt")
		// 创建基础模板
		tmpl := template.New("base")
		// 遍历文件系统
		err := fs.WalkDir(nuxtFs, ".", func(path string, d fs.DirEntry, err error) error {
			// 跳过目录和非 .html 文件
			if d.IsDir() || filepath.Ext(path) != ".html" {
				return nil
			}
			// 读取文件内容
			content, err := fs.ReadFile(nuxtFs, path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}
			// 生成唯一的模板名
			templateName := generateTemplateName(path)
			// 解析模板
			if _, err = tmpl.New(templateName).Parse(string(content)); err != nil {
				return fmt.Errorf("error parsing template %s: %v", path, err)
			}
			fmt.Printf("Loaded template: %s -> %s\n", path, templateName)
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking the filesystem: %v\n", err)
		}
		ht = tmpl
	})
}

// generateTemplateName 将文件路径转换为唯一的模板名
func generateTemplateName(path string) string {
	// 移除前导的 ./ 或 .
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, ".")
	// 替换路径分隔符为下划线
	name := strings.ReplaceAll(path, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// 移除 .html 扩展名
	name = strings.TrimSuffix(name, ".html")
	return name
}

// filteredFileSystem 包装原始文件系统并过滤掉 HTML 文件
type filteredFileSystem struct {
	fs fs.FS
}

func (f *filteredFileSystem) Open(name string) (fs.File, error) {
	// 检查是否是 HTML 文件
	if strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm") {
		return nil, fs.ErrNotExist // 返回"文件不存在"错误
	}
	return f.fs.Open(name)
}

func frontend(ginApp *gin.Engine) {
	actGroup := ginApp.Group("/")
	appFs, _ := fs.Sub(assert.GetActorFs(), "frontend/dist")
	nuxtFs, _ := fs.Sub(assert.GetActorFs(), "frontend/nuxt")
	staticFS, _ := resource.GetStaticFS()
	//filteredFs := &filteredFileSystem{fs: nuxtFs}
	actGroup.Use(middleware.CacheMiddleware).
		Use(gzip.Gzip(gzip.DefaultCompression)).
		Use(middleware.BrowserCache).
		StaticFS("static", http.FS(staticFS)).
		StaticFS("app", http.FS(appFs)).
		StaticFS("nuxt", http.FS(nuxtFs))
}

func viewRouteV2(ginApp *gin.Engine) {
	getHt()
	// no use <NuxtLink :to="`/detail?id=${comment.postId}`" class="link link-primary">{{ comment.postTitle }}</NuxtLink>
	// use <a href>
	ginApp.GET("/new-post", func(c *gin.Context) {
		// 3. 执行模板渲染到缓冲区
		var buf bytes.Buffer
		if err := ht.ExecuteTemplate(&buf, "list_index", map[string]any{
			"Title": "newgooseforum",
		}); err != nil {
			slog.Error(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 4. 直接返回渲染结果
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})
	ginApp.GET("/new-post/:id", func(c *gin.Context) {
		// 3. 执行模板渲染到缓冲区
		var buf bytes.Buffer
		if err := ht.ExecuteTemplate(&buf, "detail_index", map[string]any{
			"title": "newgooseforum",
		}); err != nil {
			slog.Error(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 4. 直接返回渲染结果
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})

	// SEO 相关路由
	ginApp.GET("/robots.txt", controllers.RenderRobotsTxt)
	ginApp.GET("/sitemap.xml", controllers.RenderSitemapXml)
	ginApp.GET("/rss.xml", controllers.RenderRssV2)
}

// 认证相关服务
func auth(ginApp *gin.Engine) {
	// 非登陆下的用户操作
	ginApp.Group("api").
		GET("get-captcha", ginUpNP(controllers.GetCaptcha)).
		POST("get-user-info-show", ginUpP(controllers.GetUserInfo))
	// 登陆状态下的用户操作
	ginApp.Group("api").Use(middleware.JWTAuth4Gin).
		GET("get-user-info", UpButterReq(controllers.UserInfo)).
		POST("set-user-info", UpButterReq(controllers.EditUserInfo)).
		POST("invitation", UpButterReq(controllers.Invitation)).
		POST("upload-avatar", controllers.UploadAvatar).
		POST("change-password", UpButterReq(controllers.ChangePassword))
	// 添加激活路由
	ginApp.GET("api/activate", controllers.ActivateAccount)
}

func viewRoute(ginApp *gin.Engine) {
	view := ginApp.Group("")
	view.Use(middleware.JWTAuth)
	view.GET("", controllers.RenderArticlesPage)
	view.GET("/post", controllers.RenderArticlesPage)
	view.GET("/post/:id", controllers.RenderArticleDetail)
	view.GET("/login", controllers.LoginPage)
	view.POST("/login", controllers.LoginHandler)
	view.POST("/register", controllers.RegisterHandle)
	view.POST("/logout", controllers.Logout)
	view.GET("/user-profile/:id", controllers.UserProfile)
	view.GET("/sponsors", controllers.Sponsors)
	view.GET("/links", controllers.Links)
	view.GET("/link-contact", controllers.Contact)

	forumApi := ginApp.Group("api/bbs")
	forumApi.POST("apply-link-add", UpButterReq(controllers.ApplyAddLink))
}

func forumRoute(ginApp *gin.Engine) {
	forumApi := ginApp.Group("api/bbs")
	// 站点统计
	forumApi.GET("get-site-statistics", ginUpNP(controllers.GetSiteStatistics))
	// 分类列表
	forumApi.GET("get-articles-enum", ginUpNP(controllers.GetArticlesEnum))
	forumApi.GET("get-articles-category", ginUpNP(controllers.GetArticlesCategory))

	loginApi := forumApi.Use(middleware.JWTAuth4Gin)
	// 通知相关接口
	loginApi.POST("notification/list", UpButterReq(controllers.GetNotificationList))
	loginApi.GET("notification/unread-count", UpButterReq(controllers.GetUnreadCount))
	loginApi.POST("notification/mark-read", UpButterReq(controllers.MarkAsRead))
	loginApi.POST("notification/mark-all-read", UpButterReq(controllers.MarkAllAsRead))
	loginApi.POST("notification/delete", UpButterReq(controllers.DeleteNotification))
	loginApi.GET("notification/types", UpButterReq(controllers.GetNotificationTypes))

	// 编辑文章时原始文章内容
	loginApi.POST("get-articles-origin", middleware.CheckLogin, UpButterReq(controllers.WriteArticlesOrigin))
	// 发布文章
	loginApi.POST("write-articles", UpButterReq(controllers.WriteArticles))
	// 回复文章
	loginApi.POST("articles-reply", UpButterReq(controllers.ArticleReply))
	// 回复评论
	loginApi.POST("articles-reply-delete", UpButterReq(controllers.DeleteReply))
	// 用户文章列表
	loginApi.POST("get-user-articles", UpButterReq(controllers.GetUserArticles))
	// 文章点赞
	loginApi.POST("like-articles", UpButterReq(controllers.LikeArticle))

	adminApi := ginApp.Group("api/admin").Use(middleware.JWTAuth4Gin)
	adminApi.POST("user-list", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.UserList))
	adminApi.POST("user-edit", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.EditUser))
	adminApi.POST("get-all-role-item", middleware.CheckPermission(permission.UserManager), UpButterReq(controllers.GetAllRoleItem))
	adminApi.POST("articles-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.ArticlesList))
	adminApi.POST("article-edit", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.EditArticle))
	adminApi.POST("get-permission-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.GetPermissionList))
	adminApi.POST("role-list", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleList))
	adminApi.POST("role-save", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleSave))
	adminApi.POST("role-delete", middleware.CheckPermission(permission.RoleManager), UpButterReq(controllers.RoleDel))
	adminApi.POST("opt-record-page", middleware.CheckPermission(permission.Admin), UpButterReq(controllers.OptRecordPage))
	adminApi.POST("category-list", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.GetCategoryList))
	adminApi.POST("category-save", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.SaveCategory))
	adminApi.POST("category-delete", middleware.CheckPermission(permission.ArticlesManager), UpButterReq(controllers.DeleteCategory))

	adminApi.POST("apply-sheet-list", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.ApplySheet))
	adminApi.GET("friend-links", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.GetFriendLinks))
	adminApi.POST("save-friend-links", middleware.CheckPermission(permission.SiteManager), UpButterReq(controllers.SaveFriendLinks))

}

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	// 文件上传接口
	r.POST("/img-upload", middleware.JWTAuth4Gin, controllers.SaveFileByGinContext)
	// 文件获取接口 - 通过路径
	r.GET("/img/*filename", middleware.BrowserCache, controllers.GetFileByFileName)
}
