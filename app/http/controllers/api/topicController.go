package api

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/postwriteservice"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/leancodebox/GooseForum/app/service/topicactionservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
	"github.com/leancodebox/GooseForum/app/service/topicwriteservice"
	"github.com/leancodebox/GooseForum/app/service/userfollowservice"
)

func GetSiteStatistics() component.Response {
	return component.SuccessResponse(hotdataserve.GetSiteStatisticsData())
}

type WriteTopicReq struct {
	ReturnReview bool     `json:"returnReview"`
	TopicId      uint64   `json:"topicId"`
	Content      string   `json:"content" validate:"required"`
	Title        string   `json:"title" validate:"required"`
	CategoryId   []uint64 `json:"categoryId" validate:"min=1,max=3"`
	TopicStatus  int8     `json:"topicStatus" validate:"oneof=0 1"`
}

// WriteTopic creates or updates a topic and its first post.
func WriteTopic(req component.BetterRequest[WriteTopicReq]) component.Response {
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	userEntity, err := req.GetUser()
	if err != nil || userEntity.Id == 0 {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}

	// 统一权限检查
	if _, err := component.CheckUserPermission(&userEntity, component.PermissionActionPost); err != nil {
		return component.FailResponseError(err)
	}
	isNew := req.Params.TopicId == 0

	if response, invalid := validateTextLength(req.Params.Title, postingConfig.TextControl.MinTitleLength, postingConfig.TextControl.MaxTitleLength, component.MessageTopicTitleTooShort, component.MessageTopicTitleTooLong); invalid {
		return response
	}
	if response, invalid := validateTextLength(req.Params.Content, postingConfig.TextControl.MinPostLength, postingConfig.TextControl.MaxPostLength, component.MessageTopicContentTooShort, component.MessageTopicContentTooLong); invalid {
		return response
	}

	// 检查新用户冷却时间
	if isNew {
		if response, coolingDown := newUserCooldownResponse(userEntity.CreatedAt, postingConfig.TextControl.NewUserPostCooldownMinutes, component.MessageTopicPostCooldown); coolingDown {
			return response
		}
	}

	result, err := topicwriteservice.Write(topicwriteservice.WriteInput{
		UserID: req.UserId, TopicID: req.Params.TopicId, Title: req.Params.Title,
		Content: req.Params.Content, CategoryIDs: req.Params.CategoryId, Status: req.Params.TopicStatus,
		DailyLimit: postingConfig.TextControl.MaxDailyTopicsPerUser,
	})
	if err != nil {
		if errors.Is(err, topicwriteservice.ErrTopicNotFound) {
			return component.FailResponseCode(component.MessageTopicNotFound, nil)
		}
		if errors.Is(err, topicwriteservice.ErrOwnerMismatch) {
			return component.FailResponseCode(component.MessageTopicOwnerMismatch, nil)
		}
		if errors.Is(err, topicwriteservice.ErrDailyLimit) {
			return component.FailResponseCode(component.MessageTopicDailyLimit, nil)
		}
		if errors.Is(err, accesscontrol.ErrRestrictedCategorySingle) {
			return component.FailResponseCode(component.MessageTopicRestrictedSingle, nil)
		}
		if errors.Is(err, topicwriteservice.ErrPermissionDenied) {
			return component.FailResponseCode(component.MessagePermissionDenied, nil)
		}
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	if req.Params.ReturnReview {
		return component.SuccessResponse(map[string]any{"id": result.ID, "moderationStatus": result.ModerationStatus, "topicStatus": result.TopicStatus})
	}
	return component.SuccessResponse(result.ID)
}

func validateTextLength(value string, minLength, maxLength int, tooShort, tooLong component.MessageCode) (component.Response, bool) {
	if len(value) < minLength {
		return component.FailResponseCode(tooShort, component.MessageParams{"minLength": minLength}), true
	}
	if len(value) > maxLength {
		return component.FailResponseCode(tooLong, component.MessageParams{"maxLength": maxLength}), true
	}
	return component.Response{}, false
}

func newUserCooldownResponse(createdAt time.Time, minutes int, code component.MessageCode) (component.Response, bool) {
	if minutes <= 0 {
		return component.Response{}, false
	}
	availableAt := createdAt.Add(time.Duration(minutes) * time.Minute)
	if !time.Now().Before(availableAt) {
		return component.Response{}, false
	}
	return component.FailResponseCode(code, component.MessageParams{
		"minutes": minutes, "availableAt": availableAt.Format("2006-01-02 15:04:05"),
	}), true
}

type TopicStatusReq struct {
	TopicId     uint64 `json:"topicId" validate:"required"`
	TopicStatus int8   `json:"topicStatus" validate:"oneof=0 1"`
}

func UpdateTopicStatus(req component.BetterRequest[TopicStatusReq]) component.Response {
	if err := topicwriteservice.UpdateStatus(req.UserId, req.Params.TopicId, req.Params.TopicStatus); err != nil {
		if errors.Is(err, topicwriteservice.ErrTopicNotFound) {
			return component.FailResponseCode(component.MessageTopicNotFound, nil)
		}
		if errors.Is(err, topicwriteservice.ErrOwnerMismatch) {
			return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
		}
		if errors.Is(err, accesscontrol.ErrRestrictedCategorySingle) {
			return component.FailResponseCode(component.MessageTopicRestrictedSingle, nil)
		}
		if errors.Is(err, topicwriteservice.ErrPermissionDenied) {
			return component.FailResponseCode(component.MessagePermissionDenied, nil)
		}
		return component.FailResponseCode(component.MessageTopicSaveFailed, nil)
	}
	return component.SuccessResponse(true)
}

type CreatePostReq struct {
	TopicId       uint64 `json:"topicId"`
	Content       string `json:"content"`
	ReplyToPostId uint64 `json:"replyToPostId"`
}

func CreatePost(req component.BetterRequest[CreatePostReq]) component.Response {
	// 获取发布设置
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	userEntity, err := req.GetUser()
	if err != nil || userEntity.Id == 0 {
		return component.FailResponseCode(component.MessageUserFetchFailed, nil)
	}

	// 统一权限检查
	if _, err := component.CheckUserPermission(&userEntity, component.PermissionActionComment); err != nil {
		return component.FailResponseError(err)
	}

	content := strings.TrimSpace(req.Params.Content)
	if response, invalid := validateTextLength(content, postingConfig.TextControl.MinPostLength, postingConfig.TextControl.MaxPostLength, component.MessageCommentContentTooShort, component.MessageCommentContentTooLong); invalid {
		return response
	}

	// 评论也受发帖冷却限制
	if response, coolingDown := newUserCooldownResponse(userEntity.CreatedAt, postingConfig.TextControl.NewUserPostCooldownMinutes, component.MessageCommentPostCooldown); coolingDown {
		return response
	}

	postEntity, err := postwriteservice.Create(postwriteservice.CreateInput{
		UserID: req.UserId, TopicID: req.Params.TopicId, Content: content, ReplyToPostID: req.Params.ReplyToPostId,
	})
	if err != nil {
		if errors.Is(err, postwriteservice.ErrTopicUnavailable) {
			return component.FailResponseCode(component.MessageTopicNotFound, nil)
		}
		if errors.Is(err, postwriteservice.ErrParentPostMissing) {
			return component.FailResponseCode(component.MessageCommentParentPostMissing, nil)
		}
		return component.FailResponseCode(
			component.MessageCommentCreateFailed,

			component.MessageParams{"error": err.Error()})

	}
	return component.SuccessResponse(map[string]any{
		"processStatus":    postEntity.ProcessStatus,
		"moderationStatus": postEntity.ModerationStatus,
		"id":               postEntity.Id,
		"postNo":           postEntity.PostNo,
		"renderedContent":  postEntity.RenderedHTML,
	})
}

type DeletePostReq struct {
	PostId uint64 `json:"postId"`
}

type UpdatePostReq struct {
	PostId  uint64 `json:"postId"`
	Content string `json:"content"`
}

func UpdatePost(req component.BetterRequest[UpdatePostReq]) component.Response {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()
	content := strings.TrimSpace(req.Params.Content)
	if response, invalid := validateTextLength(content, postingConfig.TextControl.MinPostLength, postingConfig.TextControl.MaxPostLength, component.MessageCommentContentTooShort, component.MessageCommentContentTooLong); invalid {
		return response
	}

	postEntity, err := postwriteservice.Update(postwriteservice.UpdateInput{UserID: req.UserId, PostID: req.Params.PostId, Content: content})
	if err != nil {
		if errors.Is(err, postwriteservice.ErrPostNotFound) || errors.Is(err, postwriteservice.ErrTopicUnavailable) {
			return component.FailResponseCode(component.MessagePostNotFound, nil)
		}
		if errors.Is(err, postwriteservice.ErrOwnerMismatch) {
			return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
		}
		return component.FailResponseCode(
			component.MessagePostUpdateFailed,

			component.MessageParams{"error": err.Error()})

	}
	return component.SuccessResponse(map[string]any{
		"id":               postEntity.Id,
		"postNo":           postEntity.PostNo,
		"content":          postEntity.Content,
		"renderedContent":  postEntity.RenderedHTML,
		"updatedAt":        postEntity.UpdatedAt.Format(time.DateTime),
		"moderationStatus": postEntity.ModerationStatus,
		"processStatus":    postEntity.ProcessStatus,
	})
}

func DeletePost(req component.BetterRequest[DeletePostReq]) component.Response {
	postEntity := posts.Get(req.Params.PostId)
	if postEntity.Id == 0 {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	topicEntity := topics.Get(postEntity.TopicId)
	if err := authorizePublishedTopic(req.UserId, topicEntity, accesscontrol.CapabilityRead); err != nil {
		return component.FailResponseCode(component.MessagePostNotFound, nil)
	}
	if postEntity.UserId != req.UserId {
		return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
	}
	if postEntity.PostNo <= 1 {
		if topicEntity.UserId != req.UserId || topicEntity.FirstPostId != postEntity.Id {
			return component.FailResponseCode(component.MessageTopicOperationDenied, nil)
		}
		topicEntity.ProcessStatus = 1
		if _, err := searchservice.BuildSingleTopicSearchDocument(&topicEntity, &postEntity); err != nil {
			slog.Error("failed to remove self-deleted topic from search", "topicId", topicEntity.Id, "err", err)
		}
		if err := topicservice.DeleteTopic(&topicEntity); err != nil {
			return component.FailResponseCode(component.MessageOperationFailed, nil)
		}
		hotdataserve.ClearTopicWriteCaches(true)
		return component.SuccessResponse(true)
	}
	posts.DeleteEntity(&postEntity)
	if topicEntity.Id > 0 && postEntity.ProcessStatus == 0 {
		postservice.SyncTopicPostStats(topicEntity, postEntity, true)
		hotdataserve.ClearTopicWriteCaches(false)
	}
	return component.SuccessResponse(true)
}

type LikeTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 点赞，2 取消
}

func LikeTopic(req component.BetterRequest[LikeTopicReq]) component.Response {
	if err := topicactionservice.SetLiked(req.UserId, req.Params.TopicId, req.Params.Action == 1); err != nil {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	return component.SuccessResponse(true)
}

type BookmarkTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 收藏，2 取消
}

func BookmarkTopic(req component.BetterRequest[BookmarkTopicReq]) component.Response {
	if err := topicactionservice.SetBookmarked(req.UserId, req.Params.TopicId, req.Params.Action == 1); err != nil {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	return component.SuccessResponse(true)
}

type WatchTopicReq struct {
	TopicId uint64 `json:"topicId"`
	Action  int    `json:"action" validate:"min=1,max=2"` // 1 关注，2 取消
}

func WatchTopic(req component.BetterRequest[WatchTopicReq]) component.Response {
	if err := topicactionservice.SetWatched(req.UserId, req.Params.TopicId, req.Params.Action == 1); err != nil {
		return component.FailResponseCode(component.MessageTopicNotFound, nil)
	}
	return component.SuccessResponse(true)
}

type FollowUserReq struct {
	Id     uint64 `json:"id"`
	Action int    `json:"action" validate:"min=1,max=2"` // 1 关注，2 取消
}

func FollowUser(req component.BetterRequest[FollowUserReq]) component.Response {
	if err := userfollowservice.SetFollowed(req.UserId, req.Params.Id, req.Params.Action == 1); err != nil {
		return component.FailResponseCode(component.MessageUserNotFound, nil)
	}
	return component.SuccessResponse(true)
}
