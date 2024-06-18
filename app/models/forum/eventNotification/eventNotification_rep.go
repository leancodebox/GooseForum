package eventNotification

import "github.com/leancodebox/goose/queryopt"

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

//func save(entity *Entity) int64 {
//	result := builder().Save(entity)
//	return result.RowsAffected
//}

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

func GetByUserId(userId any) (entities []*Entity) {
	b := builder().Where(queryopt.Eq(fieldUserId, userId))
	b.Find(&entities)
	return
}
