package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	array "github.com/leancodebox/goose/collectionopt"
	"time"
)

type GetArticlesPageRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
	UserId   uint64 `form:"userId"`
}

type ArticlesSimpleDto struct {
	Id             uint64 `json:"id"`
	Title          string `json:"title"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Username       string `json:"username"`
}

// GetArticlesPage 文章列表
func GetArticlesPage(param GetArticlesPageRequest) component.Response {
	pageData := articles.Page(articles.PageQuery{Page: max(param.Page, 1), PageSize: param.PageSize, UserId: param.UserId})
	userIds := array.ArrayMap(func(t articles.Entity) uint64 {
		return t.UserId
	}, pageData.Data)
	userMap := users.GetMapByIds(userIds)
	return component.SuccessPage(
		array.ArrayMap(func(t articles.Entity) ArticlesSimpleDto {
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
		}, pageData.Data),
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
	Content    string `json:"Content"`
	CreateTime string `json:"createTime"`
}

// GetArticlesDetail 文章详情
func GetArticlesDetail(req GetArticlesDetailRequest) component.Response {
	entity := articles.Get(req.Id)
	replyEntities := reply.GetByMaxIdPage(req.Id, req.MaxCommentId, boundPageSize(req.PageSize))
	commentList := array.ArrayMap(func(item reply.Entity) ReplyDto {
		return ReplyDto{
			ArticleId:  item.ArticleId,
			UserId:     item.UserId,
			Content:    item.Content,
			CreateTime: item.CreatedAt.Format(time.RFC3339),
		}
	}, replyEntities)
	return component.SuccessResponse(map[string]any{
		"articleTitle":   &entity.Title,
		"articleContent": &entity.Content,
		"commentList":    commentList,
	})

}

type WriteArticleReq struct {
	Id          int64    `json:"id"`
	Content     string   `json:"content" validate:"required"`
	HtmlContent string   `json:"htmlContent" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Tags        []uint64 `json:"tags"`
	CategoryId  uint64   `json:"categoryId"`
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
	}
	article.Content = req.Params.Content
	article.Title = req.Params.Title
	if article.Id > 0 {
		articles.Save(&article)
	} else {
		articles.Create(&article)
		pointservice.RewardPoints(req.UserId, 10, pointservice.RewardPoints4WriteArticles)
	}
	return component.SuccessResponse(map[string]any{})
}

type ArticleReplyId struct {
	ArticleId uint64 `json:"articleId"`
	Comment   string `json:"comment"`
	ReplyId   uint64 `json:"repoId"`
}

func ArticleReply(req component.BetterRequest[ArticleReplyId]) component.Response {
	if articles.Get(req.Params.ArticleId).Id == 0 {
		return component.FailResponse("文章不存在")
	}
	if req.Params.ReplyId > 0 && reply.Get(req.Params.ReplyId).Id == 0 {
		return component.FailResponse("要回复的评论不存在")
	}
	reply.Create(&reply.Entity{Content: req.Params.Comment, UserId: req.UserId})
	pointservice.RewardPoints(req.UserId, 10, pointservice.RewardPoints4Reply)
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

type ApplyShowReq struct {
	Title     string   `json:"comment"`
	Desc      string   `json:"desc"`
	ImageList []string `json:"imageList"`
}

// ApplyShow 申请展示
// todo 低优先级 是可以考虑两个地方，一个地方是侧边栏，一个地方是置顶
func ApplyShow(req component.BetterRequest[ApplyShowReq]) component.Response {
	return component.SuccessResponse("success")
}
