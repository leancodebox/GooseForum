package routes

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func oldPath(parts ...string) string {
	return strings.Join(parts, "")
}

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
		oldPath("POST /api/forum/write-", "art", "icles"),
		oldPath("POST /api/forum/", "art", "icle-status"),
		oldPath("POST /api/forum/", "art", "icles-reply"),
		oldPath("POST /api/forum/", "art", "icles-reply-update"),
		oldPath("POST /api/forum/", "art", "icles-reply-delete"),
		oldPath("POST /api/forum/like-", "art", "icles"),
		oldPath("POST /api/forum/bookmark-", "art", "icle"),
		oldPath("POST /api/forum/watch-", "art", "icle"),
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
		oldPath("GET /api/forum/", "art", "icle-replies-window"),
		oldPath("POST /api/forum/moderation/", "art", "icle-status"),
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
		oldPath("POST /api/admin/", "art", "icles-list"),
		oldPath("POST /api/admin/", "art", "icle-source"),
		oldPath("POST /api/admin/", "art", "icle-edit"),
		oldPath("POST /api/admin/", "art", "icle-delete"),
		oldPath("POST /api/admin/", "art", "icle-pin-edit"),
		oldPath("POST /api/admin/", "art", "icle-categories-edit"),
	} {
		if registered[route] {
			t.Fatalf("%s should not be registered", route)
		}
	}
}
