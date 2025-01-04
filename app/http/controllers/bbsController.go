package controllers

import (
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	array "github.com/leancodebox/goose/collectionopt"
)

func GetArticlesCategory() component.Response {
	res := array.Map(articleCategory.All(), func(t *articleCategory.Entity) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{
			Name:  t.Category,
			Value: t.Id,
		}
	})
	return component.SuccessResponse(res)
}

type GetArticlesPageRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
}

type ArticlesSimpleDto struct {
	Id             uint64   `json:"id"`
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	CreateTime     string   `json:"createTime"`
	LastUpdateTime string   `json:"lastUpdateTime"`
	Username       string   `json:"username"`
	ViewCount      uint64   `json:"viewCount"`
	CommentCount   uint64   `json:"commentCount"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
}

// GetArticlesPage 文章列表
func GetArticlesPage(param GetArticlesPageRequest) component.Response {
	pageData := articles.Page(articles.PageQuery{Page: max(param.Page, 1), PageSize: param.PageSize, FilterStatus: true})
	userIds := array.Map(pageData.Data, func(t articles.Entity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)
	return component.SuccessPage(
		array.Map(pageData.Data, func(t articles.Entity) ArticlesSimpleDto {
			username := ""
			if user, _ := userMap[t.UserId]; user != nil {
				username = user.Username
			}
			return ArticlesSimpleDto{
				Id:             t.Id,
				Title:          t.Title,
				LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
				Username:       username,
			}
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
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
	Username        string `json:"username"`
	Content         string `json:"content"`
	CreateTime      string `json:"createTime"`
	ReplyToId       uint64 `json:"replyToId,omitempty"`
	ReplyToUsername string `json:"replyToUsername,omitempty"`
}

// GetArticlesDetail 文章详情
func GetArticlesDetail(req GetArticlesDetailRequest) component.Response {
	entity := articles.Get(req.Id)
	replyEntities := reply.GetByMaxIdPage(req.Id, req.MaxCommentId, boundPageSizeWithRange(req.PageSize, 10, 100))
	userIds := array.Map(replyEntities, func(item reply.Entity) uint64 {
		return item.UserId
	})
	userIds = append(userIds, entity.UserId)
	userMap := users.GetMapByIds(userIds)
	author := "陶渊明"
	if user, ok := userMap[entity.UserId]; ok {
		author = user.Username
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
	return component.SuccessResponse(map[string]any{
		"userId":         entity.UserId,
		"username":       author,
		"articleTitle":   entity.Title,
		"articleContent": entity.Content,
		"commentList":    replyList,
		"replyList":      replyList,
	})

}

type WriteArticlesOriginReq struct {
	Id int64 `json:"id"`
}

// WriteArticlesOrigin 写文章
func WriteArticlesOrigin(req component.BetterRequest[WriteArticleReq]) component.Response {
	entity := articles.Get(req.Params.Id)
	if entity.Id == 0 {
		return component.FailResponse("不存在")
	}
	if entity.UserId != req.UserId {
		return component.FailResponse("不存在")
	}
	return component.SuccessResponse(map[string]any{
		"userId":         entity.UserId,
		"articleTitle":   entity.Title,
		"articleContent": entity.Content,
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
	if article.Id > 0 {
		articles.Save(&article)
	} else {
		articles.Create(&article)
		for _, item := range req.Params.CategoryId {
			rs := articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: item}
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

	pointservice.RewardPoints(req.UserId, 2, pointservice.RewardPoints4Reply)
	return component.SuccessResponse(true)
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
	pageData := articles.Page(articles.PageQuery{
		Page:         max(req.Params.Page, 1),
		PageSize:     req.Params.PageSize,
		UserId:       req.UserId,
		FilterStatus: true,
	})

	//获取文章的分类信息
	articleIds := array.Map(pageData.Data, func(t articles.Entity) uint64 {
		return t.Id
	})
	categoryRs := articleCategoryRs.GetByArticleIds(articleIds)
	categoryIds := array.Map(categoryRs, func(t *articleCategoryRs.Entity) uint64 {
		return t.ArticleCategoryId
	})
	categoryMap := articleCategory.GetMapByIds(categoryIds)

	return component.SuccessPage(
		array.Map(pageData.Data, func(t articles.Entity) ArticlesSimpleDto {
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
				Content:        t.Content,
				CreateTime:     t.CreatedAt.Format("2006-01-02 15:04:05"),
				LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
				Username:       "", // 这里不需要用户名，因为是自己的文章
				ViewCount:      t.ViewCount,
				CommentCount:   t.ReplyCount,
				Category:       FirstOr(categoryNames, "未分类"),
				Tags:           []string{"文章", "技术"}, // 暂时使用固定标签，后续可以添加标签系统
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
