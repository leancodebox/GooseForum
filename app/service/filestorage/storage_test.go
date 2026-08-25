package filestorage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

type fakeStore struct {
	driver    string
	putName   string
	putSize   int64
	putData   []byte
	putErr    error
	deleteErr error
	object    *StoredObject
	deleted   string
}

func (store *fakeStore) Driver() string { return store.driver }

func (store *fakeStore) Put(_ context.Context, request WriteRequest) error {
	store.putName = request.Name
	store.putSize = request.Size
	store.putData, _ = io.ReadAll(request.Body)
	return store.putErr
}

func (store *fakeStore) Open(_ context.Context, _ string) (*StoredObject, error) {
	return store.object, nil
}

func (store *fakeStore) Delete(_ context.Context, name string) error {
	store.deleted = name
	return store.deleteErr
}

type fakeMetadataRepository struct {
	metadata          *Metadata
	deleted           string
	deleteCtxCanceled bool
}

func (repository *fakeMetadataRepository) Create(_ context.Context, metadata Metadata) (*Metadata, error) {
	metadata.Id = 1
	metadata.StorageStatus = filedata.StorageStatusPending
	repository.metadata = &metadata
	return repository.metadata, nil
}

func (repository *fakeMetadataRepository) MarkReady(_ context.Context, name string) (*Metadata, error) {
	if repository.metadata == nil || repository.metadata.Name != name {
		return nil, errors.New("file not found")
	}
	repository.metadata.StorageStatus = filedata.StorageStatusReady
	copy := *repository.metadata
	return &copy, nil
}

func (repository *fakeMetadataRepository) Get(_ context.Context, name string) (*Metadata, error) {
	if repository.metadata == nil || repository.metadata.Name != name || repository.metadata.StorageStatus != filedata.StorageStatusReady {
		return nil, errors.New("file not found")
	}
	copy := *repository.metadata
	return &copy, nil
}

func (repository *fakeMetadataRepository) GetPending(_ context.Context, name string) (*Metadata, error) {
	if repository.metadata == nil || repository.metadata.Name != name || repository.metadata.StorageStatus != filedata.StorageStatusPending {
		return nil, errors.New("file not found")
	}
	copy := *repository.metadata
	return &copy, nil
}

func (repository *fakeMetadataRepository) ListPendingBefore(_ context.Context, before time.Time, limit int) ([]Metadata, error) {
	if repository.metadata == nil || repository.metadata.StorageStatus != filedata.StorageStatusPending || !repository.metadata.CreatedAt.Before(before) || limit == 0 {
		return nil, nil
	}
	return []Metadata{*repository.metadata}, nil
}

func (repository *fakeMetadataRepository) Delete(ctx context.Context, name string) error {
	repository.deleted = name
	repository.deleteCtxCanceled = ctx.Err() != nil
	repository.metadata = nil
	return nil
}

func TestServiceSeparatesMetadataFromObjectStorage(t *testing.T) {
	store := &fakeStore{
		driver: "fake",
		object: &StoredObject{Size: 4, Body: io.NopCloser(bytes.NewReader([]byte("data")))},
	}
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := PutRequest{
		Name:        "images/example.webp",
		ContentType: "image/webp",
		Size:        4,
		UserId:      7,
		Body:        bytes.NewReader([]byte("data")),
	}

	stored, err := service.Put(context.Background(), request)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if stored.Name != request.Name || stored.UserId != request.UserId || stored.StorageDriver != store.driver || stored.StorageStatus != filedata.StorageStatusReady {
		t.Fatalf("stored metadata = %#v", stored)
	}
	if store.putName != request.Name || store.putSize != request.Size || string(store.putData) != "data" {
		t.Fatalf("stored object = name %q size %d data %q", store.putName, store.putSize, store.putData)
	}

	object, err := service.Open(context.Background(), request.Name)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = object.Body.Close() }()
	content, err := io.ReadAll(object.Body)
	if err != nil || string(content) != "data" || object.Metadata.UserId != request.UserId {
		t.Fatalf("object = %#v, content = %q, err = %v", object.Metadata, content, err)
	}

	if err := service.Delete(context.Background(), request.Name); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if store.deleted != request.Name || repository.deleted != request.Name {
		t.Fatalf("deleted store/repository = %q/%q", store.deleted, repository.deleted)
	}
}

func TestServiceRollsBackMetadataWithUncancelledContext(t *testing.T) {
	store := &fakeStore{driver: "fake", putErr: errors.New("put failed")}
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = service.Put(ctx, PutRequest{
		Name:        "images/failure.webp",
		ContentType: "image/webp",
		Size:        4,
		Body:        bytes.NewReader([]byte("data")),
	})
	if err == nil {
		t.Fatal("put succeeded")
	}
	if store.deleted != "images/failure.webp" || repository.deleted != "images/failure.webp" || repository.deleteCtxCanceled {
		t.Fatalf("rollback = store %q repository %q canceled %v", store.deleted, repository.deleted, repository.deleteCtxCanceled)
	}
}

