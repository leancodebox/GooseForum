package controllers

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
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
	"github.com/spf13/cast"
)

type PostDetailV3Data struct {
	PostDetailData
	LatestArticles []vo.ArticlesSimpleDto
}

func PostDetailV3(c *gin.Context) {
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
	replyEntities := reply.GetByArticleId(entity.Id)
	userIds := array.Map(replyEntities, func(item *reply.Entity) uint64 {
		return item.UserId
	})
	userIds = append(userIds, entity.UserId)
	userMap := users.GetMapByIds(userIds)
	author := "陶渊明"
	avatarUrl := urlconfig.GetDefaultAvatar()
	authorUserInfo := users.EntityComplete{}
	authorInfoStatistics := userStatistics.Get(entity.UserId)
	if user, ok := userMap[entity.UserId]; ok {
		author = user.Username
		avatarUrl = user.GetWebAvatarUrl()
		authorUserInfo = *user
	}
	replyMap := array.Slice2Map(replyEntities, func(item *reply.Entity) uint64 {
		return item.Id
	})
	replyList := array.Map(replyEntities, func(item *reply.Entity) ReplyDto {
		username := "陶渊明"
		userAvatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[item.UserId]; ok {
			username = user.Username
			userAvatarUrl = user.GetWebAvatarUrl()
		}
		// 获取被回复评论的用户名
		replyToUsername := ""
		var replyToUserId uint64 = 0
		if item.ReplyId > 0 {
			if replyTo, ok := replyMap[item.ReplyId]; ok {
				if replyUser, ok := userMap[replyTo.UserId]; ok {
					replyToUsername = replyUser.Username
					replyToUserId = replyTo.UserId
				}
			}
		}
		return ReplyDto{
			Id:              item.Id,
			ArticleId:       item.ArticleId,
			UserId:          item.UserId,
			UserAvatarUrl:   userAvatarUrl,
			Username:        username,
			Content:         item.Content,
			CreateTime:      item.CreatedAt.Format(time.DateTime),
			ReplyToUsername: replyToUsername,
			ReplyToUserId:   replyToUserId,
		}
	})

	// 复用现有的数据获取逻辑
	authorId := entity.UserId

	if entity.RenderedVersion < markdown2html.GetVersion() || entity.RenderedHTML == "" {
		mdInfo := markdown2html.MarkdownToHTML(entity.Content)
		entity.RenderedHTML = mdInfo
		entity.RenderedVersion = markdown2html.GetVersion()
		articles.SaveNoUpdate(&entity)
	}

	authorArticles := hotdataserve.GetOrLoad(fmt.Sprintf("authorId:hot:%v", authorId), func() ([]articles.SmallEntity, error) {
		return articles.GetRecommendedArticlesByAuthorId(cast.ToUint64(authorId), 5)
	})

	categoryMap := hotdataserve.ArticleCategoryMap()
	articleCategory := array.Map(entity.CategoryId, func(item uint64) string {
		if cateItem, ok := categoryMap[item]; ok {
			return cateItem.Category
		} else {
			return ""
		}
	})
	iLike := false
	isFollowing := false
	isBookmarked := false
	currentUserId := component.LoginUserId(c)
	if currentUserId != 0 {
		iLike = articleLike.GetByArticleId(currentUserId, entity.Id).Status == 1
		// 检查是否已关注作者
		if currentUserId != entity.UserId {
			followEntity := userFollow.GetByUserId(currentUserId, entity.UserId)
			isFollowing = followEntity.Status == 1
		}
		isBookmarked = articleBookmark.GetByArticleId(currentUserId, entity.Id).Status == 1
	}

	authorCard := transform.User2UserCard(authorUserInfo, authorInfoStatistics, isFollowing, currentUserId == entity.UserId, currentUserId > 0, false)

	data := PostDetailV3Data{
		PostDetailData: PostDetailData{
			Article:              entity,
			Username:             author,
			CommentList:          replyList,
			AvatarUrl:            avatarUrl,
			AuthorArticles:       authorArticles,
			ArticleCategory:      articleCategory,
			AuthorUserInfo:       authorUserInfo,
			AuthorInfoStatistics: authorInfoStatistics,
			AuthorCard:           authorCard,
			IsOwnArticle:         currentUserId == entity.UserId,
			ArticleCategoryList:  hotdataserve.ArticleCategoryLabel(),
			ILike:                iLike,
			IsFollowing:          isFollowing,
			IsBookmarked:         isBookmarked,
		},
		LatestArticles: hotdataserve.GetLatestArticleSimpleDto(),
	}

	pageMeta := viewrender.NewPageMetaBuilder().
		SetArticle(
			entity.Title,
			entity.Description,
			author,
			articleCategory,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		).
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		SetSchemaOrg(generateArticleJSONLDV3(c, entity, author)).
		Build()

	viewrender.SafeRender(c, "detail_v3.gohtml", data, pageMeta)

	articles.IncrementView(entity)
}

// generateArticleJSONLDV3 生成文章的JSON-LD结构化数据
func generateArticleJSONLDV3(c *gin.Context, entity articles.Entity, author string) template.JS {
	jsonLD := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Article",
		"headline": entity.Title,
		"author": map[string]any{
			"@type": "Person",
			"name":  author,
			"url":   fmt.Sprintf("%s/user/%d", component.GetBaseUri(c), entity.UserId),
		},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  "GooseForum",
			"url":   component.GetBaseUri(c),
		},
		"datePublished": entity.CreatedAt.Format(time.RFC3339),
		"url":           component.BuildCanonicalHref(c),
		"interactionStatistic": map[string]any{
			"@type":                "InteractionCounter",
			"interactionType":      "https://schema.org/ViewAction",
			"userInteractionCount": entity.ViewCount,
		},
	}
	jsonString := jsonopt.EncodeFormat(jsonLD)
	return template.JS(jsonString)
}
