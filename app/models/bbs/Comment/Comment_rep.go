package Comment

import (
	"context"
	"github.com/leancodebox/goose/querymaker"
)

type Rep struct {
	ctx *context.Context
}

func NewRep(ctx *context.Context) Rep {
	return Rep{
		ctx: ctx,
	}
}

func (itself Rep) Save(entity *Comment) int64 {
	result := builder().WithContext(*itself.ctx).Save(entity)
	return result.RowsAffected
}

func Create(entity *Comment) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func Save(entity *Comment) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveAll(entities *[]Comment) int64 {
	result := builder().Save(entities)
	return result.RowsAffected
}

func Delete(entity *Comment) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

func Get(id any) (entity Comment) {
	builder().Where(pid, id).First(entity)
	return
}

func GetBy(field, value string) (entity Comment) {
	builder().Where(field+" = ?", value).First(&entity)
	return
}

func All() (entities []Comment) {
	builder().Find(&entities)
	return
}

func IsExist(field, value string) bool {
	var count int64
	builder().Where(field+" = ?", value).Count(&count)
	return count > 0
}

func GetByMaxIdPage(articleId uint64, id uint64, pageSize int) (entities []Comment) {
	builder().Where(querymaker.Eq(fieldArticleId, articleId)).Where(querymaker.Gt(pid, id)).Limit(pageSize).Find(&entities)
	return
}
