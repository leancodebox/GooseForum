package role

import "github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"

func create(entity *Entity) error {
	result := builder().Create(entity)
	return result.Error
}

func save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
}

func SaveOrCreateById(entity *Entity) error {
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

func GetByRoleIds(roleIds []uint64) (entities []*Entity) {
	builder().Where(queryopt.In(pid, roleIds)).Find(&entities)
	return
}

func AllEffective() (entities []*Entity) {
	builder().Find(&entities)
	return
}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

func DeleteEntity(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

type PageQuery struct {
	Page, PageSize int
	RoleName       string
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
	if q.RoleName != "" {
		b.Where(queryopt.Like(fieldRoleName, q.RoleName))
	}
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order(queryopt.Desc(pid)).Find(&list)
	var total int64
	b.Count(&total)
	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []Entity
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
}
