package filestorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

const DatabaseDriver = "database"

type databaseStore struct{}

func NewDatabaseStore() Store {
	return databaseStore{}
}

func (databaseStore) Driver() string {
	return DatabaseDriver
}

func (databaseStore) Put(ctx context.Context, request WriteRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := io.ReadAll(io.LimitReader(request.Body, request.Size+1))
	if err != nil {
		return err
	}
	if int64(len(data)) != request.Size {
		return fmt.Errorf("file storage size mismatch: got %d, want %d", len(data), request.Size)
	}
	return filedata.UpdateFileContent(ctx, request.Name, data)
}

func (databaseStore) Open(ctx context.Context, name string) (*StoredObject, error) {
	entity, err := filedata.GetFileByNameContext(ctx, name)
	if err != nil {
		return nil, err
	}
	return &StoredObject{
		Size: int64(len(entity.Data)),
		Body: io.NopCloser(bytes.NewReader(entity.Data)),
	}, nil
}

func (databaseStore) Delete(ctx context.Context, _ string) error {
	// Database content and metadata share one row; repository deletion removes both atomically.
	return ctx.Err()
}

type databaseMetadataRepository struct{}

func (databaseMetadataRepository) Create(ctx context.Context, metadata Metadata) (*Metadata, error) {
	entity, err := filedata.CreateFileMetadata(ctx, metadata.UserId, metadata.Name, metadata.ContentType, metadata.Size, metadata.StorageDriver)
	if err != nil {
		return nil, err
	}
	result := metadataFromEntity(entity)
	return &result, nil
}

func (databaseMetadataRepository) Get(ctx context.Context, name string) (*Metadata, error) {
	entity, err := filedata.GetFileMetadataByNameContext(ctx, name)
	if err != nil {
		return nil, err
	}
	result := metadataFromEntity(entity)
	return &result, nil
}

func (databaseMetadataRepository) MarkReady(ctx context.Context, name string) (*Metadata, error) {
	entity, err := filedata.MarkFileReady(ctx, name)
	if err != nil {
		return nil, err
	}
	result := metadataFromEntity(entity)
	return &result, nil
}

func (databaseMetadataRepository) Delete(ctx context.Context, name string) error {
	return filedata.DeleteByNameContext(ctx, name)
}

func metadataFromEntity(entity *filedata.Entity) Metadata {
	storageDriver := entity.StorageDriver
	if storageDriver == "" {
		storageDriver = DatabaseDriver
	}
	storageStatus := entity.StorageStatus
	if storageStatus == "" {
		storageStatus = filedata.StorageStatusReady
	}
	return Metadata{
		Id:            entity.Id,
		Name:          entity.Name,
		ContentType:   entity.Type,
		Size:          entity.Size,
		UserId:        entity.UserId,
		StorageDriver: storageDriver,
		StorageStatus: storageStatus,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}
