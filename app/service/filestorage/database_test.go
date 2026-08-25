package filestorage

import (
	"bytes"
	"context"
	"io"
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

func setupDatabaseStoreTest(t *testing.T) {
	t.Helper()
	conn := db.Connect()
	if err := conn.AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatalf("migrate file data: %v", err)
	}
	conn.Where("1 = 1").Delete(&filedata.Entity{})
}

func TestDatabaseStoreRoundTrip(t *testing.T) {
	setupDatabaseStoreTest(t)
	service, err := New(NewDatabaseStore())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ctx := context.Background()
	request := PutRequest{
		Name:        "images/round-trip.webp",
		ContentType: "image/webp",
		Size:        7,
		UserId:      42,
		Body:        bytes.NewReader([]byte("content")),
	}
	stored, err := service.Put(ctx, request)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if stored.Id == 0 || stored.Name != request.Name || stored.Size != request.Size || stored.UserId != request.UserId || stored.StorageDriver != DatabaseDriver {
		t.Fatalf("stored metadata = %#v", stored)
	}

	stat, err := service.Stat(ctx, request.Name)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if stat.Id != stored.Id || stat.ContentType != request.ContentType || stat.Size != request.Size || stat.UserId != request.UserId {
		t.Fatalf("stat metadata = %#v", stat)
	}

	object, err := service.Open(ctx, request.Name)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	content, readErr := io.ReadAll(object.Body)
	closeErr := object.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read/close: %v / %v", readErr, closeErr)
	}
	if string(content) != "content" || object.Metadata.Size != request.Size {
		t.Fatalf("object = %#v, content = %q", object.Metadata, content)
	}

	if err := service.Delete(ctx, request.Name); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := service.Open(ctx, request.Name); err == nil {
		t.Fatal("deleted object can still be opened")
	}
}

func TestDatabaseStoreRejectsSizeMismatch(t *testing.T) {
	setupDatabaseStoreTest(t)
	service, err := New(NewDatabaseStore())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = service.Put(context.Background(), PutRequest{
		Name:        "images/bad-size.webp",
		ContentType: "image/webp",
		Size:        3,
		Body:        bytes.NewReader([]byte("four")),
	})
	if err == nil {
		t.Fatal("size mismatch succeeded")
	}
	if _, err := service.Stat(context.Background(), "images/bad-size.webp"); err == nil {
		t.Fatal("failed object write left metadata behind")
	}
}
