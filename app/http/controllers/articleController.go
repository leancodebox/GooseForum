package controllers

import (
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/service/eventnotice"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"strings"
	"time"
)

func GetSiteStatistics() component.Response {
	return component.SuccessResponse(GetSiteStatisticsData())
}

func GetArticlesEnum() component.Response {
	res := articleCategoryLabel()
	return component.SuccessResponse(map[string]any{
		"category": res,
		"type":     articlesType,
	})
}

type GetArticlesPageRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	Search     string `form:"search"`
	Categories []int  `form:"categories"`
}

type ArticlesSimpleDto struct {
	Id             uint64   `json:"id"`
	Title          string   `json:"title,omitempty"`
	Content        string   `json:"content,omitempty"`
	CreateTime     string   `json:"createTime,omitempty"`
	LastUpdateTime string   `json:"lastUpdateTime,omitempty"`
	Username       string   `json:"username,omitempty"`
	AuthorId       uint64   `json:"authorId,omitempty"`
	ViewCount      uint64   `json:"viewCount,omitempty"`
	CommentCount   uint64   `json:"commentCount"`
	Type           int8     `json:"type,omitempty"`
	TypeStr        string   `json:"typeStr,omitempty"`
	Category       string   `json:"category,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	CategoriesId   []uint64 `json:"categoriesId,omitempty"`
	AvatarUrl      string   `json:"avatarUrl,omitempty"`
}

type GetArticlesDetailRequest struct {
	Id           uint64 `json:"id"`
	MaxCommentId uint64 `json:"maxCommentId"`
	PageSize     int    `json:"pageSize"`
}

type ReplyDto struct {
	Id              uint64 `json:"id"`
	ArticleId       uint64 `json:"articleId"`
	UserId          uint64 `json:"userId"`
	UserAvatarUrl   string `json:"userAvatarUrl"`
	Username        string `json:"username"`
	Content         string `json:"content"`
	CreateTime      string `json:"createTime"`
	ReplyToId       uint64 `json:"replyToId,omitempty"`
	ReplyToUsername string `json:"replyToUsername,omitempty"`
}

type WriteArticlesOriginReq struct {
	Id int64 `json:"id"`
}

// WriteArticlesOrigin 写文章
func WriteArticlesOrigin(req component.BetterRequest[WriteArticlesOriginReq]) component.Response {
	entity := articles.Get(uint64(req.Params.Id))
	if entity.Id == 0 {
		return component.FailResponse("不存在")
	}
	if entity.UserId != req.UserId {
		return component.FailResponse("不存在")
	}
	categoryRs := articleCategoryRs.GetByArticleIdsEffective([]uint64{entity.Id})

	return component.SuccessResponse(map[string]any{
		"userId":         entity.UserId,
		"type":           entity.Type,
		"articleTitle":   entity.Title,
		"articleContent": entity.Content,
		"categoryId": array.Map(categoryRs, func(rs *articleCategoryRs.Entity) uint64 {
			return rs.ArticleCategoryId
		}),
	})
}

type WriteArticleReq struct {
	Id         int64    `json:"id"`
	Content    string   `json:"content" validate:"required"`
	Title      string   `json:"title" validate:"required"`
	Type       int8     `json:"type"`
	CategoryId []uint64 `json:"categoryId" validate:"min=1,max=3"`
}

// WriteArticles 写文章
func WriteArticles(req component.BetterRequest[WriteArticleReq]) component.Response {
	if articles.CantWriteNew(req.UserId, 10) {
		return component.FailResponse("您当天已发布较多，为保证质量，请明天再发布新文章")
	}
	var article articles.Entity
	if req.Params.Id != 0 {
		article = articles.Get(req.Params.Id)
		if article.UserId != req.UserId {
			return component.FailResponse("不要更改别人发出的帖子哦")
		}
	} else {
		article.UserId = req.UserId
		article.Type = req.Params.Type
	}
	article.ArticleStatus = 1
	article.Content = req.Params.Content
	article.Title = req.Params.Title
	article.RenderedVersion = markdown2html.GetVersion()
	article.RenderedHTML = "" // 用户提交后不用渲染，避免提交时间过长。
	if article.Id > 0 {
		articles.Save(&article)
		categoryIDMap := make(map[uint64]bool)
		for _, id := range req.Params.CategoryId {
			categoryIDMap[id] = true
		}

		// 遍历 rsList，更新或删除无效的条目
		for _, item := range articleCategoryRs.GetByArticleId(article.Id) {
			if _, ok := categoryIDMap[item.ArticleCategoryId]; ok {
				item.Effective = 1
				articleCategoryRs.SaveOrCreateById(item)
				// 如果已经存在，从 map 中删除，避免重复插入
				delete(categoryIDMap, item.ArticleCategoryId)
			} else {
				item.Effective = 0
				articleCategoryRs.SaveOrCreateById(item)
			}
		}
		// 插入新的条目
		for id, _ := range categoryIDMap {
			rs := &articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: id, Effective: 1}
			fmt.Println(*rs)
			articleCategoryRs.SaveOrCreateById(rs)
		}
	} else {
		articles.Create(&article)
		userStatistics.WriteArticle(req.UserId)
		for _, item := range req.Params.CategoryId {
			rs := articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: item, Effective: 1}
			articleCategoryRs.SaveOrCreateById(&rs)
		}
		pointservice.RewardPoints(req.UserId, 10, pointservice.RewardPoints4WriteArticles)
	}
	return component.SuccessResponse(article.Id)
}

type ArticleReplyId struct {
	ArticleId uint64 `json:"articleId"`
	Content   string `json:"content"`
	ReplyId   uint64 `json:"replyId"`
}

func ArticleReply(req component.BetterRequest[ArticleReplyId]) component.Response {
	if len(strings.TrimSpace(req.Params.Content)) == 0 {
		return component.FailResponse("评论内容不能为空")
	}

	articleEntity := articles.Get(req.Params.ArticleId)
	if articleEntity.Id == 0 {
		return component.FailResponse("文章不存在")
	}

	if req.Params.ReplyId > 0 && reply.Get(req.Params.ReplyId).Id == 0 {
		return component.FailResponse("要回复的评论不存在")
	}

	replyEntity := &reply.Entity{
		ArticleId: req.Params.ArticleId,
		Content:   req.Params.Content,
		UserId:    req.UserId,
		ReplyId:   req.Params.ReplyId,
	}

	err := reply.Create(replyEntity)
	if err != nil {
		return component.FailResponse("评论失败:" + err.Error())
	}
	articles.IncrementReply(articleEntity)
	userStatistics.WriteComment(req.UserId)
	pointservice.RewardPoints(req.UserId, 2, pointservice.RewardPoints4Reply)
	if articleEntity.UserId != req.UserId {
		eventnotice.SendCommentNotification(articleEntity.UserId, articleEntity.Id,
			TakeUpTo64Chars(req.Params.Content), req.UserId)
	}
	return component.SuccessResponse(true)
}

// TakeUpTo64Chars 按字符数截取字符串，最多取 64 个字符
func TakeUpTo64Chars(s string) string {
	// 将字符串转换为 rune 切片
	runes := []rune(s)
	// 如果 rune 切片的长度超过 64 个字符，截取前 64 个字符
	if len(runes) > 64 {
		return string(runes[:64])
	}
	// 否则返回整个字符串
	return s
}

type DeleteReplyId struct {
	ReplyId uint64 `json:"repoId"`
}

func DeleteReply(req component.BetterRequest[DeleteReplyId]) component.Response {
	replyEntity := reply.Get(req.Params.ReplyId)
	if replyEntity.Id == 0 {
		return component.FailResponse("回复不存在")
	}
	if replyEntity.UserId != req.UserId {
		return component.FailResponse("不可操作")
	}
	reply.DeleteEntity(&replyEntity)
	return component.SuccessResponse(true)
}

// 添加新的请求结构体
type GetUserArticlesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GetUserArticles 获取用户文章列表
func GetUserArticles(req component.BetterRequest[GetUserArticlesRequest]) component.Response {
	pageData := articles.Page[articles.SmallEntity](articles.PageQuery{
		Page:         max(req.Params.Page, 1),
		PageSize:     req.Params.PageSize,
		UserId:       req.UserId,
		FilterStatus: true,
	})

	//获取文章的分类信息
	articleIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.Id
	})
	categoryRs := articleCategoryRs.GetByArticleIdsEffective(articleIds)
	categoryMap := articleCategoryMap()

	return component.SuccessPage(
		array.Map(pageData.Data, func(t articles.SmallEntity) ArticlesSimpleDto {
			// 获取文章的分类和标签
			categories := array.Filter(categoryRs, func(rs *articleCategoryRs.Entity) bool {
				return rs.ArticleId == t.Id
			})
			categoryNames := array.Map(categories, func(rs *articleCategoryRs.Entity) string {
				if category, ok := categoryMap[rs.ArticleCategoryId]; ok {
					return category.Category
				}
				return ""
			})
			return ArticlesSimpleDto{
				Id:             t.Id,
				Title:          t.Title,
				CreateTime:     t.CreatedAt.Format(time.DateTime),
				LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
				Username:       "", // 这里不需要用户名，因为是自己的文章
				ViewCount:      t.ViewCount,
				CommentCount:   t.ReplyCount,
				Category:       FirstOr(categoryNames, "未分类"),
				Categories:     categoryNames,
				TypeStr:        articlesTypeMap[int(t.Type)].Name,
			}
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}
func FirstOr[T any](d []T, defaultValue T) T {
	if len(d) > 1 {
		return d[0]
	}
	return defaultValue
}

type LikeArticleReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 like 2 cancel
}

func LikeArticle(req component.BetterRequest[LikeArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponse("文章不存在")
	}
	oldLike := articleLike.GetByArticleId(req.UserId, articleEntity.Id)
	targetStatus := 0
	if req.Params.Action == 1 {
		if oldLike.Id == 0 {
			oldLike.UserId = req.UserId
			oldLike.ArticleId = articleEntity.Id
		}
		targetStatus = 1
	} else {
		if oldLike.Id == 0 {
			return component.SuccessResponse(true)
		}
		targetStatus = 0
	}
	if oldLike.Status == targetStatus {
		return component.SuccessResponse(true)
	}
	oldLike.Status = targetStatus
	if articleLike.SaveOrCreateById(&oldLike) > 0 {
		if req.Params.Action == 1 {
			articles.IncrementLike(articleEntity)
			userStatistics.LikeArticle(req.UserId)
			userStatistics.GivenLike(articleEntity.UserId)
		} else {
			articles.DecrementLike(articleEntity)
			userStatistics.CancelLikeArticle(req.UserId)
			userStatistics.CancelGivenLike(articleEntity.UserId)
		}
	}
	return component.SuccessResponse(true)
}
