package articles

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/spf13/cast"
)

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func Save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
}
func SaveNoUpdate(entity *Entity) error {
	result := builder().Omit(fieldUpdatedAt).Save(entity)
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
func GetMonthCount() int64 {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var count int64
	builder().Where(queryopt.Ge(fieldCreatedAt, firstOfMonth)).Count(&count)
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

func GetByIds(ids []uint64) (entities []*SmallEntity) {
	if len(ids) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, ids)).Find(&entities)
	return
}

func GetMapByIds(ids []uint64) map[uint64]*SmallEntity {
	return collectionopt.Slice2Map(GetByIds(ids), func(v *SmallEntity) uint64 {
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

func Page[ResType SmallEntity](q PageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	Data     []ResType
} {
	var list []ResType
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
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
	if len(q.Categories) > 0 {
		b.Joins("left join article_category_rs on articles.id = article_category_rs.article_id")
		b.Where("article_category_rs.article_category_id IN (?)", q.Categories)
		b.Where("article_category_rs.effective = ? ", 1)
		b.Group("articles.id")
	}
	//b.Count(&total)
	b.Select(" articles.id, articles.title,articles.description, articles.type, articles.user_id, articles.article_status, articles.process_status," +
		" articles.view_count, articles.reply_count, articles.like_count, articles.created_at, articles.updated_at, articles.deleted_at")
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order(" articles.updated_at desc").Find(&list)
	var total int64
	total = 1200
	if len(list) < q.PageSize {
		total = cast.ToInt64(q.Page*q.PageSize + len(list))
	}
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

func IncrementLike(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET like_count = like_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

func DecrementLike(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET like_count = like_count-1 where id = ?", entity.Id)
	return result.RowsAffected
}

// GetLatestArticles 获取最新的n篇文章
func GetLatestArticles(limit int) ([]SmallEntity, error) {
	var articles []SmallEntity
	b := builder()
	b.Where(queryopt.Eq(fieldArticleStatus, 1))
	b.Where(queryopt.Eq(fieldProcessStatus, 0))
	err := b.
		Order(queryopt.Desc(pid)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func GetRecommendedArticles(limit int) ([]SmallEntity, error) {
	var articles []SmallEntity
	b := builder()
	b.Where(queryopt.Eq(fieldArticleStatus, 1))
	b.Where(queryopt.Eq(fieldProcessStatus, 0))
	err := b.
		Order(queryopt.Desc(fieldReplyCount)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func GetRecommendedArticlesByAuthorId(authorId uint64, limit int) ([]SmallEntity, error) {
	var articles []SmallEntity
	b := builder()
	b.Where(queryopt.Eq(fieldUserId, authorId))
	b.Where(queryopt.Eq(fieldArticleStatus, 1))
	b.Where(queryopt.Eq(fieldProcessStatus, 0))
	err := b.
		Order(queryopt.Desc(fieldReplyCount)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func GetLatestArticlesByUserId(userId uint64, limit int) ([]SmallEntity, error) {
	var articles []SmallEntity
	b := builder()
	b.Where(queryopt.Eq(fieldArticleStatus, 1))
	b.Where(queryopt.Eq(fieldProcessStatus, 0))
	b.Where(queryopt.Eq(fieldUserId, userId))
	err := b.
		Order("id desc").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func GetUserCount(userId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Count(&count)
	return count
}

// QueryById 根据ID批量查询文章
func QueryById(startId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Gt(pid, startId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}
