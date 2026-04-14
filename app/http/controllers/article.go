package controllers

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/datastruct/common"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func PostDetail(c *gin.Context) {
	id := cast.ToUint64(c.Param("id"))
	if id == 0 {
		errorPage(c, "页面不存在", "页面不存在")
		return
	}

	entity := articles.Get(id)
	if entity.Id == 0 {
		errorPage(c, "文章不存在", "文章不存在")
		return
	}

	if entity.ProcessStatus != 0 && !component.GetLoginUser(c).IsAdmin {
		errorPage(c, "文章不存在", "文章不存在")
		return
	}

	currentUserId := component.LoginUserId(c)
	replyEntities := reply.GetByArticleId(entity.Id)

	// 1. 收集所有相关的用户 ID 并获取 UserMap
	userIds := []uint64{entity.UserId}
	userIds = append(userIds, lo.Map(replyEntities, func(r *reply.Entity, _ int) uint64 { return r.UserId })...)
	userIds = append(userIds, lo.Map(entity.Posters, func(p articles.Poster, _ int) uint64 { return p.UserID })...)
	userMap := users.GetMapByIds(lo.Uniq(userIds))

	// 辅助闭包：获取用户信息
	getUserInfo := func(userId uint64) (string, string, *users.EntityComplete) {
		if u, ok := userMap[userId]; ok {
			return u.Username, u.GetWebAvatarUrl(), u
		}
		return common.DefaultUserName, urlconfig.GetDefaultAvatar(), &users.EntityComplete{}
	}

	// 2. 准备作者信息
	author, avatarUrl, authorUserInfo := getUserInfo(entity.UserId)
	authorInfoStatistics := userStatistics.Get(entity.UserId)

	// 3. 准备活跃用户 (Posters)
	posterVos := lo.Map(entity.Posters, func(p articles.Poster, _ int) vo.PosterVo {
		name, avatar, _ := getUserInfo(p.UserID)
		return vo.PosterVo{Id: p.UserID, Username: name, AvatarUrl: avatar}
	})

	// 4. 准备评论列表
	replyMap := lo.KeyBy(replyEntities, func(r *reply.Entity) uint64 { return r.Id })
	replyList := lo.Map(replyEntities, func(item *reply.Entity, _ int) ReplyVo {
		name, avatar, _ := getUserInfo(item.UserId)
		replyToName, replyToId := "", uint64(0)
		if item.ReplyId > 0 {
			if parent, ok := replyMap[item.ReplyId]; ok {
				replyToName, _, _ = getUserInfo(parent.UserId)
				replyToId = parent.UserId
			}
		}
		return ReplyVo{
			Id: item.Id, ArticleId: item.ArticleId, UserId: item.UserId,
			UserAvatarUrl: avatar, Username: name, Content: item.Content,
			CreateTime:      item.CreatedAt.Format(time.DateTime),
			ReplyToUsername: replyToName, ReplyToUserId: replyToId,
			IsOwnReply: currentUserId == item.UserId,
		}
	})

	// 5. 渲染 Markdown (按需)
	if entity.RenderedVersion < markdown2html.GetVersion() || entity.RenderedHTML == "" {
		entity.RenderedHTML = markdown2html.MarkdownToHTML(entity.Content)
		entity.RenderedVersion = markdown2html.GetVersion()
		articles.SaveNoUpdate(&entity)
	}

	// 6. 其它数据 (推荐、分类、交互状态)
	authorArticles := hotdataserve.GetOrLoad(fmt.Sprintf("authorId:hot:%v", entity.UserId), func() ([]*articles.SmallEntity, error) {
		return articles.GetRecommendedArticlesByAuthorId(entity.UserId, 5)
	})

	categoryMap := hotdataserve.ArticleCategoryMap()
	articleCategory := lo.FilterMap(entity.CategoryId, func(id uint64, _ int) (string, bool) {
		category, ok := categoryMap[id]
		return category.Category, ok
	})

	iLike, isFollowing, isBookmarked := false, false, false
	if currentUserId != 0 {
		iLike = articleLike.GetByArticleId(currentUserId, entity.Id).Status == 1
		if currentUserId != entity.UserId {
			isFollowing = userFollow.GetByUserId(currentUserId, entity.UserId).Status == 1
		}
		isBookmarked = articleBookmark.GetByArticleId(currentUserId, entity.Id).Status == 1
	}

	// 7. 渲染页面
	CommonData := GetCommonData(c)
	CommonData.Sidebar.SetActive("topics")
	if len(entity.CategoryId) > 0 {
		if cat := hotdataserve.GetCategoryById(entity.CategoryId[0]); cat != nil {
			CommonData.Category = cat
			CommonData.Sidebar.SetActiveCategory(cat.Id)
		}
	}

	viewrender.SafeRender(c, "detail.gohtml", PostDetailDataVo{
		CommonDataVo: CommonData,
		PostDetailData: PostDetailData{
			Article: entity, Username: author, CommentList: replyList, AvatarUrl: avatarUrl,
			AuthorArticles: authorArticles, ArticleCategory: articleCategory,
			AuthorUserInfo: *authorUserInfo, AuthorInfoStatistics: authorInfoStatistics,
			IsOwnArticle: currentUserId == entity.UserId, ArticleCategoryList: CommonData.ArticleCategoryList,
			ILike: iLike, IsFollowing: isFollowing, IsBookmarked: isBookmarked, Posters: posterVos,
		},
		LatestArticles:      hotdataserve.GetLatestArticleSimpleVo(),
		ArticleCategoryList: CommonData.ArticleCategoryList,
	}, viewrender.NewPageMetaBuilder().
		SetArticle(entity.Title, entity.Description, author, articleCategory, &entity.CreatedAt, &entity.UpdatedAt).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		SetSchemaOrg(generateArticleJSONLD(c, entity, author)).
		Build())

	articles.IncrementView(entity)
}

// generateArticleJSONLD 生成文章的JSON-LD结构化数据
func generateArticleJSONLD(c *gin.Context, entity articles.Entity, author string) template.JS {
	jsonLD := vo.ArticleJSONLD{
		Context:  "https://schema.org",
		Type:     "Article",
		Headline: entity.Title,
		Author: vo.Person{
			Type: "Person",
			Name: author,
			URL:  fmt.Sprintf("%s/u/%d", component.GetBaseUri(c), entity.UserId),
		},
		Publisher: vo.Organization{
			Type: "Organization",
			Name: "GooseForum",
			URL:  component.GetBaseUri(c),
		},
		DatePublished: entity.CreatedAt.Format(time.RFC3339),
		URL:           component.BuildCanonicalHref(c),
		InteractionStatistic: vo.InteractionCounter{
			Type:                 "InteractionCounter",
			InteractionType:      "https://schema.org/ViewAction",
			UserInteractionCount: entity.ViewCount,
		},
	}
	jsonString := jsonopt.EncodeFormat(jsonLD)
	return template.JS(jsonString)
}
