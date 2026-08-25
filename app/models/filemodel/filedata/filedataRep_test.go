package filedata

import (
	"context"
	"testing"
	"time"

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
	if page.Total != 3 {
		t.Fatalf("total = %d, want 3", page.Total)
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

func TestFileResourcePageHandlesPendingIDGaps(t *testing.T) {
	setupFileDataTestDB(t)
	first, err := SaveFile(1, "images/first.webp", "image/webp", []byte("first"))
	if err != nil {
		t.Fatalf("save first: %v", err)
	}
	if _, err := CreateFileMetadata(context.Background(), 1, "images/pending.webp", "image/webp", 7, "s3"); err != nil {
		t.Fatalf("create pending: %v", err)
	}
	second, err := SaveFile(1, "images/second.webp", "image/webp", []byte("second"))
	if err != nil {
		t.Fatalf("save second: %v", err)
	}

	page1 := FileResourcePage(1, 1)
	page2 := FileResourcePage(2, 1)
	if page1.Total != 2 || page2.Total != 2 || len(page1.List) != 1 || len(page2.List) != 1 {
		t.Fatalf("pages = %#v / %#v", page1, page2)
	}
	if page1.List[0].Id != second.Id || page2.List[0].Id != first.Id {
		t.Fatalf("page ids = %d / %d, want %d / %d", page1.List[0].Id, page2.List[0].Id, second.Id, first.Id)
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
	if count := CountDailyUploads(11); count != 1 {
		t.Fatalf("pending upload attempt count = %d", count)
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

func TestListPendingFilesBeforeRespectsCutoffAndLimit(t *testing.T) {
	setupFileDataTestDB(t)
	ctx := context.Background()
	old, err := CreateFileMetadata(ctx, 1, "pending/old.webp", "image/webp", 10, "s3")
	if err != nil {
		t.Fatalf("create old: %v", err)
	}
	newer, err := CreateFileMetadata(ctx, 1, "pending/new.webp", "image/webp", 10, "s3")
	if err != nil {
		t.Fatalf("create new: %v", err)
	}
	cutoff := time.Now().Add(-time.Hour)
	if err := builder().Where("id = ?", old.Id).UpdateColumn("created_at", cutoff.Add(-time.Minute)).Error; err != nil {
		t.Fatalf("age old: %v", err)
	}
	if err := builder().Where("id = ?", newer.Id).UpdateColumn("created_at", cutoff.Add(time.Minute)).Error; err != nil {
		t.Fatalf("age new: %v", err)
	}
	items, err := ListPendingFilesBefore(ctx, cutoff, 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].Id != old.Id {
		t.Fatalf("items = %#v", items)
	}
}
