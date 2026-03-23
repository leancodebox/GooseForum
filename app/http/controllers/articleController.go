package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/articleBookmark"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/articlesUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func GetSiteStatistics() component.Response {
	return component.SuccessResponse(hotdataserve.GetSiteStatisticsData())
}

func GetArticlesEnum() component.Response {
	res := hotdataserve.ArticleCategoryLabel()
	return component.SuccessResponse(map[string]any{
		"category": res,
		"type":     hotdataserve.GetArticlesType(),
	})
}

type GetArticlesPageRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	Search     string `form:"search"`
	Categories []int  `form:"categories"`
}

type GetArticlesDetailRequest struct {
	Id           uint64 `json:"id"`
	MaxCommentId uint64 `json:"maxCommentId"`
	PageSize     int    `json:"pageSize"`
}

type ReplyVo struct {
	Id              uint64 `json:"id"`
	ArticleId       uint64 `json:"articleId"`
	UserId          uint64 `json:"userId"`
	UserAvatarUrl   string `json:"userAvatarUrl"`
	Username        string `json:"username"`
	Content         string `json:"content"`
	CreateTime      string `json:"createTime"`
	ReplyToId       uint64 `json:"replyToId,omitempty"`
	ReplyToUsername string `json:"replyToUsername,omitempty"`
	ReplyToUserId   uint64 `json:"replyToUserId,omitempty"`
	IsOwnReply      bool   `json:"isOwnReply"`
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

	return component.SuccessResponse(map[string]any{
		"userId":         entity.UserId,
		"type":           entity.Type,
		"articleTitle":   entity.Title,
		"articleContent": entity.Content,
		"categoryId":     entity.CategoryId,
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
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	if len(req.Params.Title) < postingConfig.TextControl.MinTitleLength {
		return component.FailResponse(fmt.Sprintf("标题长度不能少于%d位", postingConfig.TextControl.MinTitleLength))
	}

	if len(req.Params.Content) < postingConfig.TextControl.MinPostLength {
		return component.FailResponse(fmt.Sprintf("正文长度不能少于%d位", postingConfig.TextControl.MinPostLength))
	}

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
	article.CategoryId = req.Params.CategoryId
	article.ArticleStatus = 1
	article.Content = req.Params.Content
	article.Title = req.Params.Title
	// 自动生成文章描述
	article.Description = markdown2html.ExtractDescription(req.Params.Content, 200)
	article.RenderedVersion = markdown2html.GetVersion()
	article.RenderedHTML = "" // 用户提交后不用渲染，避免提交时间过长。
	if article.Id > 0 {
		articles.Save(&article)
		categoryIDMap := lo.SliceToMap(req.Params.CategoryId, func(id uint64) (uint64, bool) {
			return id, true
		})

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
		for id := range categoryIDMap {
			rs := &articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: id, Effective: 1}
			fmt.Println(*rs)
			articleCategoryRs.SaveOrCreateById(rs)
		}
		eventbus.Publish(context.Background(), &eventhandlers.ArticleUpdatedEvent{Article: &article})
	} else {
		articles.Create(&article)
		userStatistics.WriteArticle(req.UserId)
		lo.ForEach(req.Params.CategoryId, func(item uint64, _ int) {
			rs := articleCategoryRs.Entity{ArticleId: article.Id, ArticleCategoryId: item, Effective: 1}
			articleCategoryRs.SaveOrCreateById(&rs)
		})
		eventbus.Publish(context.Background(), &eventhandlers.ArticlePublishedEvent{Article: &article})
	}
	return component.SuccessResponse(article.Id)
}

type DeleteArticleReq struct {
	Id uint64 `json:"id"`
}

func DeleteArticle(req component.BetterRequest[DeleteArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponse("文章不存在")
	}
	if articleEntity.UserId != req.UserId {
		return component.FailResponse("不可操作")
	}
	articles.Delete(&articleEntity)
	articleCategoryRs.DisableByArticleId(articleEntity.Id)
	return component.SuccessResponse(true)
}

type ArticleReplyId struct {
	ArticleId uint64 `json:"articleId"`
	Content   string `json:"content"`
	ReplyId   uint64 `json:"replyId"`
}

func ArticleReply(req component.BetterRequest[ArticleReplyId]) component.Response {
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	if len(strings.TrimSpace(req.Params.Content)) < postingConfig.TextControl.MinPostLength {
		return component.FailResponse(fmt.Sprintf("评论内容长度不能少于%d位", postingConfig.TextControl.MinPostLength))
	}

	articleEntity := articles.GetSimple(req.Params.ArticleId)
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
	userStatistics.WriteComment(req.UserId)
	updateArticleStat(articleEntity, req.UserId, false)

	// 获取父评论作者ID
	var parentReplyAuthorId uint64
	if req.Params.ReplyId > 0 {
		parentReply := reply.Get(req.Params.ReplyId)
		parentReplyAuthorId = parentReply.UserId
	}

	// 发布统一的评论创建事件
	eventbus.Publish(context.Background(), &eventhandlers.CommentCreatedEvent{
		ArticleId:           articleEntity.Id,
		CommentId:           replyEntity.Id,
		UserId:              req.UserId,
		Content:             req.Params.Content,
		ArticleAuthorId:     articleEntity.UserId,
		ParentReplyId:       req.Params.ReplyId,
		ParentReplyAuthorId: parentReplyAuthorId,
	})

	return component.SuccessResponse(true)
}

type DeleteReplyId struct {
	ReplyId uint64 `json:"replyId"`
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
	articleEntity := articles.GetSimple(replyEntity.ArticleId)
	if articleEntity.Id > 0 {
		updateArticleStat(articleEntity, req.UserId, true)
	}
	return component.SuccessResponse(true)
}

func updateArticleStat(article articles.SmallEntity, userId uint64, isDelete bool) {
	if isDelete {
		articlesUserStat.DecrementUserReply(article.Id, userId)
	} else {
		articlesUserStat.IncrementUserReply(article.Id, userId)
	}
	list := articlesUserStat.SyncArticlePosters(article.Id)

	// 过滤掉作者ID，避免重复
	filteredList := lo.Filter(list, func(t uint64, _ int) bool {
		return t != article.UserId
	})

	// 将作者ID放到第一位
	finalList := append([]uint64{article.UserId}, filteredList...)

	pList := lo.Map(finalList, func(t uint64, _ int) articles.Poster {
		return articles.Poster{
			UserID: t,
		}
	})
	if isDelete {
		articles.DecrementReplyFast(article.Id, pList)
	} else {
		articles.IncrementReplyFast(article.Id, pList)
	}
}

// GetUserArticlesRequest 添加新的请求结构体
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
	authorInfoStatistics := userStatistics.Get(req.UserId)
	user, _ := req.GetUser()
	categoryMap := hotdataserve.ArticleCategoryMap()
	return component.SuccessPage(
		lo.Map(pageData.Data, func(t articles.SmallEntity, _ int) vo.ArticlesSimpleVo {

			categoryNames := lo.Map(t.CategoryId, func(t uint64, _ int) string {
				if category, ok := categoryMap[t]; ok {
					return category.Category
				}
				return ""
			})

			// 获取作者信息（虽然是当前用户，为了前端统一处理，也返回完整信息）
			username := user.Username
			avatarUrl := user.GetWebAvatarUrl()

			return vo.ArticlesSimpleVo{
				Id:             t.Id,
				Title:          t.Title,
				CreateTime:     t.CreatedAt.Format(time.DateTime),
				LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
				Username:       username,
				AuthorId:       t.UserId,
				AvatarUrl:      avatarUrl,
				ViewCount:      t.ViewCount,
				CommentCount:   t.ReplyCount,
				Categories:     categoryNames,
				TypeStr:        hotdataserve.GetArticlesTypeName(int(t.Type)),
			}
		}),
		pageData.Page,
		pageData.PageSize,
		cast.ToInt64(authorInfoStatistics.ArticleCount),
	)
}

