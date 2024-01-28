package Articles

import (
	"context"
	"github.com/leancodebox/goose/queryopt"
	"time"
)

type Rep struct {
	ctx *context.Context
}

func NewRep(ctx *context.Context) Rep {
	return Rep{
		ctx: ctx,
	}
}

func (itself Rep) Save(entity *Entity) int64 {
	result := builder().WithContext(*itself.ctx).Save(entity)
	return result.RowsAffected
}

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func Save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveAll(entities *[]Entity) int64 {
	result := builder().Save(entities)
	return result.RowsAffected
}

func Delete(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

func Get(id any) (entity Entity) {
	builder().Where(queryopt.Eq(pid, id)).First(&entity)
	return
}

func GetBy(field, value string) (entity Entity) {
	builder().Where(field+" = ?", value).First(&entity)
	return
}

func All() (entities []Entity) {
	builder().Find(&entities)
	return
}

func IsExist(field, value string) bool {
	var count int64
	builder().Where(field+" = ?", value).Count(&count)
	return count > 0
}

func GetByMaxIdPage(id uint64, pageSize int) (entities []*Entity) {
	builder().Where(queryopt.Gt(pid, id)).Order(queryopt.Desc(fieldUpdateTime)).Limit(pageSize).Find(&entities)
	return
}

func CantWriteNew(userId uint64, maxCount int64) bool {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Gt(fieldCreateTime, time.Now().Format("2006-01-02"))).Count(&count)
	return count > maxCount
}

type PageQuery struct {
	Page, PageSize int
	Search         string
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
	if q.Search != "" {
		b.Where(queryopt.Like(fieldContent, q.Search))
	}
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order("id desc").Find(&list)

	var total int64
	if q.Search != "" {
		builder().Where(queryopt.Like(fieldContent, q.Search)).Count(&total)
	} else {
		builder().Count(&total)
	}
	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []Entity
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
}
