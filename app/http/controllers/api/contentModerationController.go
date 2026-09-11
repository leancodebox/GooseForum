package api

import (
	"context"
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

type ReviewContentReq struct {
	Id      uint64 `json:"id" validate:"required"`
	Version uint64 `json:"version"`
	Action  string `json:"action" validate:"oneof=approve reject recheck"`
	Reason  string `json:"reason" validate:"required,max=300"`
}

func ReviewAdminTopic(req component.BetterRequest[ReviewContentReq]) component.Response {
	topic := topics.Get(req.Params.Id)
	if topic.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	if topic.ModerationVersion != req.Params.Version {
		return component.FailResponseError(fmt.Errorf("内容已更新，请刷新后重新审核"))
	}
	// Never publish a private draft through an approval action.
	if topic.Status == 0 && (topic.ModerationStatus == "none" || topic.ModerationStatus == "approved" || topic.ModerationStatus == "") {
		return component.FailResponseError(fmt.Errorf("草稿不能通过审核操作发布"))
	}
	post := posts.Get(topic.FirstPostId)
	if post.Id == 0 {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	legacyHidden := post.ModerationStatus == "rejected" || post.ModerationStatus == "pending"
	firstPublication := topic.PublishedAt == nil
	now := time.Now()
	if req.Params.Action == "recheck" {
		if !sensitivewordservice.Enabled() {
			return component.FailResponseError(fmt.Errorf("请先启用敏感词检测"))
		}
		topic.Status = 1
		contentmoderationservice.ReviewTopic(&topic, &post)
	} else {
		topic.ModerationVersion++
		post.ModerationVersion++
		topic.ModerationStatus, topic.Status = "approved", 1
		if req.Params.Action == "reject" {
			topic.ModerationStatus, topic.Status = "denied", 0
		}
		topic.ModerationReason, topic.ModeratedAt = req.Params.Reason, &now
		post.ModerationStatus, post.ModerationReason, post.ModeratedAt = topic.ModerationStatus, req.Params.Reason, &now
		// Legacy automatic review also hid the first post.
		if legacyHidden {
			post.ProcessStatus = 0
		}
	}
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds, ExpectedVersion: &req.Params.Version}); err != nil {
		return component.FailResponseError(err)
	}
	hotdataserve.ClearTopicCategoryCache()
	fileusageservice.ReplaceTopic(topic.Id, topic.UserId, post.Content)
	publishTopicReviewResult(&topic, &post, firstPublication)
	logContentReview(req, "topic", topic.Id)
	return component.SuccessResponse(true)
}

type ReviewPostsReq struct {
	Page             int    `json:"page"`
	PageSize         int    `json:"pageSize"`
	ModerationStatus string `json:"moderationStatus"`
	Search           string `json:"search"`
}

