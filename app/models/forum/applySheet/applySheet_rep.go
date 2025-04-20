package applySheet

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
	"github.com/spf13/cast"
	"math"
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

type PageQuery struct {
	Page, PageSize int
}

func Page[ResType Entity](q PageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	Data     []ResType
} {
	var list []ResType
	if q.Page > 0 {
		q.Page -= 1
	} else {
		q.Page = 0
	}
	if q.PageSize < 1 {
		q.PageSize = 1
	}
	var total int64
	var bigEntity Entity
	builder().Limit(1).Order(queryopt.Desc(pid)).Find(&bigEntity)
	total = cast.ToInt64(bigEntity.Id)
	lastId := math.MaxInt
	if q.Page > 0 {
		lastId = cast.ToInt(total) - cast.ToInt(q.PageSize*q.Page)
	}
	builder().Where(queryopt.Le(pid, lastId)).Limit(q.PageSize).Order("id desc").Find(&list)

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
