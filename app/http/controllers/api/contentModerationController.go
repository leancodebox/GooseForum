package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/mentionservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
)

type ReviewContentReq struct {
	Id      uint64 `json:"id" validate:"required"`
	Version uint64 `json:"version"`
	Action  string `json:"action" validate:"oneof=approve reject recheck"`
	Reason  string `json:"reason" validate:"required,max=300"`
}

type ReviewBatchItem struct {
	Id      uint64 `json:"id" validate:"required"`
	Version uint64 `json:"version"`
}

type ReviewBatchReq struct {
	Items  []ReviewBatchItem `json:"items" validate:"required,min=1,max=50"`
	Action string            `json:"action" validate:"oneof=approve reject recheck"`
	Reason string            `json:"reason" validate:"required,max=300"`
}

type ReviewBatchItemResult struct {
	Id          uint64                `json:"id"`
	Success     bool                  `json:"success"`
	MessageCode component.MessageCode `json:"messageCode,omitempty"`
}

type ReviewBatchResult struct {
	Items []ReviewBatchItemResult `json:"items"`
}

func ReviewAdminTopic(req component.BetterRequest[ReviewContentReq]) component.Response {
	if err := reviewAdminTopic(req.UserId, req.Params); err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(true)
}

func ReviewAdminTopicBatch(req component.BetterRequest[ReviewBatchReq]) component.Response {
	return reviewAdminBatch(req.UserId, req.Params, "topic")
}

func reviewAdminTopic(userID uint64, params ReviewContentReq) error {
	topic := topics.Get(params.Id)
	if topic.Id == 0 {
		return component.NewMessageError(component.MessageTopicNotFound, "主题不存在", nil)
	}
	if topic.ModerationVersion != params.Version {
		return fmt.Errorf("内容已更新，请刷新后重新审核")
	}
	// Never publish a private draft through an approval action.
	if topic.Status == 0 && (topic.ModerationStatus == "none" || topic.ModerationStatus == "approved" || topic.ModerationStatus == "") {
		return fmt.Errorf("草稿不能通过审核操作发布")
	}
	post := posts.Get(topic.FirstPostId)
	if post.Id == 0 {
		return component.NewMessageError(component.MessagePostNotFound, "帖子不存在", nil)
	}
	legacyHidden := post.ModerationStatus == "rejected" || post.ModerationStatus == "pending"
	firstPublication := topic.PublishedAt == nil
	wasCounted := topic.IsPublished()
	now := time.Now()
	if params.Action == "recheck" {
		if !sensitivewordservice.Enabled() {
			return fmt.Errorf("请先启用敏感词检测")
		}
		topic.Status = 1
		contentmoderationservice.PrepareTopic(&topic, &post)
	} else {
		topic.ModerationVersion++
		post.ModerationVersion++
		topic.ModerationStatus, topic.Status = "approved", 1
		if params.Action == "reject" {
			topic.ModerationStatus, topic.Status = "rejected", 0
		}
		topic.ModerationReason, topic.ModeratedAt = params.Reason, &now
		post.ModerationStatus, post.ModerationReason, post.ModeratedAt = topic.ModerationStatus, params.Reason, &now
		// Legacy automatic review also hid the first post.
		if legacyHidden {
			post.ProcessStatus = 0
		}
	}
	if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds, ExpectedVersion: &params.Version, ReviewOnly: true}); err != nil {
		return err
	}
	hotdataserve.ClearTopicWriteCaches(wasCounted != topic.IsPublished())
	fileusageservice.ReplaceTopic(topic.Id, topic.UserId, post.Content, post.SourceVersion)
	eventhandlers.PublishTopicReviewResult(&topic, &post, firstPublication)
	logContentReview(userID, params, "topic", topic.Id)
	enqueueContentReview(true, topic.Id, topic.ModerationVersion, topic.ModerationStatus)
	return nil
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
	if req.Params.ModerationStatus == "rejected" {
		q = q.Where("posts.moderation_status IN ?", []string{"rejected", "denied"})
	} else if req.Params.ModerationStatus != "" {
		q = q.Where("posts.moderation_status = ?", req.Params.ModerationStatus)
	}
	if req.Params.Search != "" {
		q = q.Where("topics.title LIKE ?", "%"+req.Params.Search+"%")
	}
	if err := q.Order("posts.updated_at ASC").Order("posts.id ASC").Offset((page - 1) * size).Limit(size + 1).Scan(&rows).Error; err != nil {
		return component.FailResponseError(err)
	}
	rows, hasNext := pageutil.TrimPage(rows, size)
	return component.SuccessResponse(component.Page[ReviewPostVo]{List: rows, Page: page, Size: size, HasNext: hasNext})
}