type ReviewPostVo struct {
	Id                uint64    `json:"id"`
	TopicId           uint64    `json:"topicId"`
	TopicTitle        string    `json:"topicTitle"`
	UserId            uint64    `json:"userId"`
	PostNo            uint64    `json:"postNo"`
	Content           string    `json:"content"`
	ModerationStatus  string    `json:"moderationStatus"`
	ModerationReason  string    `json:"moderationReason"`
	ModerationVersion uint64    `json:"moderationVersion"`
	ProcessStatus     int8      `json:"processStatus"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func ReviewPostsList(req component.BetterRequest[ReviewPostsReq]) component.Response {
	page, size := max(req.Params.Page, 1), min(max(req.Params.PageSize, 1), 50)
	var rows []ReviewPostVo
	q := dbconnect.Connect().Model(&posts.Entity{}).
		Select("posts.id, posts.topic_id, topics.title AS topic_title, posts.user_id, posts.post_no, posts.content, posts.moderation_status, posts.moderation_reason, posts.moderation_version, posts.process_status, posts.updated_at").
		Joins("JOIN topics ON topics.id = posts.topic_id AND topics.deleted_at IS NULL").Where("posts.post_no > 1")
	if req.Params.ModerationStatus != "" {
		q = q.Where("posts.moderation_status = ?", req.Params.ModerationStatus)
	}
	if req.Params.Search != "" {
		q = q.Where("topics.title LIKE ?", "%"+req.Params.Search+"%")
	}
	if err := q.Order("posts.updated_at ASC").Order("posts.id ASC").Offset((page - 1) * size).Limit(size + 1).Scan(&rows).Error; err != nil {
		return component.FailResponseError(err)
	}
	hasNext := len(rows) > size
	if hasNext {
		rows = rows[:size]
	}
	return component.SuccessResponse(component.Page[ReviewPostVo]{List: rows, Page: page, Size: size, HasNext: hasNext})
}

func ReviewAdminPost(req component.BetterRequest[ReviewContentReq]) component.Response {
	post := posts.Get(req.Params.Id)
	if post.Id == 0 || post.PostNo <= 1 {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	if post.ModerationVersion != req.Params.Version {
		return component.FailResponseError(fmt.Errorf("内容已更新，请刷新后重新审核"))
	}
	topic := topics.Get(post.TopicId)
	if topic.Id == 0 {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	wasPublished := post.WasPublished()
	wasVisible := post.ProcessStatus == 0
	if req.Params.Action == "recheck" {
		if !sensitivewordservice.Enabled() {
			return component.FailResponseError(fmt.Errorf("请先启用敏感词检测"))
		}
		contentmoderationservice.ReviewPost(&post)
	} else {
		post.ModerationVersion++
		post.ModerationStatus, post.ProcessStatus = "approved", 0
		if req.Params.Action == "reject" {
			post.ModerationStatus, post.ProcessStatus = "denied", 1
		}
		now := time.Now()
		post.ModerationReason, post.ModeratedAt = req.Params.Reason, &now
	}
	if post.PublishedAt == nil && (wasPublished || post.ProcessStatus == 0) {
		now := time.Now()
		post.PublishedAt = &now
	}
	if err := posts.SaveReviewed(&post, req.Params.Version); err != nil {
		return component.FailResponseError(err)
	}
	if wasVisible != (post.ProcessStatus == 0) {
		postservice.SyncTopicPostStats(topic, post, post.ProcessStatus != 0)
	}
	hotdataserve.ClearTopicListCache()
	fileusageservice.ReplacePost(post.Id, post.UserId, post.Content)
	if !wasPublished && post.ProcessStatus == 0 {
		publishVisiblePost(topic, post)
	}
	logContentReview(req, "post", topic.Id)
	return component.SuccessResponse(true)
}

func logContentReview(req component.BetterRequest[ReviewContentReq], kind string, id uint64) {
	optlogger.UserOptCode(req.UserId, optlogger.EditTopic, id, "admin.opt.content.reviewed", optlogger.MessageParams{
		"type": kind, "subjectId": req.Params.Id, "action": req.Params.Action, "reason": req.Params.Reason, "version": req.Params.Version,
	})
}

func publishTopicReviewResult(topic *topics.Entity, post *posts.Entity, firstPublication bool) {
	userservice.InvalidateUserPublicProfileCache(topic.UserId)
	if firstPublication && topic.Status == 1 && topic.ProcessStatus == 0 {
		userStatistics.WriteTopic(topic.UserId)
		eventbus.Publish(context.Background(), &eventhandlers.TopicPublishedEvent{Topic: topic, FirstPost: post})
	} else {
		eventbus.Publish(context.Background(), &eventhandlers.TopicUpdatedEvent{Topic: topic, FirstPost: post})
	}
}

func publishVisiblePost(topic topics.Entity, post posts.Entity) {
	if topic.Status != 1 || topic.ProcessStatus != 0 || post.ProcessStatus != 0 {
		return
	}
	userStatistics.WriteComment(post.UserId)
	userservice.InvalidateUserPublicProfileCache(post.UserId)
	hotdataserve.ClearTopicListCache()
	var parentUserID uint64
	if post.ReplyToPostId != 0 {
		parentUserID = posts.Get(post.ReplyToPostId).UserId
	}
	eventbus.Publish(context.Background(), &eventhandlers.CommentCreatedEvent{
		TopicId: topic.Id, PostId: post.Id, PostNo: post.PostNo, UserId: post.UserId,
		Content: post.Content, TopicAuthorId: topic.UserId, ReplyToPostId: post.ReplyToPostId, ReplyToPostAuthorId: parentUserID,
	})
}
