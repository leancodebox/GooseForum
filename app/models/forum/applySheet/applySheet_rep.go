package applySheet

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
	"time"
)

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	} else {
		return save(entity)
	}
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

func CantWriteNew(applyType SheetType, maxCount int64) bool {
	var count int64
	builder().
		Where(queryopt.Lt(fieldType, applyType)).
		Where(queryopt.Gt(fieldCreatedAt, time.Now().Format("2006-01-02"))).Count(&count)
	return count > maxCount
}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}
