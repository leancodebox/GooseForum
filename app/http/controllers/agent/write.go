package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/agenttokens"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/mentionservice"
	"github.com/leancodebox/GooseForum/app/service/postingpolicy"
	"github.com/leancodebox/GooseForum/app/service/postwriteservice"
	"github.com/leancodebox/GooseForum/app/service/topicwriteservice"
)

type createTopicRequest struct {
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	CategoryIDs     []string `json:"categoryIds"`
	ClientRequestID string   `json:"clientRequestId"`
}
type createPostRequest struct {
	Content         string `json:"content"`
	ReplyToPostID   string `json:"replyToPostId,omitempty"`
	ClientRequestID string `json:"clientRequestId"`
}

func writeAllowed(c *gin.Context) bool {
	a := current(c)
	if _, err := component.CheckUserPermission(&a.User, component.PermissionActionWrite); err != nil {
		fail(c, 403, "account_not_writable", "Account is restricted or requires email verification")
		return false
	}
	return true
}

func requestKey(c *gin.Context, value string) bool {
	u, err := uuid.Parse(value)
	if err != nil || u == uuid.Nil || u.String() != value {
		fail(c, 400, "invalid_request", "clientRequestId must be a canonical nonzero UUID")
		return false
	}
	return true
}

func fingerprint(operation string, body any) string {
	data, _ := json.Marshal([]any{operation, body})
	return agenttokens.Digest(string(data))
}

func validateContent(c *gin.Context, content string, title *string) bool {
	if strings.TrimSpace(content) == "" || title != nil && strings.TrimSpace(*title) == "" {
		fail(c, 422, "empty_content", "Title and content must not be empty")
		return false
	}
	config := hotdataserve.GetPostingSettingsConfigCache().TextControl
	if !postingpolicy.ValidLength(content, config.MinPostLength, config.MaxPostLength) {
		fail(c, 422, "invalid_content_length", "Content length is outside the site's UTF-8 byte limits")
		return false
	}
	if title != nil && !postingpolicy.ValidLength(*title, config.MinTitleLength, config.MaxTitleLength) {
		fail(c, 422, "invalid_title_length", "Title length is outside the site's UTF-8 byte limits")
		return false
	}
	available := postingpolicy.AvailableAt(current(c).User.CreatedAt, config.NewUserPostCooldownMinutes)
	if available.After(time.Now()) {
		seconds := int(math.Ceil(time.Until(available).Seconds()))
		c.Header("Retry-After", strconv.Itoa(seconds))
		fail(c, 429, "posting_cooldown", "Posting is available at "+available.UTC().Format(time.RFC3339))
		return false
	}
	return true
}

func (h *Handler) existingSubmission(c *gin.Context, key, hash string) bool {
	row, err := posts.GetAgentSubmission(current(c).Source, current(c).User.Id, 0, key)
	if notFound(err) {
		return false
	}
	if err != nil {
		fail(c, 500, "internal_error", "Unable to check submission")
		return true
	}
	if row.RequestFingerprint != hash {
		fail(c, 409, "client_request_conflict", "clientRequestId was already used with different parameters")
		return true
	}
	h.submissionResult(c, row, 200, true)
	return true
}

