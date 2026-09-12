package urlconfig

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestVersionBuiltinAvatar(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"/static/pic/1.webp", "/static/pic/1.webp?t=1788958424"},
		{"/static/pic/12_medium.webp?t=old&size=96#avatar", "/static/pic/12_medium.webp?size=96&t=1788958424#avatar"},
		{"https://cdn.example.com/static/pic/2.webp", "https://cdn.example.com/static/pic/2.webp?t=1788958424"},
		{"/static/pic/13.webp", "/static/pic/13.webp?t=1788958424"},
		{"/static/pic/default-avatar.webp", "/static/pic/default-avatar.webp?t=1788958424"},
		{"/static/pic/icons/favicon-32.png", "/static/pic/icons/favicon-32.png?t=1788958424"},
		{"/static/picture/avatar.webp", "/static/picture/avatar.webp"},
		{"/file/img/avatar.webp?t=123", "/file/img/avatar.webp?t=123"},
	} {
		if got := VersionBuiltinAvatar(tt.input); got != tt.want {
			t.Errorf("VersionBuiltinAvatar(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefaultAvatarUsesCDNWhenConfigured(t *testing.T) {
	old := preferences.GetString("app.cdn_url", "")
	t.Cleanup(func() {
		preferences.Set("app.cdn_url", old)
	})

	preferences.Set("app.cdn_url", "")
	if got := GetDefaultAvatar(); got != "/static/pic/default-avatar.webp?t=1788958424" {
		t.Fatalf("default avatar = %q, want local path", got)
	}

	preferences.Set("app.cdn_url", "https://cdn.example.com")
	if got := GetDefaultAvatar(); got != "https://cdn.example.com/static/pic/default-avatar.webp?t=1788958424" {
		t.Fatalf("cdn default avatar = %q, want CDN path", got)
	}
}

func TestBannedAvatarUsesCDNWhenConfigured(t *testing.T) {
	old := preferences.GetString("app.cdn_url", "")
	t.Cleanup(func() {
		preferences.Set("app.cdn_url", old)
	})

	preferences.Set("app.cdn_url", "")
	if got := GetBannedAvatar(); got != "/static/pic/banned-avatar.png?t=1788958424" {
		t.Fatalf("banned avatar = %q, want local path", got)
	}

	preferences.Set("app.cdn_url", "https://cdn.example.com")
	if got := GetBannedAvatar(); got != "https://cdn.example.com/static/pic/banned-avatar.png?t=1788958424" {
		t.Fatalf("cdn banned avatar = %q, want CDN path", got)
	}
}

func TestFilePath(t *testing.T) {
	if got := FilePath("avatar.webp"); got != "/file/img/avatar.webp" {
		t.Fatalf("FilePath = %q, want /file/img/avatar.webp", got)
	}
	if got := FilePath("/nested/avatar.webp"); got != "/file/img/nested/avatar.webp" {
		t.Fatalf("FilePath nested = %q, want normalized image path", got)
	}
}

func TestStoredFilePathUsesS3PublicURLOnlyForS3Objects(t *testing.T) {
	old := preferences.GetString("storage.s3.publicUrl")
	preferences.Set("storage.s3.publicUrl", "https://files.example.com/forum/")
	t.Cleanup(func() { preferences.Set("storage.s3.publicUrl", old) })

	if got := StoredFilePath("s3", "2026/09/image.webp"); got != "https://files.example.com/forum/2026/09/image.webp" {
		t.Fatalf("S3 file path = %q", got)
	}
	if got := StoredFilePath("database", "2026/09/image.webp"); got != "/file/img/2026/09/image.webp" {
		t.Fatalf("database file path = %q", got)
	}
}

func TestCategoryEscapesSlug(t *testing.T) {
	if got := Category("吐槽/脑洞", 8); got != "/c/%E5%90%90%E6%A7%BD%2F%E8%84%91%E6%B4%9E/8" {
		t.Fatalf("Category = %q", got)
	}
}

func TestStaticRoutes(t *testing.T) {
	tests := map[string]string{
		"home":          Home(),
		"post":          Post(),
		"docs":          Docs(),
		"links":         Links(),
		"sponsors":      Sponsors(),
		"publish":       Publish(),
		"search":        Search(),
		"register":      Register(),
		"login":         Login(),
		"messages":      Messages(),
		"drafts":        Drafts(),
		"settings":      Settings(),
		"notifications": Notifications(),
		"activate":      Activate(),
		"resetPassword": ResetPassword(),
		"admin":         Admin(),
		"rss":           Rss(),
	}

	want := map[string]string{
		"home":          "/",
		"post":          "/p/post",
		"docs":          "/docs",
		"links":         "/links",
		"sponsors":      "/sponsors",
		"publish":       "/publish",
		"search":        "/search",
		"register":      "/login",
		"login":         "/login",
		"messages":      "/messages",
		"drafts":        "/drafts",
		"settings":      "/settings",
		"notifications": "/notifications",
		"activate":      "/activate",
		"resetPassword": "/reset-password",
		"admin":         "/admin",
		"rss":           "/rss.xml",
	}

	for name, got := range tests {
		if got != want[name] {
			t.Fatalf("%s route = %q, want %q", name, got, want[name])
		}
	}
}

func TestDynamicRoutes(t *testing.T) {
	if got := PostDetail(42); got != "/p/post/42" {
		t.Fatalf("PostDetail = %q, want /p/post/42", got)
	}
	if got := User("alice"); got != "/u/alice" {
		t.Fatalf("User = %q, want /u/alice", got)
	}
	if got := DocsProject("goose"); got != "/docs/goose" {
		t.Fatalf("DocsProject = %q, want /docs/goose", got)
	}
	if got := DocsContent("goose", "v1", "intro"); got != "/docs/goose/v1/intro" {
		t.Fatalf("DocsContent = %q, want /docs/goose/v1/intro", got)
	}
}
