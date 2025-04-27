package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/jsonopt"
	jwt "github.com/leancodebox/GooseForum/app/bundles/goose/jwtopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/applySheet"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/spf13/cast"
	"html/template"
	"regexp"
	"strings"

	"log/slog"
	"net/http"
	"time"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{6,32}$`)
)

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func Logout(c *gin.Context) {
	jwt.TokenClean(c)
	c.JSON(http.StatusOK, component.SuccessData(
		"再见",
	))
}

// RegisterHandle 注册
func RegisterHandle(c *gin.Context) {
	var r RegReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(200, component.FailData(err))
		return
	}
	if err := validate.Valid(r); err != nil {
		c.JSON(200, component.FailData(validate.FormatError(err)))
		return
	}
	if !ValidateUsername(r.Username) {
		c.JSON(200, component.FailData("用户名仅允许字母、数字、下划线、连字符，长度6-32"))
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
	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {

		c.JSON(200, component.FailData("注册异常，尝试登陆"))
	}
	// 设置Cookie
	jwt.TokenSetting(c, token)

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
	token, err := jwt.CreateNewTokenDefault(userEntity.Id)
	if err != nil {
		slog.Info(cast.ToString(err))
		c.JSON(200, component.FailData(err))
		return
	}
	jwt.TokenSetting(c, token)
	c.JSON(http.StatusOK, component.SuccessData(
		"登录成功",
	))
}
func GetLoginUser(c *gin.Context) UserInfoShow {
	userId := c.GetUint64("userId")
	return GetUserShowByUserId(userId)
}

func GetUserShowByUserId(userId uint64) UserInfoShow {
	if userId == 0 {
		return UserInfoShow{}
	}
	user, _ := users.Get(userId)
	if user.Id == 0 {
		return UserInfoShow{}
	}
	//userPoint := userPoints.Get(user.Id)
	// 如果有头像，添加域名前缀
	avatarUrl := user.GetWebAvatarUrl()

	return UserInfoShow{
		UserId:     userId,
		Username:   user.Username,
		Prestige:   user.Prestige,
		AvatarUrl:  avatarUrl,
		CreateTime: user.CreatedAt,
		//UserPoint: userPoint.CurrentPoints,
	}
}

func RenderIndex(c *gin.Context) {
	last, _ := articles.GetLatestArticles(5)
	templateData := gin.H{
		"FeaturedArticles": articlesSmallEntity2Dto(getRecommendedArticles()),
		"LatestArticles":   articlesSmallEntity2Dto(last),
		"Stats":            GetSiteStatisticsData(),
		"User":             GetLoginUser(c),
		"title":            "GooseForum",
		"description":      "GooseForum的首页",
		"canonicalHref":    buildCanonicalHref(c),
	}
	c.HTML(http.StatusOK, "home.gohtml", templateData)
}

// RenderArticlesPage 渲染文章列表页面
func RenderArticlesPage(c *gin.Context) {
	filters := c.DefaultQuery("filters", "")
	categories := array.Filter(array.Map(strings.Split(filters, "-"), func(t string) int {
		return cast.ToInt(t)
	}), func(i int) bool {
		return i > 0
	})
	param := GetArticlesPageRequest{
		Page:       cast.ToInt(c.DefaultQuery("page", "1")),
		PageSize:   cast.ToInt(c.DefaultQuery("pageSize", "20")),
		Search:     c.Query("search"),
		Categories: categories,
	}
	pageData := articles.Page[articles.SmallEntity](
		articles.PageQuery{
			Page:         max(param.Page, 1),
			PageSize:     param.PageSize,
			FilterStatus: true,
			Categories:   param.Categories,
		})
	userIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)

	//获取文章的分类信息
	articleIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.Id
	})
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryIds := array.Map(categoryRs, func(t *articleCategoryRs.Entity) uint64 {
		return t.ArticleCategoryId
	})
	categoryMap := articleCategory.GetMapByIds(categoryIds)
	// 获取文章的分类和标签
	categoriesGroup := array.GroupBy(categoryRs, func(rs *articleCategoryRs.Entity) uint64 {
		return rs.ArticleId
	})

	articleList := array.Map(pageData.Data, func(t articles.SmallEntity) ArticlesSimpleDto {
		categoryNames := array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) string {
			if category, ok := categoryMap[rs.ArticleCategoryId]; ok {
				return category.Category
			}
			return ""
		})
		username := ""
		avatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			avatarUrl = user.GetWebAvatarUrl()
		}
		return ArticlesSimpleDto{
			Id:             t.Id,
			Title:          t.Title,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			Username:       username,
			AuthorId:       t.UserId,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Category:       FirstOr(categoryNames, "未分类"),
			Categories:     categoryNames,
			CategoriesId: array.Map(categoriesGroup[t.Id], func(rs *articleCategoryRs.Entity) uint64 {
				return rs.ArticleCategoryId
			}),
			Type:    t.Type,
			TypeStr: articlesTypeMap[int(t.Type)].Name,
		}
	})
	// 计算总页数
	totalPages := (cast.ToInt(pageData.Total) + param.PageSize - 1) / param.PageSize
	articleCategoryList := array.Map(articleCategory.All(), func(t *articleCategory.Entity) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{
			Name:  t.Category,
			Value: t.Id,
		}
	})
	pagination := []PageButton{}
	start := max(pageData.Page-3, 1)
	for i := 1; i <= 7; i++ {
		pagination = append(pagination, PageButton{Index: i, Page: start})
		start += 1
	}
	// 构建模板数据
	templateData := gin.H{
		"title":               "GooseForum",
		"description":         "知无不言,言无不尽",
		"year":                time.Now().Year(),
		"Data":                articleList,
		"Page":                pageData.Page,
		"PageSize":            param.PageSize,
		"Total":               pageData.Total,
		"TotalPages":          totalPages,
		"PrevPage":            max(pageData.Page-1, 1),
		"NextPage":            min(max(pageData.Page, 1)+1, totalPages),
		"User":                GetLoginUser(c),
		"articleCategoryList": articleCategoryList,
		"recommendedArticles": getRecommendedArticles(),
		"canonicalHref":       buildCanonicalHref(c),
		"Filters":             filters,
		"pagination":          pagination,
	}

	c.HTML(http.StatusOK, "list.gohtml", templateData)
}

type PageButton struct {
	Index int
	Page  int
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
	entity := articles.Get(req.Id)
	if entity.Id == 0 {
		errorPage(c, "文章不存在", "文章不存在")
		return
	}
	replyEntities := reply.GetByMaxIdPage(req.Id, req.MaxCommentId, boundPageSizeWithRange(req.PageSize, 10, 100))
	userIds := array.Map(replyEntities, func(item reply.Entity) uint64 {
		return item.UserId
	})
	userIds = append(userIds, entity.UserId)
	userMap := users.GetMapByIds(userIds)
	author := "陶渊明"
	avatarUrl := urlconfig.GetDefaultAvatar()
	if user, ok := userMap[entity.UserId]; ok {
		author = user.Username
		avatarUrl = user.GetWebAvatarUrl()
	}
	replyList := array.Map(replyEntities, func(item reply.Entity) ReplyDto {
		username := "陶渊明"
		if user, ok := userMap[item.UserId]; ok {
			username = user.Username
		}
		// 获取被回复评论的用户名
		replyToUsername := ""
		if item.ReplyId > 0 {
			if replyTo := reply.Get(item.ReplyId); replyTo.Id > 0 {
				if replyUser, ok := userMap[replyTo.UserId]; ok {
					replyToUsername = replyUser.Username
				}
			}
		}
		return ReplyDto{
			Id:              item.Id,
			ArticleId:       item.ArticleId,
			UserId:          item.UserId,
			Username:        username,
			Content:         item.Content,
			CreateTime:      item.CreatedAt.Format(time.RFC3339),
			ReplyToUsername: replyToUsername,
		}
	})
	articles.IncrementView(entity)
	// 复用现有的数据获取逻辑
	authorId := entity.UserId

	if entity.RenderedVersion < markdown2html.GetVersion() || entity.RenderedHTML == "" {
		mdInfo := markdown2html.MarkdownToHTML(entity.Content)
		entity.RenderedHTML = mdInfo
		entity.RenderedVersion = markdown2html.GetVersion()
		articles.SaveNoUpdate(&entity)
	}

	authorArticles, _ := articles.GetRecommendedArticlesByAuthorId(cast.ToUint64(authorId), 5)
	acMap := articleCategoryMapList([]uint64{id})
	iLike := false
	loginUser := GetLoginUser(c)
	if loginUser.UserId != 0 {
		iLike = articleLike.GetByArticleId(loginUser.UserId, entity.Id).Status == 1
	}
	// 构建模板数据
	templateData := gin.H{
		"articleId":       id,
		"authorId":        authorId,
		"title":           entity.Title + " - GooseForum",
		"description":     TakeUpTo64Chars(entity.Content),
		"year":            time.Now().Year(),
		"articleTitle":    entity.Title,
		"articleContent":  template.HTML(entity.RenderedHTML),
		"LikeCount":       entity.LikeCount,
		"ILike":           iLike,
		"username":        author,
		"commentList":     replyList,
		"avatarUrl":       avatarUrl,
		"User":            GetLoginUser(c),
		"canonicalHref":   buildCanonicalHref(c),
		"authorArticles":  authorArticles,
		"articleCategory": acMap[id],
	}
	c.HTML(http.StatusOK, "detail.gohtml", templateData)
}

func articleCategoryMapList(articleIds []uint64) map[uint64][]string {
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryIds := array.Map(categoryRs, func(t *articleCategoryRs.Entity) uint64 {
		return t.ArticleCategoryId
	})
	categoryMap := articleCategory.GetMapByIds(categoryIds)
	// 获取文章的分类和标签
	categoriesGroup := array.GroupBy(categoryRs, func(rs *articleCategoryRs.Entity) uint64 {
		return rs.ArticleId
	})
	res := make(map[uint64][]string, len(categoriesGroup))
	for aId, ids := range categoriesGroup {
		res[aId] = array.Map(array.Map(ids, func(rs *articleCategoryRs.Entity) uint64 {
			return rs.ArticleCategoryId
		}), func(item uint64) string {
			if cateItem, ok := categoryMap[item]; ok {
				return cateItem.Category
			} else {
				return ""
			}
		})
	}
	return res
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
	c.HTML(http.StatusOK, "login.gohtml", gin.H{"title": "登录/注册 - GooseForum", "User": GetLoginUser(c)})
}

func UserProfile(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	showUser := GetUserShowByUserId(id)
	if showUser.UserId == 0 {
		errorPage(c, "用户不存在", "用户不存在")
		return
	}
	last, _ := articles.GetLatestArticlesByUserId(id, 5)
	templateData := gin.H{
		"Articles":    articlesSmallEntity2Dto(last),
		"Author":      showUser,
		"User":        GetLoginUser(c),
		"title":       showUser.Username + " - GooseForum",
		"description": showUser.Username + " 的个人简介 ",
	}
	c.HTML(http.StatusOK, "user_profile.gohtml", templateData)
}

func Sponsors(c *gin.Context) {
	c.HTML(http.StatusOK, "sponsors.gohtml", gin.H{"title": "赞助商 - GooseForum", "User": GetLoginUser(c)})
}

func buildCanonicalHref(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	baseUri := preferences.Get("server.url", host)
	return baseUri + c.Request.URL.String()
}

func getHost(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}

func Links(c *gin.Context) {

	configEntity := pageConfig.GetByPageType(FriendShipLinks)
	res := jsonopt.Decode[[]FriendLinksGroup](configEntity.Config)
	c.HTML(http.StatusOK, "links.gohtml", gin.H{
		"title":            "友情链接 - GooseForum",
		"User":             GetLoginUser(c),
		"FriendLinksGroup": res,
	})
}

func Contact(c *gin.Context) {
	c.HTML(http.StatusOK, "contact.gohtml", gin.H{"title": "友情链接 - GooseForum", "User": GetLoginUser(c)})
}

type ApplyAddLinkReq struct {
	SiteName string `json:"siteName" validate:"required"`
	SiteUrl  string `json:"siteUrl" validate:"required"`
	SiteLogo string `json:"siteLogo" validate:"required"`
	SiteDesc string `json:"siteDesc" validate:"required"`
	Contact  string `json:"contact" validate:"required"`
}

func ApplyAddLink(req component.BetterRequest[ApplyAddLinkReq]) component.Response {
	if applySheet.CantWriteNew(applySheet.ApplyAddLink, 33) {
		return component.FailResponse("今日网站已经收到很多申请，请明日再来提交")
	}
	entity := applySheet.Entity{
		UserId: req.UserId,
		ApplyUserInfo: jsonopt.Encode(map[string]any{
			"ip": "127.0.0.1",
		}),
		Type:    applySheet.ApplyAddLink,
		Title:   "友情链接申请",
		Content: jsonopt.Encode(req.Params),
	}
	applySheet.SaveOrCreateById(&entity)

	return component.SuccessResponse("")
}
