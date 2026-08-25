package filestorage

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeS3Client struct {
	putName        string
	putContentType string
	putSize        int64
	putData        []byte
	opened         string
	openErr        error
	removed        string
	stat           s3ObjectInfo
	statErr        error
	presignName    string
	presignType    string
	presignSize    int64
	presignExpiry  time.Time
	presignErr     error
}

func (client *fakeS3Client) PutObject(_ context.Context, name string, body io.Reader, size int64, contentType string) error {
	client.putName = name
	client.putContentType = contentType
	client.putSize = size
	client.putData, _ = io.ReadAll(body)
	return nil
}

func (client *fakeS3Client) OpenObject(_ context.Context, name string) (io.ReadCloser, error) {
	client.opened = name
	if client.openErr != nil {
		return nil, client.openErr
	}
	return io.NopCloser(bytes.NewReader([]byte("object"))), nil
}

func (client *fakeS3Client) StatObject(_ context.Context, _ string) (s3ObjectInfo, error) {
	return client.stat, client.statErr
}

func (client *fakeS3Client) RemoveObject(_ context.Context, name string) error {
	client.removed = name
	return nil
}

func (client *fakeS3Client) PresignPost(_ context.Context, name, contentType string, size int64, expiresAt time.Time) (string, map[string]string, error) {
	client.presignName = name
	client.presignType = contentType
	client.presignSize = size
	client.presignExpiry = expiresAt
	return "https://objects.example.com/forum", map[string]string{"key": name}, client.presignErr
}

func TestS3StoreImplementsObjectOperations(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	ctx := context.Background()
	request := WriteRequest{Name: "images/example.webp", ContentType: "image/webp", Size: 6, Body: strings.NewReader("object")}

	if err := store.Put(ctx, request); err != nil {
		t.Fatalf("put: %v", err)
	}
	if client.putName != request.Name || client.putContentType != request.ContentType || client.putSize != request.Size || string(client.putData) != "object" {
		t.Fatalf("put request = %q %q %d %q", client.putName, client.putContentType, client.putSize, client.putData)
	}
	object, err := store.Open(ctx, request.Name)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	content, readErr := io.ReadAll(object.Body)
	closeErr := object.Body.Close()
	if readErr != nil || closeErr != nil || string(content) != "object" || object.Size != request.Size {
		t.Fatalf("object = size %d content %q errors %v/%v", object.Size, content, readErr, closeErr)
	}
	if err := store.Delete(ctx, request.Name); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if client.removed != request.Name {
		t.Fatalf("removed = %q", client.removed)
	}
}

func TestS3StorePresignsConstrainedUploadAndVerifiesObject(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	store.now = func() time.Time { return now }
	request := DirectUploadRequest{Name: "images/example.webp", ContentType: "image/webp", Size: 6, ExpiresIn: 15 * time.Minute}

	upload, err := store.PresignUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if upload.Method != "POST" || upload.URL != "https://objects.example.com/forum" || upload.Fields["key"] != request.Name || !upload.ExpiresAt.Equal(now.Add(request.ExpiresIn)) {
		t.Fatalf("upload = %#v", upload)
	}
	if client.presignType != request.ContentType || client.presignSize != request.Size || !client.presignExpiry.Equal(upload.ExpiresAt) {
		t.Fatalf("presign constraints = %q %d %s", client.presignType, client.presignSize, client.presignExpiry)
	}
	if err := store.VerifyUpload(context.Background(), request); err != nil {
		t.Fatalf("verify: %v", err)
	}

	client.stat.Size++
	if err := store.VerifyUpload(context.Background(), request); err == nil {
		t.Fatal("size mismatch verified")
	}
	client.stat.Size = request.Size
	client.stat.ContentType = "image/png"
	if err := store.VerifyUpload(context.Background(), request); err == nil {
		t.Fatal("content type mismatch verified")
	}
	client.stat.ContentType = ""
	if err := store.VerifyUpload(context.Background(), request); err == nil {
		t.Fatal("empty content type verified")
	}
}

