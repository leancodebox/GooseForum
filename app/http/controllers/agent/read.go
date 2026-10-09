package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"gorm.io/gorm"
)

type topicRow = topics.AgentTopic
type postRow = posts.AgentPost
type cursor struct {
	Kind, Filter string
	ID           uint64
}

func cursorToken(kind, filter string, value uint64) string {
	data, _ := json.Marshal(cursor{kind, filter, value})
	return base64.RawURLEncoding.EncodeToString(data)
}
func readCursor(c *gin.Context, kind, filter string) (uint64, bool) {
	raw := c.Query("cursor")
	if raw == "" {
		return 0, true
	}
	if len(raw) > 512 {
		fail(c, 400, "invalid_cursor", "Cursor is too long")
		return 0, false
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	var value cursor
	if err != nil || len(raw) > 512 || json.Unmarshal(data, &value) != nil || value.Kind != kind || value.Filter != filter || value.ID == 0 {
		fail(c, 400, "invalid_cursor", "Cursor does not match this query")
		return 0, false
	}
	return value.ID, true
}

func baseURL() string {
	return strings.TrimRight(hotdataserve.GetSiteSettingsConfigCache().SiteUrl, "/")
}
func postURL(topicID, postNo uint64) string {
	return fmt.Sprintf("%s/p/post/%d/%d", baseURL(), topicID, postNo)
}

func postData(p postRow) gin.H {
	return gin.H{"id": fmt.Sprint(p.Id), "topicId": fmt.Sprint(p.TopicId), "author": gin.H{"id": fmt.Sprint(p.UserId)}, "postNo": p.PostNo, "replyToPostId": optionalID(p.ReplyToPostId), "content": p.Content, "contentFormat": "markdown", "sourceVersion": p.SourceVersion, "createdAt": p.CreatedAt.UTC(), "updatedAt": p.UpdatedAt.UTC(), "url": postURL(p.TopicId, p.PostNo)}
}
func optionalID(value uint64) any {
	if value == 0 {
		return nil
	}
	return fmt.Sprint(value)
}

func (h *Handler) me(c *gin.Context) {
	a := current(c)
	success(c, 200, gin.H{"id": fmt.Sprint(a.User.Id), "username": a.User.Username, "clientId": a.ClientID, "scopes": a.Scopes, "canWrite": a.canWrite(), "restriction": a.User.EffectiveRestriction(time.Now())})
}

func (a *actor) canWrite() bool {
	if a.User.Id == 0 {
		return false
	}
	if a.writable == nil {
		_, err := component.CheckUserPermission(&a.User, component.PermissionActionWrite)
		allowed := err == nil
		a.writable = &allowed
	}
	return *a.writable
}

func (h *Handler) categories(c *gin.Context) {
	a := current(c)
	rows, err := category.ListAgentCategories(a.Access.ReadableCategoryIDs(), a.Access.HasGlobalManage())
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read categories")
		return
	}
	writable := a.canWrite()
	data := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		data = append(data, gin.H{"id": fmt.Sprint(row.Id), "name": row.Name, "description": row.Desc, "slug": row.Slug, "capabilities": gin.H{"createTopic": writable && hasScope(a, "topics:create") && a.Access.CanCreateCategory(row.Id), "reply": writable && hasScope(a, "posts:create") && a.Access.CanReplyCategory(row.Id)}})
	}
	success(c, 200, data)
}

func hasScope(a *actor, scope string) bool {
	for _, s := range a.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func (a *actor) topicQuery() topics.AgentTopicQuery {
	return topics.AgentTopicQuery{ReadableCategoryIDs: a.Access.ReadableCategoryIDs(), GlobalManage: a.Access.HasGlobalManage()}
}

func (h *Handler) loadTopic(c *gin.Context, value string) (topicRow, bool) {
	topicID, ok := id(value)
	if !ok {
		fail(c, 400, "invalid_request", "Invalid topic ID")
		return topicRow{}, false
	}
	row, found, err := topics.GetAgentTopic(current(c).topicQuery(), topicID)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read topic")
		return row, false
	}
	if !found {
		fail(c, 404, "not_found", "Topic is unavailable")
		return row, false
	}
	return row, true
}

func (h *Handler) topicData(c *gin.Context, row topicRow) gin.H {
	a := current(c)
	return gin.H{"id": fmt.Sprint(row.Id), "title": row.Title, "author": gin.H{"id": fmt.Sprint(row.UserId)}, "categoryIds": []string{fmt.Sprint(row.MainCategoryId)}, "createdAt": row.CreatedAt.UTC(), "updatedAt": row.UpdatedAt.UTC(), "url": postURL(row.Id, 1), "capabilities": gin.H{"reply": a.canWrite() && hasScope(a, "posts:create") && a.Access.CanReplyCategory(row.MainCategoryId)}}
}

