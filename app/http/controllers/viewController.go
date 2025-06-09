package controllers

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	jwt "github.com/leancodebox/GooseForum/app/bundles/jwtopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"math/rand"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/validate"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"github.com/spf13/cast"

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

// Register 注册
func Register(c *gin.Context) {
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
	if !captchaOpt.VerifyCaptcha(r.CaptchaId, r.CaptchaCode) {
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
	userEntity.Nickname = generateGooseNickname()
	err := users.Create(userEntity)
	if err != nil {
		c.JSON(200, component.FailData("注册失败"))
	}

	if err = SendAEmail4User(userEntity); err != nil {
		slog.Error("添加邮件任务到队列失败", "error", err)
	}

	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)

	if userEntity.Id == 1 {
		// For the first user registered, elevate it to admin group.
		userservice.FirstUserInit(userEntity)
		WriteArticles(component.BetterRequest[WriteArticleReq]{
			Params: WriteArticleReq{
				Id:         0,
				Content:    userservice.GetInitBlog(),
				Title:      "Hi With GooseForum",
				Type:       1,
				CategoryId: []uint64{1},
			},
			UserId: userEntity.Id,
		})
	}

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

type LoginReq struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, component.FailData("验证失败"))
		return
	}
	username := req.Username
	password := req.Password
	captchaId := req.CaptchaId
	captchaCode := req.CaptchaCode

	if !captchaOpt.VerifyCaptcha(captchaId, captchaCode) {
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
		Bio:        user.Bio,
		Signature:  user.Signature,
		Prestige:   user.Prestige,
		AvatarUrl:  avatarUrl,
		CreateTime: user.CreatedAt,
		IsAdmin:    user.RoleId > 0,
		//UserPoint: userPoint.CurrentPoints,
	}
}

type PageButton struct {
	Index int
	Page  int
}

func articleCategoryMapList(articleIds []uint64) map[uint64][]string {
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryMap := articleCategoryMap()
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

// 新增生成鹅相关昵称的函数
func generateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]
	// 使用纳秒级时间戳+随机数确保唯一性
	now := time.Now()
	timestamp := now.UnixNano()
	randomPart := rand.Intn(1000)
	// 组合成16进制字符串
	return fmt.Sprintf("%s%x%03d", prefix, timestamp, randomPart)
}
