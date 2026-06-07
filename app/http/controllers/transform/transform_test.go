package transform

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestUser2userShow(t *testing.T) {
	createdAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	user := users.EntityComplete{
		Id:          42,
		Username:    "goose",
		Email:       "goose@example.com",
		Nickname:    "Goose",
		Bio:         "bio",
		Signature:   "signature",
		Prestige:    12,
		AvatarUrl:   "avatar.webp",
		CreatedAt:   createdAt,
		RoleId:      1,
		IsActivated: 1,
	}

	got := User2userShow(user)
	if got.UserId != user.Id {
		t.Fatalf("UserId = %d, want %d", got.UserId, user.Id)
	}
	if got.Username != user.Username || got.Email != user.Email || got.Nickname != user.Nickname {
		t.Fatalf("basic fields not mapped: %#v", got)
	}
	if got.AvatarUrl != "/file/img/avatar.webp" {
		t.Fatalf("AvatarUrl = %q, want /file/img/avatar.webp", got.AvatarUrl)
	}
	if !got.CanAccessAdmin {
		t.Fatalf("expected user to access admin")
	}
	if !got.CreateTime.Equal(createdAt) {
		t.Fatalf("CreateTime = %v, want %v", got.CreateTime, createdAt)
	}
}

func TestUser2UserDetailedVo(t *testing.T) {
	createdAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	user := users.EntityComplete{
		Id:          7,
		Username:    "duck",
		Email:       "duck@example.com",
		Nickname:    "Duck",
		AvatarUrl:   "/static/pic/default-avatar.webp",
		Bio:         "bio",
		Signature:   "signature",
		WebsiteName: "site",
		Website:     "https://example.com",
		Prestige:    5,
		CreatedAt:   createdAt,
	}

	got := User2UserDetailedVo(user)
	if got.Id != user.Id {
		t.Fatalf("Id = %d, want %d", got.Id, user.Id)
	}
	if got.AvatarUrl != "/static/pic/default-avatar.webp" {
		t.Fatalf("AvatarUrl = %q, want static avatar path", got.AvatarUrl)
	}
	if got.WebsiteName != user.WebsiteName || got.Website != user.Website {
		t.Fatalf("website fields not mapped: %#v", got)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}
}
