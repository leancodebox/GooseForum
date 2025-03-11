package controllers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	jwt "github.com/leancodebox/GooseForum/app/bundles/goose/jwtopt"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"github.com/spf13/cast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RegisterHandle 注册
func RegisterHandle(c *gin.Context) {
	var r RegReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(200, component.FailData("验证失败"))
		return
	}
	// 首先验证验证码
	if !VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
		c.JSON(200, component.FailData("验证码错误或已过期"))
		return
	}

	// 检查用户名是否已存在
	if users.ExistUsername(r.Username) {
		c.JSON(200, component.FailData("用户名已存在"))
		return
	}

	// 检查邮箱是否已存在
	if users.ExistEmail(r.Email) {
		c.JSON(200, component.FailData("邮箱已被使用"))
		return
	}

	userEntity := users.MakeUser(r.Username, r.Password, r.Email)
	err := users.Create(userEntity)
	if err != nil {
		c.JSON(200, component.FailData("注册失败"))
	}

	if err = SendAEmail4User(userEntity); err != nil {
		slog.Error("添加邮件任务到队列失败", "error", err)
	}

	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)

	// 生成 token
	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {

		c.JSON(200, component.FailData("注册异常，尝试登陆"))
	}

	// 设置Cookie
	c.SetCookie(
		"access_token",
		token,
		86400, // 24小时
		"/",
		"",    // 域名，为空表示当前域名
		false, // 仅HTTPS
		true,  // HttpOnly
	)
	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}

type LoginHandlerReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// LoginHandler 处理登录请求
func LoginHandler(c *gin.Context) {
	var req LoginHandlerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, component.FailData("验证失败"))
		return
	}
	username := req.Username
	password := req.Password
	captchaId := req.CaptchaId
	captchaCode := req.CaptchaCode

	if !VerifyCaptcha(captchaId, captchaCode) {
		c.JSON(200, component.FailData("验证失败"))
		return
	}
	userEntity, err := users.Verify(username, password)
	if err != nil {
		slog.Info(cast.ToString(err))
		c.JSON(200, component.FailData(err))
		return
	}
	token, err := jwt.CreateNewToken(userEntity.Id, expireTime)
	if err != nil {
		slog.Info(cast.ToString(err))
		c.JSON(200, component.FailData(err))
		return
	}
	// 设置Cookie
	c.SetCookie(
		"access_token",
		token,
		86400, // 24小时
		"/",
		"",    // 域名，为空表示当前域名
		false, // 仅HTTPS
		true,  // HttpOnly
	)
	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}
func GetLoginUser(c *gin.Context) UserInfoShow {
	userId := c.GetUint64("userId")
	if userId == 0 {
		return UserInfoShow{}
	}
	user, _ := users.Get(userId)
	if user.Id == 0 {
		return UserInfoShow{}
	}
	//userPoint := userPoints.Get(user.Id)
	// 如果有头像，添加域名前缀
	avatarUrl := "/file/img/default.png"
	if user.AvatarUrl != "" {
		avatarUrl = strings.ReplaceAll(component.FilePath(user.AvatarUrl), "\\", "/")
	}

	return UserInfoShow{
		UserId:    userId,
		Username:  user.Username,
		Prestige:  user.Prestige,
		AvatarUrl: avatarUrl,
		//UserPoint: userPoint.CurrentPoints,
	}
}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// 添加新的服务端渲染的控制器方法
func markdownToHTML(markdown string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		slog.Error("转化失败", "err", err)
	}
	return template.HTML(buf.String())
}
func RenderIndex(c *gin.Context) {
	last, _ := articles.GetLatestArticles(10)
	templateData := gin.H{
		"FeaturedArticles": articlesSmallEntity2Dto(getRecommendedArticles()),
		"LatestArticles":   articlesSmallEntity2Dto(last),
		"Stats":            GetSiteStatisticsData(),
		"User":             GetLoginUser(c),
	}
	c.HTML(http.StatusOK, "home.gohtml", templateData)
}

