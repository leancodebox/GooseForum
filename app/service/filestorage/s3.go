package filestorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

const (
	S3Driver                  = "s3"
	defaultPresignExpiry      = 10 * time.Minute
	maximumPresignExpiry      = time.Hour
	presignedUploadHTTPMethod = "POST"
)

type S3Config struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	Secure       bool
	PathStyle    bool
}

type S3Store struct {
	client s3Client
	now    func() time.Time
}

type s3ObjectInfo struct {
	Size        int64
	ContentType string
}

type s3Client interface {
	PutObject(context.Context, string, io.Reader, int64, string) error
	OpenObject(context.Context, string) (io.ReadCloser, error)
	StatObject(context.Context, string) (s3ObjectInfo, error)
	RemoveObject(context.Context, string) error
	PresignPost(context.Context, string, string, int64, time.Time) (string, map[string]string, error)
}

func NewS3Store(config S3Config) (*S3Store, error) {
	endpoint, secure, err := normalizeS3Endpoint(config.Endpoint, config.Secure)
	if err != nil {
		return nil, err
	}
	bucket := strings.TrimSpace(config.Bucket)
	if bucket == "" {
		return nil, errors.New("s3 bucket is required")
	}
	if err := s3utils.CheckValidBucketNameStrict(bucket); err != nil {
		return nil, fmt.Errorf("invalid s3 bucket: %w", err)
	}
	if strings.TrimSpace(config.AccessKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, errors.New("s3 access key and secret key are required")
	}
	lookup := minio.BucketLookupAuto
	if config.PathStyle {
		lookup = minio.BucketLookupPath
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(config.AccessKey, config.SecretKey, config.SessionToken),
		Secure:       secure,
		Region:       strings.TrimSpace(config.Region),
		BucketLookup: lookup,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}
	return newS3Store(&minioS3Client{client: client, bucket: bucket}), nil
}

func newS3Store(client s3Client) *S3Store {
	return &S3Store{client: client, now: time.Now}
}

func normalizeS3Endpoint(raw string, secure bool) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, errors.New("s3 endpoint is required")
	}
	if !strings.Contains(raw, "://") {
		return strings.TrimSuffix(raw, "/"), secure, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", false, errors.New("invalid s3 endpoint")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false, errors.New("s3 endpoint scheme must be http or https")
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false, errors.New("s3 endpoint must not contain a path, query, or fragment")
	}
	return parsed.Host, parsed.Scheme == "https", nil
}

func (*S3Store) Driver() string { return S3Driver }

func (store *S3Store) Put(ctx context.Context, request WriteRequest) error {
	if err := validateObjectWrite(request.Name, request.ContentType, request.Size); err != nil {
		return err
	}
	return store.client.PutObject(ctx, request.Name, request.Body, request.Size, request.ContentType)
}

func (store *S3Store) Open(ctx context.Context, name string) (*StoredObject, error) {
	info, err := store.client.StatObject(ctx, name)
	if err != nil {
		return nil, err
	}
	body, err := store.client.OpenObject(ctx, name)
	if err != nil {
		return nil, err
	}
	return &StoredObject{Size: info.Size, Body: body}, nil
}

func (store *S3Store) Delete(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("s3 object name is required")
	}
	return store.client.RemoveObject(ctx, name)
}

func (store *S3Store) PresignUpload(ctx context.Context, request DirectUploadRequest) (*DirectUpload, error) {
	if err := validateObjectWrite(request.Name, request.ContentType, request.Size); err != nil {
		return nil, err
	}
	expiresIn := request.ExpiresIn
	if expiresIn == 0 {
		expiresIn = defaultPresignExpiry
	}
	if expiresIn < time.Minute || expiresIn > maximumPresignExpiry {
		return nil, fmt.Errorf("s3 upload expiry must be between %s and %s", time.Minute, maximumPresignExpiry)
	}
	expiresAt := store.now().UTC().Add(expiresIn)
	uploadURL, fields, err := store.client.PresignPost(ctx, request.Name, request.ContentType, request.Size, expiresAt)
	if err != nil {
		return nil, err
	}
	return &DirectUpload{URL: uploadURL, Method: presignedUploadHTTPMethod, Fields: fields, ExpiresAt: expiresAt}, nil
}

func (store *S3Store) VerifyUpload(ctx context.Context, request DirectUploadRequest) error {
	if err := validateObjectWrite(request.Name, request.ContentType, request.Size); err != nil {
		return err
	}
	info, err := store.client.StatObject(ctx, request.Name)
	if err != nil {
		return err
	}
	if info.Size != request.Size {
		return fmt.Errorf("%w: s3 object size mismatch: got %d, want %d", ErrDirectUploadInvalidObject, info.Size, request.Size)
	}
	if !strings.EqualFold(info.ContentType, request.ContentType) {
		return fmt.Errorf("%w: s3 object content type mismatch: got %q, want %q", ErrDirectUploadInvalidObject, info.ContentType, request.ContentType)
	}
	return nil
}

func validateObjectWrite(name, contentType string, size int64) error {
	if name == "" || strings.TrimSpace(name) != name {
		return errors.New("s3 object name is required")
	}
	if strings.HasPrefix(name, "/") || name == ".." || strings.HasPrefix(name, "../") || path.Clean(name) != name {
		return errors.New("s3 object name must be a clean relative key")
	}
	if err := s3utils.CheckValidObjectName(name); err != nil {
		return fmt.Errorf("invalid s3 object name: %w", err)
	}
	if strings.TrimSpace(contentType) == "" {
		return errors.New("s3 content type is required")
	}
	if size <= 0 {
		return errors.New("s3 object size must be positive")
	}
	return nil
}

type minioS3Client struct {
	client *minio.Client
	bucket string
}

func (client *minioS3Client) PutObject(ctx context.Context, name string, body io.Reader, size int64, contentType string) error {
	_, err := client.client.PutObject(ctx, client.bucket, name, body, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (client *minioS3Client) OpenObject(ctx context.Context, name string) (io.ReadCloser, error) {
	return client.client.GetObject(ctx, client.bucket, name, minio.GetObjectOptions{})
}

func (client *minioS3Client) StatObject(ctx context.Context, name string) (s3ObjectInfo, error) {
	info, err := client.client.StatObject(ctx, client.bucket, name, minio.StatObjectOptions{})
	return s3ObjectInfo{Size: info.Size, ContentType: info.ContentType}, err
}

func (client *minioS3Client) RemoveObject(ctx context.Context, name string) error {
	return client.client.RemoveObject(ctx, client.bucket, name, minio.RemoveObjectOptions{})
}

func (client *minioS3Client) PresignPost(ctx context.Context, name, contentType string, size int64, expiresAt time.Time) (string, map[string]string, error) {
	policy := minio.NewPostPolicy()
	setters := []func() error{
		func() error { return policy.SetBucket(client.bucket) },
		func() error { return policy.SetKey(name) },
		func() error { return policy.SetContentType(contentType) },
		func() error { return policy.SetContentLengthRange(size, size) },
		func() error { return policy.SetExpires(expiresAt) },
		func() error { return policy.SetSuccessStatusAction("204") },
	}
	for _, set := range setters {
		if err := set(); err != nil {
			return "", nil, err
		}
	}
	uploadURL, fields, err := client.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", nil, err
	}
	return uploadURL.String(), fields, nil
}

var _ Store = (*S3Store)(nil)
var _ DirectUploadStore = (*S3Store)(nil)
