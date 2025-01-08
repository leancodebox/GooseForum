package optRecord

import "github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return Create(entity)
	} else {
		return save(entity)
	}
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
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

type PageQuery struct {
	Page, PageSize int
	OptUserId      uint64
	OptType        int
	TargetType     int
	TargetId       int
}

func Page(q PageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	Data     []Entity
} {
	var list []Entity
	if q.Page > 0 {
		q.Page -= 1
	} else {
		q.Page = 0
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
	b := builder()
	if q.OptUserId != 0 {
		b.Where(queryopt.Eq(fieldOptUserId, q.OptUserId))
	}
	if q.OptType != 0 {
		b.Where(queryopt.Eq(fieldOptType, q.OptType))
	}
	if q.TargetType != 0 {
		b.Where(queryopt.Eq(fieldTargetType, q.TargetType))
	}

	if q.TargetId != 0 {
		b.Where(queryopt.Eq(fieldTargetId, q.TargetId))
	}
	var total int64
	b.Count(&total)
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order("id desc").Find(&list)
	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []Entity
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
}
