package applySheet

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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
	}

	return save(entity)
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

type PageQuery struct {
	Page, PageSize int
	Title          string
	Type           int8
	Status         int8
	UserId         uint64
}

func Page[ResType Entity](q PageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	Data     []ResType
} {
	var list []ResType
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}

	db := builder()
	if q.Title != "" {
		db = db.Where(queryopt.Like(fieldTitle, "%"+q.Title+"%"))
	}
	if q.Type > 0 {
		db = db.Where(queryopt.Eq(fieldType, q.Type))
	}
	if q.Status > 0 {
		db = db.Where(queryopt.Eq(fieldStatus, q.Status))
	}
	if q.UserId > 0 {
		db = db.Where(queryopt.Eq(fieldUserId, q.UserId))
	}

	var total int64
	db.Model(&Entity{}).Count(&total)

	db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Order("id desc").Find(&list)

	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []ResType
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
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