func (h *Handler) topics(c *gin.Context) {
	n, ok := limit(c)
	if !ok {
		return
	}
	sort := c.DefaultQuery("sort", "newest")
	if sort != "newest" {
		fail(c, 400, "invalid_request", "Only newest sorting is supported")
		return
	}
	filter := c.Query("categoryId")
	query := current(c).topicQuery()
	if filter != "" {
		categoryID, valid := id(filter)
		if !valid {
			fail(c, 400, "invalid_request", "Invalid category ID")
			return
		}
		query.CategoryID = categoryID
	}
	before, ok := readCursor(c, "topics", filter)
	if !ok {
		return
	}
	query.BeforeID = before
	rows, err := topics.ListAgentTopics(query, n+1)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read topics")
		return
	}
	more := len(rows) > n
	if more {
		rows = rows[:n]
	}
	data := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		data = append(data, h.topicData(c, row))
	}
	next := ""
	if more {
		next = cursorToken("topics", filter, rows[len(rows)-1].Id)
	}
	page(c, data, gin.H{"hasMore": more, "nextCursor": next})
}

func (h *Handler) topic(c *gin.Context) {
	row, ok := h.loadTopic(c, c.Param("topicId"))
	if !ok {
		return
	}
	first, found, err := posts.GetAgentPost(row.FirstPostId)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read first post")
		return
	}
	if !found {
		fail(c, 404, "not_found", "Topic is unavailable")
		return
	}
	items := []postRow{first}
	if !h.filterReplyTargets(c, items) {
		return
	}
	first = items[0]
	data := h.topicData(c, row)
	data["firstPost"] = postData(first)
	success(c, 200, data)
}

func (h *Handler) posts(c *gin.Context) {
	topic, ok := h.loadTopic(c, c.Param("topicId"))
	if !ok {
		return
	}
	n, ok := limit(c)
	if !ok {
		return
	}
	after, ok := readCursor(c, "posts", fmt.Sprint(topic.Id))
	if !ok {
		return
	}
	rows, err := posts.ListAgentPosts(topic.Id, after, n+1)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read posts")
		return
	}
	more := len(rows) > n
	if more {
		rows = rows[:n]
	}
	if !h.filterReplyTargets(c, rows) {
		return
	}
	data := make([]gin.H, 0, len(rows))
	for _, p := range rows {
		data = append(data, postData(p))
	}
	next := ""
	if more {
		next = cursorToken("posts", fmt.Sprint(topic.Id), rows[len(rows)-1].PostNo)
	}
	page(c, data, gin.H{"hasMore": more, "nextCursor": next})
}

func (h *Handler) post(c *gin.Context) {
	postID, ok := id(c.Param("postId"))
	if !ok {
		fail(c, 400, "invalid_request", "Invalid post ID")
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
	if _, ok = h.loadTopic(c, fmt.Sprint(row.TopicId)); !ok {
		return
	}
	items := []postRow{row}
	if !h.filterReplyTargets(c, items) {
		return
	}
	row = items[0]
	success(c, 200, postData(row))
}

func (h *Handler) search(c *gin.Context) {
	n, ok := limit(c)
	if !ok {
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	p, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || p < 1 || p > 1000 || query == "" || len(query) > 500 {
		fail(c, 400, "invalid_request", "q and a page between 1 and 1000 are required")
		return
	}
	if !meiliconnect.IsAvailable() {
		fail(c, 503, "search_unavailable", "Search is unavailable")
		return
	}
	a := current(c)
	result, err := searchservice.SearchTopics(searchservice.SearchRequest{Query: query, Categories: a.Access.ReadableCategoryIDs(), FilterByCategories: !a.Access.HasGlobalManage(), Limit: n, Offset: (p - 1) * n})
	if err != nil || result == nil {
		fail(c, 503, "search_unavailable", "Search is unavailable")
		return
	}
	ids := make([]uint64, 0, len(result.Results))
	for _, hit := range result.Results {
		ids = append(ids, hit.ID)
	}
	rows, err := topics.GetAgentTopicsByIDs(a.topicQuery(), ids)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read results")
		return
	}
	byID := map[uint64]topicRow{}
	for _, row := range rows {
		byID[row.Id] = row
	}
	data := make([]gin.H, 0, len(rows))
	for _, hit := range result.Results {
		if row, ok := byID[hit.ID]; ok {
			data = append(data, h.topicData(c, row))
		}
	}
	// Do not expose counts from stale or now-private index documents.
	page(c, data, gin.H{"page": p, "hasMore": len(result.Results) == n})
}

func notFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func (h *Handler) filterReplyTargets(c *gin.Context, rows []postRow) bool {
	ids := []uint64{}
	for _, row := range rows {
		if row.ReplyToPostId > 0 {
			ids = append(ids, row.ReplyToPostId)
		}
	}
	if len(ids) == 0 {
		return true
	}
	parents, err := posts.GetAgentReplyTargets(ids)
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read reply targets")
		return false
	}
	visible := map[uint64]uint64{}
	for _, parent := range parents {
		visible[parent.Id] = parent.TopicId
	}
	for i, row := range rows {
		if visible[row.ReplyToPostId] != row.TopicId {
			rows[i].ReplyToPostId = 0
		}
	}
	return true
}
