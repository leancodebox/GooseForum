package forum

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
)

func TestHomeMetaUsesRootCanonicalAndNoindexesVariants(t *testing.T) {
	c := testSEOContext("https://example.com/?sort=hot&page=2")

	meta := buildHomeMeta(c, 2, "hot", true)

	if meta.Canonical != "http://localhost/" {
		t.Fatalf("home canonical = %q, want root", meta.Canonical)
	}
	if meta.Robots != "noindex,follow" {
		t.Fatalf("home variant robots = %q, want noindex,follow", meta.Robots)
	}
	if meta.PrevURL != "/?sort=hot" || meta.NextURL != "/?page=3&sort=hot" {
		t.Fatalf("home prev/next = %q/%q", meta.PrevURL, meta.NextURL)
	}
}

func TestHomeMetaIndexesDefaultFirstPage(t *testing.T) {
	c := testSEOContext("https://example.com/")

	meta := buildHomeMeta(c, 1, "latest", true)

	if meta.Robots != "" {
		t.Fatalf("default home robots = %q, want empty", meta.Robots)
	}
	if meta.PrevURL != "" || meta.NextURL != "/?page=2" {
		t.Fatalf("default home prev/next = %q/%q", meta.PrevURL, meta.NextURL)
	}
}

func TestCategoryMetaNoindexesSortAndPaginationVariants(t *testing.T) {
	c := testSEOContext("https://example.com/c/general/8/l/hot?page=2")

	meta := buildCategoryMeta(c, &category.Entity{Id: 8, Name: "General", Slug: "general"}, 2, "hot", true)

	if meta.Canonical != "http://localhost/c/general/8" {
		t.Fatalf("category canonical = %q", meta.Canonical)
	}
	if meta.Robots != "noindex,follow" {
		t.Fatalf("category variant robots = %q, want noindex,follow", meta.Robots)
	}
	if meta.PrevURL != "/c/general/8/l/hot" || meta.NextURL != "/c/general/8/l/hot?page=3" {
		t.Fatalf("category prev/next = %q/%q", meta.PrevURL, meta.NextURL)
	}
}

func TestUserMetaNoindexesEmptyProfiles(t *testing.T) {
	c := testSEOContext("https://example.com/u/12")

	meta := buildUserMeta(c, &vo.UserCard{UserId: 12, Username: "empty"})

	if meta.Robots != "noindex,follow" {
		t.Fatalf("empty user robots = %q, want noindex,follow", meta.Robots)
	}
}

func TestUserMetaIndexesProfilesWithPublicContent(t *testing.T) {
	c := testSEOContext("https://example.com/u/12")

	meta := buildUserMeta(c, &vo.UserCard{UserId: 12, Username: "active", TopicCount: 1})

	if meta.Robots != "" {
		t.Fatalf("active user robots = %q, want empty", meta.Robots)
	}
}

func testSEOContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}