func ReviewAdminPost(req component.BetterRequest[ReviewContentReq]) component.Response {
	if err := reviewAdminPost(req.UserId, req.Params); err != nil {
		return component.FailResponseError(err)
	}
	return component.SuccessResponse(true)
}

func ReviewAdminPostBatch(req component.BetterRequest[ReviewBatchReq]) component.Response {
	return reviewAdminBatch(req.UserId, req.Params, "post")
}

func reviewAdminBatch(userID uint64, params ReviewBatchReq, kind string) component.Response {
	if len(params.Items) == 0 || len(params.Items) > 50 || (params.Action != "approve" && params.Action != "reject" && params.Action != "recheck") || len([]rune(params.Reason)) == 0 || len([]rune(params.Reason)) > 300 {
		return component.FailResponseError(fmt.Errorf("批量审核参数无效"))
	}
	seen := make(map[uint64]struct{}, len(params.Items))
	results := make([]ReviewBatchItemResult, 0, len(params.Items))
	for _, item := range params.Items {
		if item.Id == 0 {
			return component.FailResponseError(fmt.Errorf("批量审核参数无效"))
		}
		if _, exists := seen[item.Id]; exists {
			return component.FailResponseError(fmt.Errorf("批量审核项目不能重复"))
		}
		seen[item.Id] = struct{}{}
		input := ReviewContentReq{Id: item.Id, Version: item.Version, Action: params.Action, Reason: params.Reason}
		var err error
		if kind == "topic" {
			err = reviewAdminTopic(userID, input)
		} else {
			err = reviewAdminPost(userID, input)
		}
		result := ReviewBatchItemResult{Id: item.Id, Success: err == nil}
		if err != nil {
			if messageErr, ok := errors.AsType[component.MessageError](err); ok {
				result.MessageCode = messageErr.Code
			} else {
				result.MessageCode = component.MessageOperationFailed
			}
		}
		results = append(results, result)
	}
	return component.SuccessResponse(ReviewBatchResult{Items: results})
}

func reviewAdminPost(userID uint64, params ReviewContentReq) error {
	post := posts.Get(params.Id)
	if post.Id == 0 || post.PostNo <= 1 {
		return component.NewMessageError(component.MessagePostNotFound, "帖子不存在", nil)
	}
	if post.ModerationVersion != params.Version {
		return fmt.Errorf("内容已更新，请刷新后重新审核")
	}
	topic := topics.Get(post.TopicId)
	if topic.Id == 0 {
		return component.NewMessageError(component.MessageTopicNotFound, "主题不存在", nil)
	}
	wasPublished := post.WasPublished()
	wasVisible := post.ProcessStatus == 0
	if params.Action == "recheck" {
		if !sensitivewordservice.Enabled() {
			return fmt.Errorf("请先启用敏感词检测")
		}
		contentmoderationservice.PreparePost(&post)
	} else {
		post.ModerationVersion++
		post.ModerationStatus, post.ProcessStatus = "approved", 0
		if params.Action == "reject" {
			post.ModerationStatus, post.ProcessStatus = "rejected", 1
		}
		now := time.Now()
		post.ModerationReason, post.ModeratedAt = params.Reason, &now
	}
	if post.PublishedAt == nil && (wasPublished || post.ProcessStatus == 0) {
		now := time.Now()
		post.PublishedAt = &now
	}
	if err := posts.SaveReviewed(&post, params.Version); err != nil {
		return err
	}
	if wasVisible != (post.ProcessStatus == 0) {
		postservice.SyncTopicPostStats(topic, post, post.ProcessStatus != 0)
	}
	fileusageservice.ReplacePost(post.Id, post.UserId, post.Content, post.SourceVersion)
	if eventhandlers.ShouldPublishVisiblePost(topic, post, wasPublished) {
		eventhandlers.PublishVisiblePost(topic, post)
	} else {
		hotdataserve.ClearTopicListCache()
		mentionservice.Notify(topic, post)
		eventbus.Publish(context.Background(), &eventhandlers.PostUpdatedEvent{TopicID: topic.Id})
	}
	logContentReview(userID, params, "post", topic.Id)
	enqueueContentReview(false, post.Id, post.ModerationVersion, post.ModerationStatus)
	return nil
}

func logContentReview(userID uint64, params ReviewContentReq, kind string, id uint64) {
	optlogger.UserOptCode(userID, optlogger.EditTopic, id, "admin.opt.content.reviewed", optlogger.MessageParams{
		"type": kind, "subjectId": params.Id, "action": params.Action, "reason": params.Reason, "version": params.Version,
	})
}

func enqueueContentReview(topic bool, id, version uint64, status string) {
	if status != "pending" {
		return
	}
	eventbus.Publish(context.Background(), &eventhandlers.ContentReviewRequestedEvent{
		Topic: topic, ID: id, Version: version,
	})
}
