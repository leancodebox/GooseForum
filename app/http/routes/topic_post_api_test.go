package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestForumTopicPostWriteRoutesUseTopicPostNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoute(router)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		"POST /api/forum/topics/write",
		"POST /api/forum/topics/status",
		"POST /api/forum/posts/create",
		"POST /api/forum/posts/update",
		"POST /api/forum/posts/delete",
		"POST /api/forum/topics/like",
		"POST /api/forum/topics/bookmark",
		"POST /api/forum/topics/watch",
	} {
		if !registered[route] {
			t.Fatalf("%s was not registered", route)
		}
	}

	for _, route := range []string{
		"POST /api/forum/write-articles",
		"POST /api/forum/article-status",
		"POST /api/forum/articles-reply",
		"POST /api/forum/articles-reply-update",
		"POST /api/forum/articles-reply-delete",
		"POST /api/forum/like-articles",
		"POST /api/forum/bookmark-article",
		"POST /api/forum/watch-article",
	} {
		if registered[route] {
			t.Fatalf("%s should not be registered", route)
		}
	}
}

func TestForumTopicPostReadAndModerationRoutesUseTopicPostNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoute(router)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		"GET /api/forum/posts/window",
		"POST /api/forum/moderation/topic-status",
		"POST /api/forum/moderation/post-status",
	} {
		if !registered[route] {
			t.Fatalf("%s was not registered", route)
		}
	}

	for _, route := range []string{
		"GET /api/forum/article-replies-window",
		"POST /api/forum/moderation/article-status",
		"POST /api/forum/moderation/reply-status",
	} {
		if registered[route] {
			t.Fatalf("%s should not be registered", route)
		}
	}
}

func TestAdminTopicManagementRoutesUseTopicNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoute(router)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		"POST /api/admin/topics/list",
		"POST /api/admin/topics/source",
		"POST /api/admin/topics/edit",
		"POST /api/admin/topics/delete",
		"POST /api/admin/topics/pin-edit",
		"POST /api/admin/topics/categories-edit",
	} {
		if !registered[route] {
			t.Fatalf("%s was not registered", route)
		}
	}

	for _, route := range []string{
		"POST /api/admin/articles-list",
		"POST /api/admin/article-source",
		"POST /api/admin/article-edit",
		"POST /api/admin/article-delete",
		"POST /api/admin/article-pin-edit",
		"POST /api/admin/article-categories-edit",
	} {
		if registered[route] {
			t.Fatalf("%s should not be registered", route)
		}
	}
}
