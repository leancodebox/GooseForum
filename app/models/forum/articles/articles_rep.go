package articles

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

func Create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func Delete(entity *Entity) int64 {
	result := builder().Delete(entity)
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

func Get(id any) (entity Entity) {
	builder().Where(queryopt.Eq(pid, id)).First(&entity)
	return
}

func GetSimple(id any) (entity SmallEntity) {
	builder().Where(queryopt.Eq(pid, id)).First(&entity)
	return
}

func GetMaxId() uint64 {
	var entity Entity
	builder().Order(queryopt.Desc(pid)).Limit(1).First(&entity)
	return entity.Id
}

func GetByIds(ids []uint64) (entities []*SmallEntity) {
	if len(ids) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, ids)).Find(&entities)
	return
}

// GetAllSimple returns paginated simple entities for export and repair jobs.
func GetAllSimple(offset, limit int) ([]*SmallEntity, error) {
	var entities []*SmallEntity
	err := builder().Offset(offset).Limit(limit).Order("id ASC").Find(&entities).Error
	return entities, err
}

// GetCountGroupByDay groups article counts by day.
func GetCountGroupByDay() ([]map[string]any, error) {
	var results []map[string]any
	err := builder().Select("DATE(created_at) as date, count(*) as count").Group("date").Order("date ASC").Find(&results).Error
	return results, err
}

func GetMapByIds(ids []uint64) map[uint64]*SmallEntity {
	return lo.KeyBy(GetByIds(ids), func(v *SmallEntity) uint64 {
		return v.Id
	})
}

func GetLast(limit int) (entities []*Entity) {
	builder().Order(queryopt.Desc(pid)).Limit(limit).Find(&entities)
	return
}

func GetBatch(minId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Gt(pid, minId)).Order(queryopt.Asc(pid)).Limit(limit).Find(&entities)
	return
}

func UpdatePosters(id uint64, posters []Poster) error {
	return builder().Where(pid, id).Select("Posters").Updates(Entity{Posters: posters}).Error
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
	Sort           string
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
		if len(q.Categories) == 1 {
			b.Where(`EXISTS (SELECT 1 FROM article_category_rs rs 
WHERE rs.article_id = articles.id AND rs.article_category_id = ?  AND rs.effective = ? )`,
				q.Categories[0], 1)
		} else {
			b.Where(`EXISTS (SELECT 1 FROM article_category_rs rs 
WHERE rs.article_id = articles.id AND rs.article_category_id IN (?) AND rs.effective = ? )`,
				q.Categories, 1)
		}
	}

	if q.Sort == "new" {
		b.Order(queryopt.Desc(pid))
	} else {
		b.Order(queryopt.Desc(fieldUpdatedAt))
	}

	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Find(&list)
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

func IncrementLike(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET like_count = like_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

func DecrementLike(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET like_count = like_count-1 where id = ?", entity.Id)
	return result.RowsAffected
}

func IncrementView(entity Entity) int64 {
	result := builder().Exec("UPDATE articles SET view_count = view_count+1 where id = ?", entity.Id)
	return result.RowsAffected
}

func IncrementReplyFast(articleId uint64, posters []Poster) error {
	return builder().Where("id = ?", articleId).Updates(map[string]any{
		"reply_count": gorm.Expr("reply_count + 1"),
		"posters":     jsonopt.Encode(posters),
		"updated_at":  time.Now(),
	}).Error
}

func DecrementReplyFast(articleId uint64, posters []Poster) error {
	return builder().Where("id = ?", articleId).Updates(map[string]any{
		"reply_count": gorm.Expr("reply_count - 1"),
		"posters":     jsonopt.Encode(posters),
	}).Error
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
		Order(queryopt.Desc(pid)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func GetLatestArticlesWithContent(limit int) ([]Entity, error) {
	var articles []Entity
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

func GetRecommendedArticlesByAuthorId(authorId uint64, limit int) ([]*SmallEntity, error) {
	var articles []*SmallEntity
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

func GetLatestArticlesByUserId(userId uint64, limit int) ([]*SmallEntity, error) {
	var articles []*SmallEntity
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
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where("deleted_at IS NULL").Count(&count)
	return count
}

// QueryById 根据ID批量查询文章
func QueryById(startId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Gt(pid, startId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}
