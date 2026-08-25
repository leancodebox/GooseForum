package fileusageservice

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
)

func TestFileNameFromURL(t *testing.T) {
	tests := map[string]string{
		"/file/img/2026/06/a.webp":              "2026/06/a.webp",
		"https://example.com/file/img/a/b.webp": "a/b.webp",
		"avatars/1/avatar.webp":                 "avatars/1/avatar.webp",
		"/static/pic/default-avatar.webp":       "",
		"https://example.com/static/a.webp":     "",
		"../secret.webp":                        "",
	}
	for input, want := range tests {
		if got := fileNameFromURL(input); got != want {
			t.Fatalf("fileNameFromURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAddUploadOwnerIsIdempotent(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&fileUsage.Entity{}); err != nil {
		t.Fatalf("migrate file usage: %v", err)
	}
	fileName := "tests/idempotent-owner.webp"
	conn.Where("file_name = ?", fileName).Delete(&fileUsage.Entity{})
	t.Cleanup(func() { conn.Where("file_name = ?", fileName).Delete(&fileUsage.Entity{}) })
	if err := AddUploadOwner(991001, fileName); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := AddUploadOwner(991001, fileName); err != nil {
		t.Fatalf("second add: %v", err)
	}
	usages, err := fileUsage.GetByFileName(fileName)
	if err != nil {
		t.Fatalf("list usages: %v", err)
	}
	if len(usages) != 1 {
		t.Fatalf("usage count = %d", len(usages))
	}
}