func (h *Handler) createTopic(c *gin.Context) {
	var req createTopicRequest
	if !decode(c, &req) || !requestKey(c, req.ClientRequestID) || !writeAllowed(c) {
		return
	}
	if len(req.CategoryIDs) < 1 || len(req.CategoryIDs) > 3 {
		fail(c, 422, "invalid_categories", "Choose between one and three categories")
		return
	}
	ids := make([]uint64, 0, len(req.CategoryIDs))
	seen := map[uint64]bool{}
	for _, value := range req.CategoryIDs {
		categoryID, ok := id(value)
		if !ok || seen[categoryID] {
			fail(c, 422, "invalid_categories", "Category IDs must be unique positive decimal strings")
			return
		}
		if !current(c).Access.CanCreateCategory(categoryID) {
			fail(c, 403, "permission_denied", "Publishing is not allowed in the selected category")
			return
		}
		seen[categoryID] = true
		ids = append(ids, categoryID)
	}
	hash := fingerprint("topics:create", req)
	if h.existingSubmission(c, req.ClientRequestID, hash) {
		return
	}
	if !validateContent(c, req.Content, &req.Title) {
		return
	}
	source := current(c).Source
	result, err := topicwriteservice.Write(topicwriteservice.WriteInput{AgentSource: &source, ClientRequestID: &req.ClientRequestID, RequestFingerprint: hash, SourceVersion: 1, UserID: current(c).User.Id, Title: req.Title, Content: req.Content, CategoryIDs: ids, Status: 1, DailyLimit: hotdataserve.GetPostingSettingsConfigCache().TextControl.MaxDailyTopicsPerUser})
	if err != nil {
		if h.existingSubmission(c, req.ClientRequestID, hash) {
			return
		}
		writeError(c, err)
		return
	}
	row, err := posts.GetTopicSubmission(result.ID)
	if err != nil {
		fail(c, 500, "submission_lookup_failed", "Query the submission using the same clientRequestId")
		return
	}
	h.submissionResult(c, row, 201, false)
}

func (h *Handler) createPost(c *gin.Context) {
	var req createPostRequest
	if !decode(c, &req) || !requestKey(c, req.ClientRequestID) || !writeAllowed(c) {
		return
	}
	topic, ok := h.loadTopic(c, c.Param("topicId"))
	if !ok {
		return
	}
	if !current(c).Access.CanReplyCategory(topic.MainCategoryId) {
		fail(c, 403, "permission_denied", "Replying is not allowed")
		return
	}
	parentID := uint64(0)
	if req.ReplyToPostID != "" {
		parentID, ok = id(req.ReplyToPostID)
		if !ok {
			fail(c, 400, "invalid_request", "Invalid replyToPostId")
			return
		}
		found, err := posts.AgentPostExists(parentID, topic.Id)
		if err != nil {
			fail(c, 500, "internal_error", "Unable to read reply target")
			return
		}
		if !found {
			fail(c, 404, "not_found", "Reply target is unavailable")
			return
		}
	}
	hash := fingerprint("posts:create:"+fmt.Sprint(topic.Id), req)
	if h.existingSubmission(c, req.ClientRequestID, hash) {
		return
	}
	content := strings.TrimSpace(req.Content)
	if !validateContent(c, content, nil) {
		return
	}
	source := current(c).Source
	row, err := postwriteservice.Create(postwriteservice.CreateInput{AgentSource: &source, ClientRequestID: &req.ClientRequestID, RequestFingerprint: hash, SourceVersion: 1, UserID: current(c).User.Id, TopicID: topic.Id, Content: content, ReplyToPostID: parentID})
	if err != nil {
		if h.existingSubmission(c, req.ClientRequestID, hash) {
			return
		}
		writeError(c, err)
		return
	}
	h.submissionResult(c, row, 201, false)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, topicwriteservice.ErrDailyLimit):
		fail(c, 429, "daily_topic_limit", "Daily topic limit reached")
	case errors.Is(err, mentionservice.ErrInvalid):
		fail(c, 422, "invalid_mention", "Invalid mention")
	case errors.Is(err, accesscontrol.ErrRestrictedCategorySingle), errors.Is(err, accesscontrol.ErrCategoryRequired), errors.Is(err, accesscontrol.ErrTooManyCategories):
		fail(c, 422, "invalid_categories", "Invalid category selection")
	case errors.Is(err, topicwriteservice.ErrPermissionDenied), errors.Is(err, accesscontrol.ErrCategoryPermissionDenied):
		fail(c, 403, "permission_denied", "Category operation is not allowed")
	case errors.Is(err, postwriteservice.ErrTopicUnavailable), errors.Is(err, postwriteservice.ErrParentPostMissing):
		fail(c, 404, "not_found", "Topic or reply target is unavailable")
	default:
		fail(c, 500, "internal_error", "Write failed; query the submission before retrying with the same clientRequestId")
	}
}

