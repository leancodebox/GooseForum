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

func TestOpaqueExtensionImagesAreNotTrackedAfterReview(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&fileUsage.Entity{}); err != nil {
		t.Fatal(err)
	}
	const userID uint64 = 991005
	const targetID uint64 = 991006
	const fileName = "tests/opaque-review.webp"
	conn.Where("file_name = ?", fileName).Delete(&fileUsage.Entity{})
	t.Cleanup(func() { conn.Where("file_name = ?", fileName).Delete(&fileUsage.Entity{}) })
	content := "[future]\n\n**title**\n\n![hidden](/file/img/" + fileName + ")\n\n[/future]"
	for _, replace := range []func(uint64, uint64, string, uint8){ReplaceTopic, ReplacePost} {
		replace(targetID, userID, content, 0)
		var count int64
		query := conn.Model(&fileUsage.Entity{}).Where("file_name = ? AND usage_type = ?", fileName, fileUsage.UsageInlineImage)
		if err := query.Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("legacy inline usages=%d %v", count, err)
		}
		replace(targetID, userID, content, 1)
		if err := query.Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("opaque inline usage persisted after review: %d %v", count, err)
		}
	}
}
