package filedata

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"

	"github.com/google/uuid"
)

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

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func CreateOrSave(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	}

	return save(entity)
}

func Get(id any) (entity Entity) {
	builder().Where(fmt.Sprintf(`%v = ?`, pid), id).First(&entity)
	return
}

func GetByName(name string) (entity Entity) {
	builder().Where(queryopt.Eq(fieldName, name)).First(&entity)
	return
}

func all() (entities []*Entity) {
	builder().Find(&entities)
	return
}

func SaveFile(userId uint64, name string, fileType string, data []byte) (*Entity, error) {
	entity := &Entity{
		Name:   name,
		Type:   fileType,
		Data:   data,
		UserId: userId,
	}
	affected := CreateOrSave(entity)
	if affected == 0 {
		return nil, fmt.Errorf("failed to save file, possibly duplicate name")
	}
	return entity, nil
}

func GetFile(id uint64) (*Entity, error) {
	entity := Get(id)
	if entity.Id == 0 {
		return nil, fmt.Errorf("file not found")
	}
	return &entity, nil
}

func GetFileByName(name string) (*Entity, error) {
	entity := GetByName(name)
	if entity.Id == 0 {
		return nil, fmt.Errorf("file not found")
	}
	return &entity, nil
}

// CountDailyUploads returns the number of files uploaded by a user today.
func CountDailyUploads(userId uint64) int64 {
	return CountUserUploadsToday(userId)
}

func SaveFileFromUpload(userId uint64, fileData []byte, filename string, customPath string) (*Entity, error) {
	if len(fileData) > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %dMB", MaxFileSize/(1024*1024))
	}

	contentType, err := CheckImageType(filename)
	if err != nil {
		return nil, err
	}

	fileExt := path.Ext(filename)
	newFilename := fmt.Sprintf("%s/%s%s",
		customPath,
		uuid.New().String(),
		fileExt)

	return SaveFile(userId, newFilename, contentType, fileData)
}

const (
	MaxFileSize = 4 * 1024 * 1024 // 4MB
	AvatarPath  = "avatars"
)

// SaveAvatar stores an uploaded avatar file.
func SaveAvatar(userId uint64, fileData []byte, filename string) (*Entity, error) {
	avatarPath := fmt.Sprintf("%s/avatar_%d_%d",
		AvatarPath,
		userId,
		time.Now().Unix())

	return SaveFileFromUpload(userId, fileData, filename, avatarPath)
}

// CountUserUploadsInTimeRange counts uploads for a user within a time range.
func CountUserUploadsInTimeRange(userId uint64, startTime, endTime time.Time) int64 {
	var count int64
	builder().Where("user_id = ? AND created_at >= ? AND created_at <= ?", userId, startTime, endTime).Count(&count)
	return count
}

// CountUserUploadsToday counts uploads for a user today.
func CountUserUploadsToday(userId uint64) int64 {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)
	return CountUserUploadsInTimeRange(userId, startOfDay, endOfDay)
}