// GetUserBookmarkedArticlesRequest 获取用户收藏文章列表请求结构体
type GetUserBookmarkedArticlesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GetUserBookmarkedArticles 获取用户收藏文章列表
func GetUserBookmarkedArticles(req component.BetterRequest[GetUserBookmarkedArticlesRequest]) component.Response {
	// 获取收藏的文章ID列表
	articleIds, total := articleBookmark.GetUserBookmarkedArticleIds(req.UserId, max(req.Params.Page, 1), req.Params.PageSize)
	if len(articleIds) == 0 {
		return component.SuccessPage(
			[]vo.ArticlesSimpleVo{},
			max(req.Params.Page, 1),
			req.Params.PageSize,
			total,
		)
	}
	authorInfoStatistics := userStatistics.Get(req.UserId)

	// 根据文章ID获取文章详情
	articleList := articles.GetByIds(articleIds)

	// 获取作者信息
	userIds := lo.Map(articleList, func(t *articles.SmallEntity, _ int) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)

	categoryMap := hotdataserve.ArticleCategoryMap()

	// 构建返回数据
	articleVos := lo.Map(articleList, func(t *articles.SmallEntity, _ int) vo.ArticlesSimpleVo {
		categoryNames := lo.Map(t.CategoryId, func(item uint64, _ int) string {
			if category, ok := categoryMap[item]; ok {
				return category.Category
			}
			return ""
		})

		username := ""
		avatarUrl := ""
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			avatarUrl = user.GetWebAvatarUrl()
		}

		return vo.ArticlesSimpleVo{
			Id:             t.Id,
			Title:          t.Title,
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			Username:       username,
			AuthorId:       t.UserId,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			Categories:     categoryNames,
			TypeStr:        hotdataserve.GetArticlesTypeName(int(t.Type)),
		}
	})

	return component.SuccessPage(
		articleVos,
		max(req.Params.Page, 1),
		req.Params.PageSize,
		cast.ToInt64(authorInfoStatistics.CollectionCount),
	)
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
			userStatistics.LikeArticle(articleEntity.UserId)
			userStatistics.GivenLike(req.UserId)

			// 发送点赞事件
			eventbus.Publish(context.Background(), &eventhandlers.ArticleLikedEvent{
				UserId:    articleEntity.UserId,
				ArticleId: articleEntity.Id,
				Title:     articleEntity.Title,
				LikierId:  req.UserId,
			})
		} else {
			articles.DecrementLike(articleEntity)
			userStatistics.CancelLikeArticle(articleEntity.UserId)
			userStatistics.CancelGivenLike(req.UserId)
		}
	}
	return component.SuccessResponse(true)
}

type BookmarkArticleReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 Bookmark 2 cancel
}

func BookmarkArticle(req component.BetterRequest[BookmarkArticleReq]) component.Response {
	articleEntity := articles.Get(req.Params.Id)
	if articleEntity.Id == 0 {
		return component.FailResponse("文章不存在")
	}
	oldBookMark := articleBookmark.GetByArticleId(req.UserId, articleEntity.Id)
	targetStatus := 0
	if req.Params.Action == 1 {
		if oldBookMark.Id == 0 {
			oldBookMark.UserId = req.UserId
			oldBookMark.ArticleId = articleEntity.Id
		}
		targetStatus = 1
	} else {
		if oldBookMark.Id == 0 {
			return component.SuccessResponse(true)
		}
		targetStatus = 0
	}
	if oldBookMark.Status == targetStatus {
		return component.SuccessResponse(true)
	}
	oldBookMark.Status = targetStatus
	articleBookmark.SaveOrCreateById(&oldBookMark)
	return component.SuccessResponse(true)
}

type FollowUserReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 like 2 cancel
}

func FollowUser(req component.BetterRequest[FollowUserReq]) component.Response {
	userEntity, _ := users.Get(req.Params.Id)
	if userEntity.Id == 0 {
		return component.FailResponse("用户不存在")
	}
	userFollowEntity := userFollow.GetByUserId(req.UserId, req.Params.Id)
	if userFollowEntity.Id == 0 {
		userFollowEntity.UserId = req.UserId
		userFollowEntity.FollowUserId = req.Params.Id
	}
	targetStatus := 0
	if req.Params.Action == 1 {
		targetStatus = 1
	} else {
		targetStatus = 0
	}

	if userFollowEntity.Status == targetStatus {
		return component.SuccessResponse(true)
	}
	userFollowEntity.Status = targetStatus
	if userFollow.SaveOrCreateById(&userFollowEntity) > 0 {
		if req.Params.Action == 1 {
			userStatistics.Following(req.UserId)
			userStatistics.Follower(req.Params.Id)

			// 发送关注通知
			followerUser, _ := req.GetUser()
			eventbus.Publish(context.Background(), &eventhandlers.UserFollowedEvent{
				UserId:       req.Params.Id,
				FollowerId:   req.UserId,
				FollowerName: followerUser.Username,
			})
		} else {
			userStatistics.CancelFollowing(req.UserId)
			userStatistics.CancelFollower(req.Params.Id)
		}
	}
	return component.SuccessResponse(true)
}
