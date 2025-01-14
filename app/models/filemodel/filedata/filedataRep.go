package filedata

import (
	"fmt"

	queryopt "github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
)

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
		return nil, fmt.Errorf("failed to save file")
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
