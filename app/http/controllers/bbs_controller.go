package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/comment"
	array "github.com/leancodebox/goose/collectionopt"
	"time"
)

type GetArticlesRequest struct {
	MaxId    uint64 `json:"maxId"`
	PageSize int    `json:"pageSize"`
}

type ArticlesDto struct {
	Id             uint64 `json:"id"`
	Content        string `json:"Content"`
	Title          string `json:"title"`
	LastUpdateTime string `json:"lastUpdateTime"`
}

// GetArticles 获取文章列表
func GetArticles(request GetArticlesRequest) component.Response {
	if request.PageSize == 0 {
		request.PageSize = 10
	}
	pageData := articles.GetByMaxIdPage(request.MaxId, request.PageSize)
	var maxId uint64
	if len(pageData) > 0 {
		maxId = pageData[0].Id
	}
	list := array.ArrayMap(func(t *articles.Entity) ArticlesDto {
		return ArticlesDto{
			Id:             t.Id,
			Title:          t.Title,
			Content:        t.Content,
			LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}, pageData)

	return component.SuccessResponse(map[string]any{
		"maxId": maxId,
		"list":  list,
	})
}

type GetArticlesPageRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
}

// GetArticlesPage 文章列表
func GetArticlesPage(param GetArticlesPageRequest) component.Response {
	pageData := articles.Page(articles.PageQuery{Page: param.Page, PageSize: param.PageSize})

	return component.SuccessResponse(component.DataMap{
		"list": array.ArrayMap(func(t articles.Entity) ArticlesDto {
			return ArticlesDto{Id: t.Id,
				Title:          t.Title,
				Content:        t.Content,
				LastUpdateTime: t.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}, pageData.Data),
		"size":    pageData.PageSize,
		"total":   pageData.Total,
		"current": param.Page,
	})
}

type GetArticlesDetailRequest struct {
	Id           uint64 `json:"id"`
	MaxCommentId uint64 `json:"maxCommentId"`
	PageSize     int    `json:"pageSize"`
}

type CommentDto struct {
	ArticleId  uint64 `json:"articleId"`
	UserId     uint64 `json:"userId"`
	Content    string `json:"Content"`
	CreateTime string `json:"createTime"`
}

// GetArticlesDetail 文章详情
func GetArticlesDetail(request GetArticlesDetailRequest) component.Response {
	if request.PageSize == 0 {
		request.PageSize = 10
	}
	entity := articles.Get(request.Id)
	comments := comment.GetByMaxIdPage(request.Id, request.MaxCommentId, request.PageSize)

	commentList := array.ArrayMap(func(item comment.Entity) CommentDto {
		return CommentDto{
			ArticleId:  item.ArticleId,
			UserId:     item.UserId,
			Content:    item.Content,
			CreateTime: item.CreatedAt.Format(time.RFC3339),
		}
	}, comments)
	return component.SuccessResponse(map[string]any{
		"articleTitle":   &entity.Title,
		"articleContent": &entity.Content,
		"commentList":    commentList,
	})

}

type WriteArticleReq struct {
	Id      int64  `json:"id"`
	Content string `json:"content" validate:"required"`
	Title   string `json:"title" validate:"required"`
}

// WriteArticles 写文章
func WriteArticles(req component.BetterRequest[WriteArticleReq]) component.Response {
	if articles.CantWriteNew(req.UserId, 10) {
		return component.FailResponse("您当天已发布较多，为保证质量，请明天再发布新帖")
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
	articles.Save(&article)
	return component.SuccessResponse(map[string]any{})
}

type ArticleCommentReq struct {
	ArticleId uint64 `json:"articleId"`
	Comment   string `json:"comment"`
}

// ArticleComment 文章添加评论
func ArticleComment(req component.BetterRequest[ArticleCommentReq]) component.Response {
	if articles.Get(req.Params.ArticleId).Id == 0 {
		return component.FailResponse("文章不存在")
	}
	comment.Save(&comment.Entity{Content: req.Params.Comment, UserId: req.UserId})
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
