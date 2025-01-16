package articles

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
	"time"
)

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func Save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
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

func GetCount() int64 {
	var count int64
	builder().Count(&count)
	return count
}

func GetMaxId() uint64 {
	var entity Entity
	builder().Order(queryopt.Desc(pid)).Limit(1).First(&entity)
	return entity.Id
}

func GetByUserAndTitle(userId, title any) (entity Entity) {
	builder().Where(queryopt.Eq(fieldTitle, title)).Where(queryopt.Eq(fieldUserId, userId)).First(&entity)
	return
}

func GetByIds(ids []uint64) (entities []*Entity) {
	if len(ids) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, ids)).Find(&entities)
	return
}

func GetMapByIds(ids []uint64) map[uint64]*Entity {
	return collectionopt.Slice2Map(GetByIds(ids), func(v *Entity) uint64 {
		return v.Id
	})
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
	Categories     []int
}

func Page[ResType SmallEntity | Entity](q PageQuery) struct {
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
		b.Where(queryopt.Eq(fieldProcessStatus, 0))
	}
	if q.Categories != nil && len(q.Categories) > 0 {
		b.Joins("left join article_category_rs on articles.id = article_category_rs.article_id")
		b.Where("article_category_rs.article_category_id IN (?)", q.Categories)
		b.Where("article_category_rs.effective = ? ", 1)
		b.Group("articles.id")
	}
	var total int64
	total = 1200
	//b.Count(&total)
	b.Select("articles.*")
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order("articles.updated_at desc").Find(&list)
	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []ResType
	}{Page: q.Page + 1, PageSize: q.PageSize, Data: list, Total: total}
}

func IncrementView(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET view_count = view_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

func IncrementReply(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET reply_count = reply_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

// GetLatestArticles 获取最新的n篇文章
func GetLatestArticles(limit int) ([]SmallEntity, error) {
	var articles []SmallEntity
	b := builder()
	b.Where(queryopt.Eq(fieldArticleStatus, 1))
	b.Where(queryopt.Eq(fieldProcessStatus, 0))
	err := b.
		Order("id desc").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}
