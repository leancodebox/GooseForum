package filestorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

var ErrDirectUploadUnsupported = errors.New("file storage does not support direct uploads")
var ErrDirectUploadOwnerMismatch = errors.New("pending upload does not belong to user")
var ErrDirectUploadInvalidObject = errors.New("uploaded object failed validation")

type DirectUploadRequest struct {
	Name        string
	ContentType string
	Size        int64
	UserId      uint64
	ExpiresIn   time.Duration
}

type DirectUpload struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Fields    map[string]string `json:"fields"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

type DirectUploadSession struct {
	Metadata Metadata
	Upload   DirectUpload
}

type CompleteDirectUploadRequest struct {
	Name      string
	UserId    uint64
	Validator func(io.Reader, string) error
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

func (service *Service) SupportsDirectUpload() bool {
	_, ok := service.writeStore.(DirectUploadStore)
	return ok
}

func (service *Service) CompleteDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) (*Metadata, error) {
	if request.Name == "" || request.UserId == 0 {
		return nil, errors.New("direct upload name and user are required")
	}
	metadata, err := service.repository.GetPending(ctx, request.Name)
	if err != nil {
		ready, readyErr := service.repository.Get(ctx, request.Name)
		if readyErr == nil {
			if ready.UserId != request.UserId {
				return nil, ErrDirectUploadOwnerMismatch
			}
			return ready, nil
		}
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
		if errors.Is(err, ErrDirectUploadInvalidObject) {
			return nil, service.rollbackStore(ctx, configuredStore, request.Name, err)
		}
		return nil, err
	}
	if request.Validator != nil {
		object, err := configuredStore.Open(ctx, request.Name)
		if err != nil {
			return nil, err
		}
		validationErr := request.Validator(object.Body, metadata.ContentType)
		closeErr := object.Body.Close()
		if validationErr != nil {
			if errors.Is(validationErr, ErrDirectUploadInvalidObject) {
				return nil, service.rollbackStore(ctx, configuredStore, request.Name, validationErr)
			}
			return nil, validationErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	return service.repository.MarkReady(ctx, request.Name)
}

func (service *Service) AbortDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) error {
	if request.Name == "" || request.UserId == 0 {
		return errors.New("direct upload name and user are required")
	}
	metadata, err := service.repository.GetPending(ctx, request.Name)
	if err != nil {
		return err
	}
	if metadata.UserId != request.UserId {
		return ErrDirectUploadOwnerMismatch
	}
	store, err := service.storeFor(*metadata)
	if err != nil {
		return err
	}
	return service.rollbackStore(ctx, store, request.Name, nil)
}

func (service *Service) CleanupPending(ctx context.Context, before time.Time, limit int) (int, error) {
	items, err := service.repository.ListPendingBefore(ctx, before, limit)
	if err != nil {
		return 0, err
	}
	removed := 0
	var cleanupErr error
	for _, metadata := range items {
		store, err := service.storeFor(metadata)
		if err == nil {
			err = service.rollbackStore(ctx, store, metadata.Name, nil)
		}
		if err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("cleanup pending file %q: %w", metadata.Name, err))
			continue
		}
		removed++
	}
	return removed, cleanupErr
}

func BeginDirectUpload(ctx context.Context, request DirectUploadRequest) (*DirectUploadSession, error) {
	return current().BeginDirectUpload(ctx, request)
}

func CompleteDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) (*Metadata, error) {
	return current().CompleteDirectUpload(ctx, request)
}

func AbortDirectUpload(ctx context.Context, request CompleteDirectUploadRequest) error {
	return current().AbortDirectUpload(ctx, request)
}

func SupportsDirectUpload() bool {
	return current().SupportsDirectUpload()
}

func CleanupPending(ctx context.Context, before time.Time, limit int) (int, error) {
	return current().CleanupPending(ctx, before, limit)
}
