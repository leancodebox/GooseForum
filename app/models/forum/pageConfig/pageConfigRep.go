package pageConfig

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
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

func GetByPageType(pageType string) (entity Entity) {
	builder().Where(queryopt.Eq(filedPageType, pageType)).First(&entity)
	return
}
