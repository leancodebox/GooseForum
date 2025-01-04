package reply

import (
	"github.com/leancodebox/goose/queryopt"
)

func Create(entity *Entity) error {
	result := builder().Create(entity)
	return result.Error
}

func Get(id any) (entity Entity) {
	builder().Where(pid, id).First(&entity)
	return
}

//func save(entity *Entity) int64 {
//	result := builder().Save(entity)
//	return result.RowsAffected
//}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

func DeleteEntity(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}

func GetByMaxIdPage(articleId uint64, id uint64, pageSize int) (entities []Entity) {
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Where(queryopt.Gt(pid, id)).Limit(pageSize).Find(&entities)
	return
}
