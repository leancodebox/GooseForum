package topics

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
)

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func Delete(entity *Entity) int64 {
	return builder().Delete(entity).RowsAffected
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func SaveNoUpdate(entity *Entity) error {
	return builder().Omit("updated_at").Save(entity).Error
}

func Get(id uint64) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetSimple(id any) (entity Entity) {
	builder().Where(queryopt.Eq("id", id)).First(&entity)
	return
}

func GetMaxId() uint64 {
	var entity Entity
	builder().Order(queryopt.Desc("id")).Limit(1).First(&entity)
	return entity.Id
}

func QueryById(startId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Gt("id", startId)).Limit(limit).Order(queryopt.Asc("id")).Find(&entities)
	return
}

func GetMapByIds(ids []uint64) map[uint64]Entity {
	var list []Entity
	if len(ids) == 0 {
		return map[uint64]Entity{}
	}
	builder().Where("id in ?", ids).Find(&list)
	result := make(map[uint64]Entity, len(list))
	for _, item := range list {
		result[item.Id] = item
	}
	return result
}

func GetPointerMapByIds(ids []uint64) map[uint64]*Entity {
	valueMap := GetMapByIds(ids)
	result := make(map[uint64]*Entity, len(valueMap))
	for id, item := range valueMap {
		entity := item
		result[id] = &entity
	}
	return result
}

func GetLatestPublished(limit int) (entities []*Entity, err error) {
	err = builder().
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0)).
		Order(queryopt.Desc("updated_at")).
		Order(queryopt.Desc("id")).
		Limit(limit).
		Find(&entities).Error
	return
}

func GetLatestPublishedByUserId(userId uint64, limit int) ([]*Entity, error) {
	var entities []*Entity
	err := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0)).
		Order(queryopt.Desc("updated_at")).
		Order(queryopt.Desc("id")).
		Limit(limit).
		Find(&entities).Error
	return entities, err
}

func GetPublishedByUserBeforeId(userId uint64, beforeId uint64, limit int) ([]*Entity, error) {
	var entities []*Entity
	query := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0))
	if beforeId > 0 {
		query = query.Where(queryopt.Lt("id", beforeId))
	}
	err := query.Order(queryopt.Desc("id")).Limit(limit).Find(&entities).Error
	return entities, err
}

func GetDraftsByUserId(userId uint64, limit int) ([]*Entity, error) {
	var entities []*Entity
	err := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("status", 0)).
		Order(queryopt.Desc("updated_at")).
		Order(queryopt.Desc("id")).
		Limit(limit).
		Find(&entities).Error
	return entities, err
}

func CantWriteNew(userId uint64, maxCount int64) bool {
	var count int64
	builder().Where(queryopt.Eq("user_id", userId)).Where(queryopt.Gt("created_at", time.Now().Format("2006-01-02"))).Count(&count)
	return count > maxCount
}

type PageQuery struct {
	Page, PageSize int
	Search         string
	UserId         uint64
	FilterStatus   bool
	CategoryId     uint64
	Sort           string
}

type AdminPageQuery struct {
	Page, PageSize int
	Search         string
	UserId         uint64
}

type ModerationPageQuery struct {
	Page, PageSize      int
	FilterProcessStatus bool
	ProcessStatus       int8
	CategoryIDs         []uint64
}

func Page(q PageQuery) struct {
	Page     int
	PageSize int
	HasNext  bool
	Data     []Entity
} {
	var list []Entity
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
	queryLimit := q.PageSize + 1
	b := builder()
	if q.Search != "" {
		b.Where(queryopt.Like("title", q.Search))
	}
	if q.UserId != 0 {
		b.Where(queryopt.Eq("user_id", q.UserId))
	}
	if q.FilterStatus {
		b.Where(queryopt.Eq("status", 1))
		b.Where(queryopt.Eq("process_status", 0))
	}
	if q.CategoryId != 0 {
		b.Where(
			`EXISTS (SELECT 1 FROM topic_category_index idx WHERE idx.topic_id = topics.id AND idx.category_id = ? AND idx.effective = ?)`,
			q.CategoryId,
			1,
		)
	}
	applyPageSort(b, q.Sort)
	b.Limit(queryLimit).Offset(q.PageSize * q.Page).Find(&list)
	hasNext := len(list) > q.PageSize
	if hasNext {
		list = list[:q.PageSize]
	}
	return struct {
		Page     int
		PageSize int
		HasNext  bool
		Data     []Entity
	}{Page: q.Page + 1, PageSize: q.PageSize, Data: list, HasNext: hasNext}
}

