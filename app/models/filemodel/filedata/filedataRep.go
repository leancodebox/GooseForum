package filedata

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"

	"github.com/google/uuid"
)

// 添加支持的图片类型映射
var supportedImageTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
}

// CheckImageType 检查文件类型是否支持，返回对应的 Content-Type
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
	} else {
		return save(entity)
	}
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

// SaveFileFromUpload 处理文件上传的通用方法
func SaveFileFromUpload(userId uint64, fileData []byte, filename string, customPath string) (*Entity, error) {
	// 验证文件大小
	if len(fileData) > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of 2MB")
	}

	// 检查文件类型
	contentType, err := CheckImageType(filename)
	if err != nil {
		return nil, err
	}

	// 生成文件路径
	fileExt := path.Ext(filename)
	newFilename := fmt.Sprintf("%s/%s%s",
		customPath,
		uuid.New().String(),
		fileExt)

	// 保存文件
	return SaveFile(userId, newFilename, contentType, fileData)
}

// 在 supportedImageTypes 映射后添加新的常量
const (
	MaxFileSize = 4 * 1024 * 1024 // 4MB
	AvatarPath  = "avatars"
)

// SaveAvatar 现在可以基于通用方法实现
func SaveAvatar(userId uint64, fileData []byte, filename string) (*Entity, error) {
	avatarPath := fmt.Sprintf("%s/avatar_%d_%d",
		AvatarPath,
		userId,
		time.Now().Unix())

	return SaveFileFromUpload(userId, fileData, filename, avatarPath)
}

// CountUserUploadsInTimeRange 统计用户在指定时间范围内的上传次数
func CountUserUploadsInTimeRange(userId uint64, startTime, endTime time.Time) int64 {
	var count int64
	builder().Where("user_id = ? AND created_at >= ? AND created_at <= ?", userId, startTime, endTime).Count(&count)
	return count
}

// CountUserUploadsToday 统计用户今日上传次数
func CountUserUploadsToday(userId uint64) int64 {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)
	return CountUserUploadsInTimeRange(userId, startOfDay, endOfDay)
}