// RenderArticlesPage 渲染文章列表页面
func RenderArticlesPage(c *gin.Context) {
	param := GetArticlesPageRequest{
		Page:     cast.ToInt(c.DefaultQuery("page", "1")),
		PageSize: cast.ToInt(c.DefaultQuery("pageSize", "20")),
		Search:   c.Query("search"),
	}

	// 复用现有的数据获取逻辑
	response := GetArticlesPage(param)
	if response.Code != 200 {
		errorPage(c, "获取文章列表失败", "获取文章列表失败")
		return
	}
	result := response.Data.Result.(component.Page[ArticlesSimpleDto])
	// 计算总页数
	totalPages := (cast.ToInt(result.Total) + param.PageSize - 1) / param.PageSize
	articleCategoryList := array.Map(articleCategory.All(), func(t *articleCategory.Entity) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{
			Name:  t.Category,
			Value: t.Id,
		}
	})
	// 构建模板数据
	templateData := gin.H{
		"title":               "文章列表",
		"description":         "GooseForum的文章列表页面",
		"year":                time.Now().Year(),
		"Data":                result.List,
		"Page":                result.Page,
		"PageSize":            param.PageSize,
		"Total":               result.Total,
		"TotalPages":          totalPages,
		"PrevPage":            max(result.Page-1, 1),
		"NextPage":            min(max(result.Page, 1)+1, totalPages),
		"User":                GetLoginUser(c),
		"articleCategoryList": articleCategoryList,
		"recommendedArticles": getRecommendedArticles(),
	}
	c.HTML(http.StatusOK, "list.gohtml", templateData)
}

// RenderArticleDetail 渲染文章详情页面
func RenderArticleDetail(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	if id == 0 {
		errorPage(c, "页面不存在", "页面不存在")
		return
	}

	req := GetArticlesDetailRequest{
		Id:           id,
		MaxCommentId: 0,
		PageSize:     50,
	}

	// 复用现有的数据获取逻辑
	response := GetArticlesDetail(req)
	result := response.Data.Result.(map[string]any)

	if _, ok := result["id"]; !ok {
		errorPage(c, "页面不存在", "文章不存在")
		return
	}
	// 构建模板数据
	templateData := gin.H{
		"articleId":      id,
		"authorId":       result["userId"],
		"title":          cast.ToString(result["articleTitle"]),
		"description":    TakeUpTo64Chars(cast.ToString(result["articleContent"])),
		"year":           time.Now().Year(),
		"articleTitle":   cast.ToString(result["articleTitle"]),
		"articleContent": markdownToHTML(cast.ToString(result["articleContent"])),
		"username":       cast.ToString(result["username"]),
		"commentList":    result["commentList"],
		"avatarUrl":      result["avatarUrl"],
		"User":           GetLoginUser(c),
	}

	c.HTML(http.StatusOK, "detail.gohtml", templateData)
}

func errorPage(c *gin.Context, title, message string) {
	c.HTML(http.StatusNotFound, "error.gohtml", gin.H{
		"title":   title,
		"message": message,
		"year":    time.Now().Year(),
		"User":    GetLoginUser(c),
	})
}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.gohtml", gin.H{"title": "登录 - GooseForum", "User": GetLoginUser(c)})
}
func Notifications(c *gin.Context) {
	c.HTML(http.StatusNotFound, "notifications.gohtml", gin.H{"title": "消息通知 - GooseForum", "User": GetLoginUser(c)})
}

func UserProfile(c *gin.Context) {
	c.HTML(http.StatusNotFound, "user_profile.gohtml", gin.H{"title": "用户主页 - GooseForum", "User": GetLoginUser(c)})
}

func Sponsors(c *gin.Context) {
	c.HTML(http.StatusNotFound, "sponsors.gohtml", gin.H{"title": "赞助商 - GooseForum", "User": GetLoginUser(c)})
}
