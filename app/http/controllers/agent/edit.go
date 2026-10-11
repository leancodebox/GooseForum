package agent

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/postwriteservice"
	"github.com/leancodebox/GooseForum/app/service/topicwriteservice"
)

type editTopicRequest struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Version *uint64 `json:"version"`
}

type editPostRequest struct {
	Content string  `json:"content"`
	Version *uint64 `json:"version"`
}

func editVersion(c *gin.Context, version *uint64) bool {
	if version == nil {
		fail(c, 400, "invalid_request", "version from the latest GET response is required")
		return false
	}
	return true
}

func (h *Handler) editTopic(c *gin.Context) {
	var req editTopicRequest
	if !decode(c, &req) || !editVersion(c, req.Version) || !writeAllowed(c) {
		return
	}
	row, ok := h.loadTopic(c, c.Param("topicId"))
	if !ok {
		return
	}
	if row.UserId != current(c).User.Id {
		fail(c, 403, "permission_denied", "Only the author can edit a topic")
		return
	}
	if !validateWriteContent(c, req.Content, &req.Title, false) {
		return
	}
	// Editing preserves the categories and publication state of the existing topic.
	topic := topics.Get(row.Id)
	result, err := topicwriteservice.Write(topicwriteservice.WriteInput{
		TopicID: row.Id, UserID: current(c).User.Id, ExpectedVersion: req.Version,
		Title: req.Title, Content: req.Content, SourceVersion: 1,
		CategoryIDs: topic.CategoryIds, Status: topic.Status,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	success(c, 200, gin.H{"topicId": fmt.Sprint(result.ID), "version": result.Version, "content": result.Content, "sourceVersion": result.SourceVersion, "moderationStatus": result.ModerationStatus, "url": postURL(result.ID, 1)})
}

func (h *Handler) editPost(c *gin.Context) {
	var req editPostRequest
	if !decode(c, &req) || !editVersion(c, req.Version) || !writeAllowed(c) {
		return
	}
	postID, ok := id(c.Param("postId"))
	if !ok {
		fail(c, 400, "invalid_request", "Invalid postId")
		return
	}
	row, found, err := posts.GetAgentPost(postID)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read post")
		return
	}
	if !found {
		fail(c, 404, "not_found", "Post is unavailable")
		return
	}
	if _, ok := h.loadTopic(c, fmt.Sprint(row.TopicId)); !ok {
		return
	}
	if row.UserId != current(c).User.Id {
		fail(c, 403, "permission_denied", "Only the author can edit a post")
		return
	}
	if row.PostNo <= 1 {
		fail(c, 422, "invalid_request", "Edit the first post through the topic endpoint")
		return
	}
	if !validateWriteContent(c, req.Content, nil, false) {
		return
	}
	post, err := postwriteservice.Update(postwriteservice.UpdateInput{PostID: postID, UserID: current(c).User.Id, ExpectedVersion: req.Version, Content: req.Content, SourceVersion: 1})
	if err != nil {
		writeError(c, err)
		return
	}
	success(c, 200, gin.H{"postId": fmt.Sprint(post.Id), "topicId": fmt.Sprint(post.TopicId), "version": post.ModerationVersion, "content": post.Content, "sourceVersion": post.SourceVersion, "moderationStatus": post.ModerationStatus, "url": postURL(post.TopicId, post.PostNo)})
}