func TestServiceCoordinatesDirectUploadLifecycle(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := DirectUploadRequest{Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute}
	session, err := service.BeginDirectUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if session.Metadata.StorageStatus != "pending" || session.Metadata.UserId != request.UserId || session.Upload.URL == "" {
		t.Fatalf("session = %#v", session)
	}
	if _, err := service.Stat(context.Background(), request.Name); err == nil {
		t.Fatal("pending direct upload is readable")
	}
	ready, err := service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{Name: request.Name, UserId: request.UserId})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if ready.StorageStatus != "ready" || ready.Name != request.Name {
		t.Fatalf("ready metadata = %#v", ready)
	}
	retried, err := service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{Name: request.Name, UserId: request.UserId})
	if err != nil || retried.Id != ready.Id {
		t.Fatalf("idempotent complete = %#v, %v", retried, err)
	}
}

func TestServiceRejectsCompletingAnotherUsersDirectUpload(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := DirectUploadRequest{Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute}
	if _, err := service.BeginDirectUpload(context.Background(), request); err != nil {
		t.Fatalf("begin: %v", err)
	}
	_, err = service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{Name: request.Name, UserId: 13})
	if !errors.Is(err, ErrDirectUploadOwnerMismatch) {
		t.Fatalf("complete error = %v", err)
	}
	if repository.metadata.StorageStatus != "pending" {
		t.Fatalf("metadata status = %q", repository.metadata.StorageStatus)
	}
}

func TestServiceRollsBackFailedDirectUploadSigning(t *testing.T) {
	client := &fakeS3Client{presignErr: errors.New("sign failed")}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = service.BeginDirectUpload(context.Background(), DirectUploadRequest{
		Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute,
	})
	if err == nil {
		t.Fatal("begin succeeded")
	}
	if repository.metadata != nil || client.removed != "images/direct.webp" {
		t.Fatalf("rollback = metadata %#v removed %q", repository.metadata, client.removed)
	}
}

func TestServiceRemovesObjectRejectedByContentValidator(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := DirectUploadRequest{Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute}
	if _, err := service.BeginDirectUpload(context.Background(), request); err != nil {
		t.Fatalf("begin: %v", err)
	}
	_, err = service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{
		Name: request.Name, UserId: request.UserId,
		Validator: func(io.Reader, string) error { return ErrDirectUploadInvalidObject },
	})
	if !errors.Is(err, ErrDirectUploadInvalidObject) {
		t.Fatalf("complete error = %v", err)
	}
	if repository.metadata != nil || client.removed != request.Name {
		t.Fatalf("cleanup = metadata %#v removed %q", repository.metadata, client.removed)
	}
}

func TestServiceKeepsPendingUploadOnTransientValidatorFailure(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := DirectUploadRequest{Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute}
	if _, err := service.BeginDirectUpload(context.Background(), request); err != nil {
		t.Fatalf("begin: %v", err)
	}
	_, err = service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{
		Name: request.Name, UserId: request.UserId,
		Validator: func(io.Reader, string) error { return errors.New("temporary stream failure") },
	})
	if err == nil {
		t.Fatal("complete succeeded")
	}
	if repository.metadata == nil || repository.metadata.StorageStatus != "pending" || client.removed != "" {
		t.Fatalf("pending state = metadata %#v removed %q", repository.metadata, client.removed)
	}
}

func TestServiceKeepsPendingUploadOnTransientReadFailure(t *testing.T) {
	client := &fakeS3Client{stat: s3ObjectInfo{Size: 6, ContentType: "image/webp"}, openErr: errors.New("temporary read failure")}
	store := newS3Store(client)
	repository := &fakeMetadataRepository{}
	service, err := newService(store, repository)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	request := DirectUploadRequest{Name: "images/direct.webp", ContentType: "image/webp", Size: 6, UserId: 12, ExpiresIn: time.Minute}
	if _, err := service.BeginDirectUpload(context.Background(), request); err != nil {
		t.Fatalf("begin: %v", err)
	}
	_, err = service.CompleteDirectUpload(context.Background(), CompleteDirectUploadRequest{
		Name: request.Name, UserId: request.UserId,
		Validator: func(io.Reader, string) error { return nil },
	})
	if err == nil {
		t.Fatal("complete succeeded")
	}
	if repository.metadata == nil || repository.metadata.StorageStatus != "pending" || client.removed != "" {
		t.Fatalf("pending state = metadata %#v removed %q", repository.metadata, client.removed)
	}
}

