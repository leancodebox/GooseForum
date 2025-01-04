package articles

import (
	"github.com/leancodebox/goose/queryopt"
	"time"
)

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

func All() (entities []*Entity) {
	builder().Find(&entities)
	return
}

func CantWriteNew(userId uint64, maxCount int64) bool {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Gt(fieldCreatedAt, time.Now().Format("2006-01-02"))).Count(&count)
	return count > maxCount
}

type PageQuery struct {
	Page, PageSize int
	Search         string
	UserId         uint64
	FilterStatus   bool
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
	if q.UserId != 0 {
		b.Where(queryopt.Eq(fieldUserId, q.UserId))
	}
	if q.FilterStatus {
		b.Where(queryopt.Eq(fieldArticleStatus, 1))
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

func IncrementView(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET view_count = view_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

func IncrementReply(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET reply_count = reply_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}