func TestServiceRejectsMetadataForAnotherDriver(t *testing.T) {
	store := &fakeStore{driver: "s3"}
	repository := &fakeMetadataRepository{metadata: &Metadata{Name: "legacy.webp", StorageDriver: DatabaseDriver, StorageStatus: filedata.StorageStatusReady}}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.Open(context.Background(), "legacy.webp"); err == nil {
		t.Fatal("open with wrong driver succeeded")
	}
	if err := service.Delete(context.Background(), "legacy.webp"); err == nil {
		t.Fatal("delete with wrong driver succeeded")
	}
}

func TestServiceRoutesExistingFilesToTheirStorageDriver(t *testing.T) {
	s3Store := &fakeStore{driver: S3Driver}
	databaseStore := &fakeStore{
		driver: DatabaseDriver,
		object: &StoredObject{Size: 6, Body: io.NopCloser(bytes.NewReader([]byte("legacy")))},
	}
	repository := &fakeMetadataRepository{metadata: &Metadata{
		Name: "legacy.webp", ContentType: "image/webp", Size: 6,
		StorageDriver: DatabaseDriver, StorageStatus: filedata.StorageStatusReady,
	}}
	service, err := newService(s3Store, repository, databaseStore)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	object, err := service.Open(context.Background(), "legacy.webp")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	content, _ := io.ReadAll(object.Body)
	_ = object.Body.Close()
	if string(content) != "legacy" {
		t.Fatalf("content = %q", content)
	}
	if err := service.Delete(context.Background(), "legacy.webp"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if databaseStore.deleted != "legacy.webp" || s3Store.deleted != "" {
		t.Fatalf("deleted database/s3 = %q/%q", databaseStore.deleted, s3Store.deleted)
	}
}

func TestServiceCleansExpiredPendingUploads(t *testing.T) {
	store := &fakeStore{driver: S3Driver}
	repository := &fakeMetadataRepository{metadata: &Metadata{
		Name: "pending.webp", StorageDriver: S3Driver, StorageStatus: filedata.StorageStatusPending,
		CreatedAt: time.Now().Add(-3 * time.Hour),
	}}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	removed, err := service.CleanupPending(context.Background(), time.Now().Add(-2*time.Hour), 10)
	if err != nil || removed != 1 {
		t.Fatalf("cleanup = %d, %v", removed, err)
	}
	if repository.metadata != nil || store.deleted != "pending.webp" {
		t.Fatalf("cleanup state = metadata %#v deleted %q", repository.metadata, store.deleted)
	}
}

func TestServiceRetainsPendingMetadataWhenObjectCleanupFails(t *testing.T) {
	store := &fakeStore{driver: S3Driver, deleteErr: errors.New("temporary delete failure")}
	repository := &fakeMetadataRepository{metadata: &Metadata{
		Name: "pending.webp", StorageDriver: S3Driver, StorageStatus: filedata.StorageStatusPending,
		CreatedAt: time.Now().Add(-3 * time.Hour),
	}}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	removed, err := service.CleanupPending(context.Background(), time.Now().Add(-2*time.Hour), 10)
	if err == nil || removed != 0 {
		t.Fatalf("cleanup = %d, %v", removed, err)
	}
	if repository.metadata == nil {
		t.Fatal("metadata was deleted after object cleanup failed")
	}
}

func TestServiceRejectsInvalidConfigurationAndPut(t *testing.T) {
	store := &fakeStore{driver: "fake"}
	service, err := newService(store, &fakeMetadataRepository{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	tests := []PutRequest{
		{ContentType: "image/webp", Size: 1, Body: bytes.NewReader([]byte("x"))},
		{Name: "x", Size: 1, Body: bytes.NewReader([]byte("x"))},
		{Name: "x", ContentType: "image/webp", Size: -1, Body: bytes.NewReader([]byte("x"))},
		{Name: "x", ContentType: "image/webp", Size: 1},
	}
	for _, request := range tests {
		if _, err := service.Put(context.Background(), request); err == nil {
			t.Fatalf("Put(%#v) succeeded", request)
		}
	}
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) succeeded")
	}
	var nilStore *fakeStore
	if _, err := New(nilStore); err == nil {
		t.Fatal("New(typed nil) succeeded")
	}
	if _, err := newService(store, nil); err == nil {
		t.Fatal("newService(store, nil) succeeded")
	}
	if _, err := newService(store, &fakeMetadataRepository{}, &fakeStore{driver: store.driver}); err == nil {
		t.Fatal("duplicate driver configuration succeeded")
	}
}
