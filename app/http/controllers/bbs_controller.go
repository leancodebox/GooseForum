package controllers

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	array "github.com/leancodebox/goose/collectionopt"
	"time"
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
	Id             uint64 `json:"id"`
	Title          string `json:"title"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Username       string `json:"username"`
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
	ArticleId  uint64 `json:"articleId"`
	UserId     uint64 `json:"userId"`
	Username   string `json:"username"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
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
		return ReplyDto{
			ArticleId:  item.ArticleId,
			UserId:     item.UserId,
			Username:   username,
			Content:    item.Content,
			CreateTime: item.CreatedAt.Format(time.RFC3339),
		}
	})
	return component.SuccessResponse(map[string]any{
		"userId":         entity.UserId,
		"username":       author,
		"articleTitle":   entity.Title,
		"articleContent": entity.Content,
		"commentList":    replyList,
		"replyList":      replyList,
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
	if articles.Get(req.Params.ArticleId).Id == 0 {
		return component.FailResponse("文章不存在")
	}
	if req.Params.ReplyId > 0 && reply.Get(req.Params.ReplyId).Id == 0 {
		return component.FailResponse("要回复的评论不存在")
	}
	reply.Create(&reply.Entity{ArticleId: req.Params.ArticleId, Content: req.Params.Content, UserId: req.UserId})
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