func PageForAdmin(q AdminPageQuery) struct {
	Page     int
	PageSize int
	HasNext  bool
	Data     []Entity
} {
	var list []Entity
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
	queryLimit := q.PageSize + 1
	b := builder()
	if q.Search != "" {
		b.Where(queryopt.Like("title", q.Search))
	}
	if q.UserId != 0 {
		b.Where(queryopt.Eq("user_id", q.UserId))
	}
	b.Limit(queryLimit).Offset(q.PageSize * q.Page).Order(queryopt.Desc("pin_weight")).Order(queryopt.Desc("updated_at")).Order(queryopt.Desc("id")).Find(&list)
	hasNext := len(list) > q.PageSize
	if hasNext {
		list = list[:q.PageSize]
	}
	return struct {
		Page     int
		PageSize int
		HasNext  bool
		Data     []Entity
	}{Page: q.Page + 1, PageSize: q.PageSize, Data: list, HasNext: hasNext}
}

func PageForModeration(q ModerationPageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	HasNext  bool
	Data     []Entity
} {
	var list []Entity
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
	queryLimit := q.PageSize + 1
	b := builder().Where(queryopt.Eq("status", 1))
	if q.FilterProcessStatus {
		b.Where(queryopt.Eq("process_status", q.ProcessStatus))
	}
	if len(q.CategoryIDs) > 0 {
		b.Where(
			`EXISTS (SELECT 1 FROM topic_category_index idx WHERE idx.topic_id = topics.id AND idx.category_id IN (?) AND idx.effective = ?)`,
			q.CategoryIDs,
			1,
		)
	}
	b.Limit(queryLimit).Offset(q.PageSize * q.Page).Order(queryopt.Desc("updated_at")).Order(queryopt.Desc("id")).Find(&list)
	hasNext := len(list) > q.PageSize
	if hasNext {
		list = list[:q.PageSize]
	}
	total := int64(q.Page*q.PageSize + len(list))
	if hasNext {
		total++
	}
	return struct {
		Page     int
		PageSize int
		Total    int64
		HasNext  bool
		Data     []Entity
	}{Page: q.Page + 1, PageSize: q.PageSize, Data: list, Total: total, HasNext: hasNext}
}

func UpdateProcessStatus(id uint64, processStatus int8) error {
	return builder().Where(queryopt.Eq("id", id)).UpdateColumn("process_status", processStatus).Error
}

func UpdatePinWeight(id uint64, pinWeight int) error {
	return builder().Where(queryopt.Eq("id", id)).Updates(map[string]any{
		"pin_weight": pinWeight,
	}).Error
}

func IncrementLike(entity Entity) int64 {
	return builder().Exec("UPDATE topics SET like_count = like_count + 1 WHERE id = ?", entity.Id).RowsAffected
}

func DecrementLike(entity Entity) int64 {
	return builder().Exec("UPDATE topics SET like_count = like_count - 1 WHERE id = ?", entity.Id).RowsAffected
}

func IncrementViews(counts map[uint64]uint64) error {
	for topicID, count := range counts {
		if topicID == 0 || count == 0 {
			continue
		}
		if err := builder().Exec("UPDATE topics SET view_count = view_count + ? WHERE id = ?", count, topicID).Error; err != nil {
			return err
		}
	}
	return nil
}

func IncrementPostFast(topicId uint64, posters []Poster) error {
	return builder().Where("id = ?", topicId).Updates(map[string]any{
		"post_count":  gorm.Expr("post_count + 1"),
		"reply_count": gorm.Expr("reply_count + 1"),
		"posters":     jsonopt.Encode(posters),
		"updated_at":  time.Now(),
	}).Error
}

func DecrementPostFast(topicId uint64, posters []Poster) error {
	return builder().Where("id = ?", topicId).Updates(map[string]any{
		"post_count":  gorm.Expr("CASE WHEN post_count > 0 THEN post_count - 1 ELSE 0 END"),
		"reply_count": gorm.Expr("CASE WHEN reply_count > 0 THEN reply_count - 1 ELSE 0 END"),
		"posters":     jsonopt.Encode(posters),
	}).Error
}

func ReservePostSequence(topicId uint64) (uint64, error) {
	result := builder().
		Where("id = ?", topicId).
		Update("post_seq", gorm.Expr("post_seq + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	var postSeq uint64
	err := builder().
		Select("post_seq").
		Where("id = ?", topicId).
		Scan(&postSeq).Error
	return postSeq, err
}

func applyPageSort(b *gorm.DB, sort string) {
	switch sort {
	case "hot":
		b.Order(queryopt.Desc("reply_count")).Order(queryopt.Desc("id"))
	case "popular":
		b.Order(queryopt.Desc("view_count")).Order(queryopt.Desc("id"))
	case "new":
		b.Order(queryopt.Desc("created_at")).Order(queryopt.Desc("id"))
	default:
		b.Order(queryopt.Desc("pin_weight")).Order(queryopt.Desc("updated_at")).Order(queryopt.Desc("id"))
	}
}
