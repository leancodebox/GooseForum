package filedata

import (
	"context"
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
)

func setupFileDataTestDB(t *testing.T) {
	t.Helper()
	conn := db.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate file data: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})
}

func TestFileResourcePageListsFilesByIDRangeWithoutContent(t *testing.T) {
	setupFileDataTestDB(t)

	text, err := SaveFile(1, "docs/readme.txt", "text/plain", []byte("text"))
	if err != nil {
		t.Fatalf("save text file: %v", err)
	}
	first, err := SaveFile(2, "images/old.png", "image/png", []byte("old"))
	if err != nil {
		t.Fatalf("save first image: %v", err)
	}
	second, err := SaveFile(3, "images/new.webp", "image/webp", []byte("new-image"))
	if err != nil {
		t.Fatalf("save second image: %v", err)
	}

	page := FileResourcePage(1, 2)
	if page.MaxId != int64(second.Id) {
		t.Fatalf("maxId = %d, want %d", page.MaxId, second.Id)
	}
	if len(page.List) != 2 {
		t.Fatalf("len = %d, want 2", len(page.List))
	}
	if page.List[0].Id != second.Id || page.List[1].Id != first.Id {
		t.Fatalf("order = [%d,%d], want [%d,%d]", page.List[0].Id, page.List[1].Id, second.Id, first.Id)
	}

	next := FileResourcePage(2, 2)
	if len(next.List) != 1 {
		t.Fatalf("next len = %d, want 1", len(next.List))
	}
	if next.List[0].Id != text.Id || next.List[0].Type != "text/plain" {
		t.Fatalf("next row = id %d type %q, want text file id %d", next.List[0].Id, next.List[0].Type, text.Id)
	}
	if page.List[0].Size != int64(len("new-image")) {
		t.Fatalf("size = %d, want %d", page.List[0].Size, len("new-image"))
	}
	if page.List[0].Data != nil {
		t.Fatal("image resource list loaded blob content")
	}
	if page.List[0].URL != "/file/img/images/new.webp" {
		t.Fatalf("url = %q, want image access path", page.List[0].URL)
	}
}

func TestGetFileMetadataByNameDoesNotLoadBlob(t *testing.T) {
	setupFileDataTestDB(t)
	stored, err := SaveFile(7, "private/metadata.webp", "image/webp", []byte("secret-image"))
	if err != nil {
		t.Fatalf("save file: %v", err)
	}
	metadata, err := GetFileMetadataByName(stored.Name)
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	if metadata.Id != stored.Id || metadata.UserId != 7 || metadata.Type != "image/webp" || metadata.Size != int64(len("secret-image")) || metadata.Data != nil {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestGetFileMetadataFallsBackToBlobSizeForLegacyRows(t *testing.T) {
	setupFileDataTestDB(t)
	result := builder().Create(map[string]any{
		"name":           "legacy/image.webp",
		"assert_type":    "image/webp",
		"content":        []byte("legacy-content"),
		"file_size":      0,
		"storage_driver": "database",
		"user_id":        9,
	})
	if result.Error != nil {
		t.Fatalf("insert legacy file: %v", result.Error)
	}
	metadata, err := GetFileMetadataByName("legacy/image.webp")
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	if metadata.Size != int64(len("legacy-content")) || metadata.Data != nil {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestPendingFileIsHiddenUntilMarkedReady(t *testing.T) {
	setupFileDataTestDB(t)
	ctx := context.Background()
	metadata, err := CreateFileMetadata(ctx, 11, "pending/image.webp", "image/webp", 7, "database")
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if metadata.StorageStatus != StorageStatusPending {
		t.Fatalf("status = %q", metadata.StorageStatus)
	}
	if _, err := GetFileMetadataByName(metadata.Name); err == nil {
		t.Fatal("pending metadata is publicly readable")
	}
	if page := FileResourcePage(1, 20); len(page.List) != 0 {
		t.Fatalf("pending file listed: %#v", page.List)
	}
	if count := CountUserUploadsToday(11); count != 0 {
		t.Fatalf("pending upload count = %d", count)
	}
	if err := UpdateFileContent(ctx, metadata.Name, []byte("content")); err != nil {
		t.Fatalf("write content: %v", err)
	}
	ready, err := MarkFileReady(ctx, metadata.Name)
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	if ready.StorageStatus != StorageStatusReady || ready.Size != 7 {
		t.Fatalf("ready metadata = %#v", ready)
	}
}
