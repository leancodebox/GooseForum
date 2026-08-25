package filestorage

import (
	"context"
	"errors"
	"time"
)

var ErrDirectUploadUnsupported = errors.New("file storage does not support direct uploads")
var ErrDirectUploadOwnerMismatch = errors.New("pending upload does not belong to user")

type DirectUploadRequest struct {
	Name        string
	ContentType string
	Size        int64
	UserId      uint64
	ExpiresIn   time.Duration
}

type DirectUpload struct {
	URL       string
	Method    string
	Fields    map[string]string
	ExpiresAt time.Time
}

type DirectUploadSession struct {
	Metadata Metadata
	Upload   DirectUpload
}

type CompleteDirectUploadRequest struct {
	Name   string
	UserId uint64
}

// DirectUploadStore is an optional capability implemented by remote stores.
type DirectUploadStore interface {
	PresignUpload(context.Context, DirectUploadRequest) (*DirectUpload, error)
	VerifyUpload(context.Context, DirectUploadRequest) error
}

func (service *Service) BeginDirectUpload(ctx context.Context, request DirectUploadRequest) (*DirectUploadSession, error) {
	store, ok := service.writeStore.(DirectUploadStore)
	if !ok {
		return nil, ErrDirectUploadUnsupported
	}
	if request.UserId == 0 {
		return nil, errors.New("direct upload user is required")
	}
	if err := validateMetadataRequest(request.Name, request.ContentType, request.Size); err != nil {
		return nil, err
	}
	metadata, err := service.repository.Create(ctx, Metadata{
		Name:          request.Name,
		ContentType:   request.ContentType,
		Size:          request.Size,
		UserId:        request.UserId,
		StorageDriver: service.writeStore.Driver(),
	})
	if err != nil {
		return nil, err
	}
	upload, err := store.PresignUpload(ctx, request)
	if err != nil {
		return nil, service.rollbackPut(ctx, request.Name, err)
	}
	return &DirectUploadSession{Metadata: *metadata, Upload: *upload}, nil
}

func (service *Service) CompleteDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) (*Metadata, error) {
	if request.Name == "" || request.UserId == 0 {
		return nil, errors.New("direct upload name and user are required")
	}
	metadata, err := service.repository.GetPending(ctx, request.Name)
	if err != nil {
		return nil, err
	}
	if metadata.UserId != request.UserId {
		return nil, ErrDirectUploadOwnerMismatch
	}
	configuredStore, err := service.storeFor(*metadata)
	if err != nil {
		return nil, err
	}
	store, ok := configuredStore.(DirectUploadStore)
	if !ok {
		return nil, ErrDirectUploadUnsupported
	}
	verifyRequest := DirectUploadRequest{Name: metadata.Name, ContentType: metadata.ContentType, Size: metadata.Size, UserId: metadata.UserId}
	if err := store.VerifyUpload(ctx, verifyRequest); err != nil {
		return nil, err
	}
	return service.repository.MarkReady(ctx, request.Name)
}

func BeginDirectUpload(ctx context.Context, request DirectUploadRequest) (*DirectUploadSession, error) {
	return current().BeginDirectUpload(ctx, request)
}

func CompleteDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) (*Metadata, error) {
	return current().CompleteDirectUpload(ctx, request)
}