func (h *Handler) submissionResult(c *gin.Context, row posts.Entity, status int, reused bool) {
	topic, err := topics.GetSubmissionTopic(row.TopicId)
	if err != nil {
		if notFound(err) {
			fail(c, 410, "submission_gone", "Submission is no longer available")
		} else {
			fail(c, 500, "internal_error", "Unable to read submission")
		}
		return
	}
	if row.DeletedAt.Valid || topic.DeletedAt.Valid {
		fail(c, 410, "submission_gone", "Submission was deleted")
		return
	}
	if !current(c).Access.CanReadCategory(topic.MainCategoryId) {
		fail(c, 404, "not_found", "Submission is unavailable")
		return
	}
	visible := topic.Status == 1 && topic.ProcessStatus == 0 && row.ProcessStatus == 0
	if visible && row.PostNo > 1 {
		found, err := posts.AgentPostExists(topic.FirstPostId, 0)
		if err != nil {
			fail(c, 500, "internal_error", "Unable to read publication state")
			return
		}
		visible = found
	}
	moderation := row.ModerationStatus
	if row.PostNo == 1 {
		moderation = topic.ModerationStatus
	}
	if moderation == "" {
		moderation = "none"
	}
	success(c, status, gin.H{"submissionId": fmt.Sprint(row.Id), "clientRequestId": row.ClientRequestID, "topicId": fmt.Sprint(row.TopicId), "postId": fmt.Sprint(row.Id), "postNo": row.PostNo, "url": postURL(row.TopicId, row.PostNo), "visible": visible, "moderationStatus": moderation, "reused": reused})
}

func (h *Handler) submission(c *gin.Context) {
	value, ok := id(c.Param("submissionId"))
	if !ok {
		fail(c, 400, "invalid_request", "Invalid submission ID")
		return
	}
	row, err := posts.GetAgentSubmission(current(c).Source, current(c).User.Id, value, "")
	if err != nil {
		if notFound(err) {
			fail(c, 404, "not_found", "Submission is unavailable")
		} else {
			fail(c, 500, "internal_error", "Unable to read submission")
		}
		return
	}
	h.submissionResult(c, row, 200, false)
}

func (h *Handler) submissions(c *gin.Context) {
	key := c.Query("clientRequestId")
	if key != "" {
		if !requestKey(c, key) {
			return
		}
		row, err := posts.GetAgentSubmission(current(c).Source, current(c).User.Id, 0, key)
		if err != nil {
			if notFound(err) {
				fail(c, 404, "not_found", "No committed submission found")
			} else {
				fail(c, 500, "internal_error", "Unable to read submission")
			}
			return
		}
		h.submissionResult(c, row, 200, false)
		return
	}
	n, ok := limit(c)
	if !ok {
		return
	}
	before, ok := readCursor(c, "submissions", current(c).Source)
	if !ok {
		return
	}
	a := current(c)
	rows, err := posts.ListAgentSubmissions(a.Source, a.User.Id, before, a.Access.ReadableCategoryIDs(), a.Access.HasGlobalManage(), n+1)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read submissions")
		return
	}
	more := len(rows) > n
	if more {
		rows = rows[:n]
	}
	data := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		data = append(data, gin.H{"submissionId": fmt.Sprint(row.Id), "topicId": fmt.Sprint(row.TopicId), "clientRequestId": row.ClientRequestID})
	}
	next := ""
	if more {
		next = cursorToken("submissions", current(c).Source, rows[len(rows)-1].Id)
	}
	page(c, data, gin.H{"hasMore": more, "nextCursor": next})
}
