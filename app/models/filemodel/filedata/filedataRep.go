package filedata

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

type FileResource struct {
	Id        uint64    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	UserId    uint64    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
	Data      []byte    `json:"-"`
}

type FileResourcePageResult struct {
	List     []FileResource
	Page     int
	PageSize int
	Total    int64
}

var supportedImageTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
}

// CheckImageType returns the image MIME type for supported extensions.
func CheckImageType(filename string) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	if contentType, ok := supportedImageTypes[ext]; ok {
		return contentType, nil
	}
	return "", fmt.Errorf("unsupported image type: %s", ext)
}

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func GetByName(name string) (entity Entity) {
	builder().Where(queryopt.Eq(fieldName, name)).First(&entity)
	return
}

func SaveFile(userId uint64, name string, fileType string, data []byte) (*Entity, error) {
	if GetByName(name).Id != 0 {
		return nil, fmt.Errorf("file already exists: %s", name)
	}
	entity := &Entity{
		Name:          name,
		Type:          fileType,
		Data:          data,
		Size:          int64(len(data)),
		StorageDriver: "database",
		StorageStatus: StorageStatusReady,
		UserId:        userId,
	}
	affected := create(entity)
	if affected == 0 {
		return nil, errors.New("failed to save file, possibly duplicate name")
	}
	return entity, nil
}

func GetFileByName(name string) (*Entity, error) {
	return GetFileByNameContext(context.Background(), name)
}

func GetFileByNameContext(ctx context.Context, name string) (*Entity, error) {
	var entity Entity
	err := builder().WithContext(ctx).Where(queryopt.Eq(fieldName, name)).First(&entity).Error
	if err != nil || entity.Id == 0 {
		return nil, errors.New("file not found")
	}
	return &entity, nil
}

func GetFileMetadataByName(name string) (*Entity, error) {
	return GetFileMetadataByNameContext(context.Background(), name)
}

func GetFileMetadataByNameContext(ctx context.Context, name string) (*Entity, error) {
	return getFileMetadataByNameAndStatus(ctx, name, StorageStatusReady)
}

func GetPendingFileMetadataByNameContext(ctx context.Context, name string) (*Entity, error) {
	return getFileMetadataByNameAndStatus(ctx, name, StorageStatusPending)
}

func getFileMetadataByNameAndStatus(ctx context.Context, name string, status string) (*Entity, error) {
	var entity Entity
	err := builder().WithContext(ctx).
		Select("id, name, assert_type, CASE WHEN file_size > 0 THEN file_size ELSE LENGTH(content) END AS file_size, storage_driver, storage_status, user_id, created_at, updated_at").
		Where(queryopt.Eq(fieldName, name)).Where("storage_status = ?", status).First(&entity).Error
	if err != nil || entity.Id == 0 {
		return nil, errors.New("file not found")
	}
	return &entity, nil
}

func DeleteByName(name string) error {
	return DeleteByNameContext(context.Background(), name)
}

func DeleteByNameContext(ctx context.Context, name string) error {
	return builder().WithContext(ctx).Where(queryopt.Eq(fieldName, name)).Delete(&Entity{}).Error
}

func CreateFileMetadata(ctx context.Context, userId uint64, name string, fileType string, size int64, storageDriver string) (*Entity, error) {
	var count int64
	result := builder().WithContext(ctx).Where(queryopt.Eq(fieldName, name)).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}
	if count > 0 {
		return nil, fmt.Errorf("file already exists: %s", name)
	}
	entity := &Entity{Name: name, Type: fileType, Size: size, StorageDriver: storageDriver, StorageStatus: StorageStatusPending, UserId: userId}
	result = builder().WithContext(ctx).Create(entity)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("failed to save file metadata, possibly duplicate name")
	}
	return entity, nil
}

func UpdateFileContent(ctx context.Context, name string, data []byte) error {
	result := builder().WithContext(ctx).Where(queryopt.Eq(fieldName, name)).UpdateColumn("content", data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file metadata not found")
	}
	return nil
}

func MarkFileReady(ctx context.Context, name string) (*Entity, error) {
	result := builder().WithContext(ctx).Where(queryopt.Eq(fieldName, name)).UpdateColumn("storage_status", StorageStatusReady)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("file metadata not found")
	}
	return GetFileMetadataByNameContext(ctx, name)
}

func FileResourcePage(page, pageSize int) FileResourcePageResult {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	var total int64
	builder().Where("storage_status = ?", StorageStatusReady).Count(&total)

	var list []FileResource
	builder().
		Where("storage_status = ?", StorageStatusReady).
		Select("id, name, assert_type AS type, CASE WHEN file_size > 0 THEN file_size ELSE LENGTH(content) END AS size, user_id, created_at").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list)
	for index := range list {
		list[index].URL = list[index].GetAccessPath()
	}
	return FileResourcePageResult{List: list, Page: page, PageSize: pageSize, Total: total}
}

func (itself FileResource) GetAccessPath() string {
	return accessPath(itself.Name)
}

// CountDailyUploads returns the number of files uploaded by a user today.
func CountDailyUploads(userId uint64) int64 {
	return CountUserUploadsToday(userId)
}

// CountUserUploadsInTimeRange counts uploads for a user within a time range.
func CountUserUploadsInTimeRange(userId uint64, startTime, endTime time.Time) int64 {
	var count int64
	builder().Where("user_id = ? AND storage_status = ? AND created_at >= ? AND created_at <= ?", userId, StorageStatusReady, startTime, endTime).Count(&count)
	return count
}

// CountUserUploadsToday counts uploads for a user today.
func CountUserUploadsToday(userId uint64) int64 {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)
	return CountUserUploadsInTimeRange(userId, startOfDay, endOfDay)
}