func TestS3StoreRejectsBadExpiryAndStatFailure(t *testing.T) {
	client := &fakeS3Client{statErr: errors.New("stat failed")}
	store := newS3Store(client)
	request := DirectUploadRequest{Name: "images/example.webp", ContentType: "image/webp", Size: 6, ExpiresIn: time.Second}
	if _, err := store.PresignUpload(context.Background(), request); err == nil {
		t.Fatal("short expiry succeeded")
	}
	request.ExpiresIn = time.Minute
	if err := store.VerifyUpload(context.Background(), request); err == nil {
		t.Fatal("stat failure verified")
	}
}

func TestNewS3StoreCreatesMinioPresignedPost(t *testing.T) {
	store, err := NewS3Store(S3Config{
		Endpoint:  "https://objects.example.com",
		Bucket:    "forum",
		AccessKey: "access",
		SecretKey: "secret",
		Region:    "us-east-1",
		PathStyle: true,
	})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	upload, err := store.PresignUpload(context.Background(), DirectUploadRequest{
		Name:        "images/example.webp",
		ContentType: "image/webp",
		Size:        6,
		ExpiresIn:   time.Minute,
	})
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	parsed, err := url.Parse(upload.URL)
	if err != nil || parsed.Host != "objects.example.com" {
		t.Fatalf("upload URL = %q, %v", upload.URL, err)
	}
	if upload.Fields["key"] != "images/example.webp" || upload.Fields["Content-Type"] != "image/webp" || upload.Fields["policy"] == "" {
		t.Fatalf("upload fields = %#v", upload.Fields)
	}
	policy, err := base64.StdEncoding.DecodeString(upload.Fields["policy"])
	if err != nil {
		t.Fatalf("decode policy: %v", err)
	}
	for _, constraint := range []string{`["eq","$bucket","forum"]`, `["eq","$key","images/example.webp"]`, `["eq","$Content-Type","image/webp"]`, `["content-length-range", 6, 6]`} {
		if !bytes.Contains(policy, []byte(constraint)) {
			t.Fatalf("policy missing %s: %s", constraint, policy)
		}
	}
}

func TestNewS3StoreRejectsInvalidConfiguration(t *testing.T) {
	tests := []S3Config{
		{},
		{Endpoint: "localhost:9000"},
		{Endpoint: "localhost:9000", Bucket: "INVALID_BUCKET", AccessKey: "access", SecretKey: "secret"},
		{Endpoint: "localhost:9000", Bucket: "forum"},
	}
	for _, config := range tests {
		if _, err := NewS3Store(config); err == nil {
			t.Fatalf("NewS3Store(%#v) succeeded", config)
		}
	}
}

func TestS3StoreRejectsUnsafeObjectNames(t *testing.T) {
	store := newS3Store(&fakeS3Client{})
	for _, name := range []string{"", "/absolute.webp", "../escape.webp", "images/../escape.webp", " images/a.webp"} {
		err := store.Put(context.Background(), WriteRequest{Name: name, ContentType: "image/webp", Size: 1, Body: strings.NewReader("x")})
		if err == nil {
			t.Fatalf("Put(%q) succeeded", name)
		}
	}
}

func TestNormalizeS3Endpoint(t *testing.T) {
	tests := []struct {
		raw        string
		secure     bool
		want       string
		wantSecure bool
		wantErr    bool
	}{
		{raw: "localhost:9000", secure: false, want: "localhost:9000"},
		{raw: "s3.example.com", secure: true, want: "s3.example.com", wantSecure: true},
		{raw: "https://s3.example.com/", want: "s3.example.com", wantSecure: true},
		{raw: "http://localhost:9000", secure: true, want: "localhost:9000"},
		{raw: "ftp://s3.example.com", wantErr: true},
		{raw: "https://s3.example.com/prefix", wantErr: true},
	}
	for _, test := range tests {
		got, gotSecure, err := normalizeS3Endpoint(test.raw, test.secure)
		if (err != nil) != test.wantErr || got != test.want || gotSecure != test.wantSecure {
			t.Fatalf("normalize(%q, %v) = %q, %v, %v", test.raw, test.secure, got, gotSecure, err)
		}
	}
}
