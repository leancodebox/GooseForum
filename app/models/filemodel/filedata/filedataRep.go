package filedata

import (
	"fmt"
	"path"
	"strings"

	queryopt "github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
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

func SaveFile(name string, fileType string, data []byte) (*Entity, error) {
	entity := &Entity{
		Name: name,
		Type: fileType,
		Data: data,
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
