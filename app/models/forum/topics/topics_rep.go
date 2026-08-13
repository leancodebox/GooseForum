package topics

import (
	"slices"
	"sort"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
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
	return GetLatestPublishedForAudience(limit, nil, false)
}

func GetLatestPublishedForAudience(limit int, readableCategoryIDs []uint64, filterAudience bool) (entities []*Entity, err error) {
	if filterAudience && len(readableCategoryIDs) == 0 {
		return []*Entity{}, nil
	}
	query := builder().
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0))
	query = applyAudienceFilter(query, readableCategoryIDs, filterAudience)
	err = query.
		Order(queryopt.Desc("updated_at")).
		Order(queryopt.Desc("id")).
		Limit(limit).
		Find(&entities).Error
	return
}

func applyAudienceFilter(query *gorm.DB, readableCategoryIDs []uint64, filterAudience bool) *gorm.DB {
	if !filterAudience {
		return query
	}
	return query.Where("topics.main_category_id IN ?", readableCategoryIDs)
}

func GetLatestPublishedByUserId(userId uint64, limit int) ([]*Entity, error) {
	return GetLatestPublishedByUserIdForAudience(userId, limit, nil, false)
}

func GetLatestPublishedByUserIdForAudience(userId uint64, limit int, readableCategoryIDs []uint64, filterAudience bool) ([]*Entity, error) {
	if filterAudience && len(readableCategoryIDs) == 0 {
		return []*Entity{}, nil
	}
	var entities []*Entity
	query := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0))
	query = applyAudienceFilter(query, readableCategoryIDs, filterAudience)
	err := query.
		Order(queryopt.Desc("updated_at")).
		Order(queryopt.Desc("id")).
		Limit(limit).
		Find(&entities).Error
	return entities, err
}

func GetPublishedByUserBeforeId(userId uint64, beforeId uint64, limit int) ([]*Entity, error) {
	return GetPublishedByUserBeforeIdForAudience(userId, beforeId, limit, nil, false)
}

func GetPublishedByUserBeforeIdForAudience(userId uint64, beforeId uint64, limit int, readableCategoryIDs []uint64, filterAudience bool) ([]*Entity, error) {
	if filterAudience && len(readableCategoryIDs) == 0 {
		return []*Entity{}, nil
	}
	var entities []*Entity
	query := builder().
		Where(queryopt.Eq("user_id", userId)).
		Where(queryopt.Eq("status", 1)).
		Where(queryopt.Eq("process_status", 0))
	if beforeId > 0 {
		query = query.Where(queryopt.Lt("id", beforeId))
	}
	query = applyAudienceFilter(query, readableCategoryIDs, filterAudience)
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
	Page, PageSize      int
	Search              string
	UserId              uint64
	FilterStatus        bool
	CategoryId          uint64
	FilterAudience      bool
	ReadableCategoryIds []uint64
	Sort                string
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

type TopicPage struct {
	Page     int
	PageSize int
	HasNext  bool
	Data     []Entity
}

func Page(q PageQuery) TopicPage {
	return PageWithDB(dbconnect.Connect(), q)
}

func PageWithDB(conn *gorm.DB, q PageQuery) TopicPage {
	var list []Entity
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
	queryLimit := q.PageSize + 1
	if conn == nil || q.FilterAudience && len(q.ReadableCategoryIds) == 0 {
		return TopicPage{Page: q.Page + 1, PageSize: q.PageSize, Data: []Entity{}}
	}
	b := conn.Table(tableName)
	if q.CategoryId != 0 {
		topicIDs, complete, err := topicCategoryIndex.ActiveTopicIDsByCategoryWithDB(conn, q.CategoryId, 256)
		if err == nil && complete {
			return pageSparseCategoryTopics(b, q, topicIDs)
		} else {
			b = b.Joins(
				`JOIN topic_category_index category_idx ON category_idx.topic_id = topics.id AND category_idx.category_id = ? AND category_idx.effective = ?`,
				q.CategoryId, 1,
			)
		}
	}
	b = applyTopicPageFilters(b, q)
	if q.FilterAudience {
		b = applyAudienceFilter(b, q.ReadableCategoryIds, true)
	}
	b = applyPageSort(b, q.Sort)
	b.Limit(queryLimit).Offset(q.PageSize * q.Page).Find(&list)
	hasNext := len(list) > q.PageSize
	if hasNext {
		list = list[:q.PageSize]
	}
	return TopicPage{Page: q.Page + 1, PageSize: q.PageSize, Data: list, HasNext: hasNext}
}

func applyTopicPageFilters(query *gorm.DB, q PageQuery) *gorm.DB {
	if q.Search != "" {
		query = query.Where(queryopt.Like("topics.title", q.Search))
	}
	if q.UserId != 0 {
		query = query.Where(queryopt.Eq("topics.user_id", q.UserId))
	}
	if q.FilterStatus {
		query = query.Where(queryopt.Eq("topics.status", 1))
		query = query.Where(queryopt.Eq("topics.process_status", 0))
	}
	return query
}

func pageSparseCategoryTopics(query *gorm.DB, q PageQuery, topicIDs []uint64) TopicPage {
	if len(topicIDs) == 0 {
		return TopicPage{Page: q.Page + 1, PageSize: q.PageSize, Data: []Entity{}}
	}
	query = query.Where("topics.id IN ?", topicIDs)
	if q.Search != "" {
		// Keep LIKE behavior (including database collation and wildcard semantics)
		// consistent with the regular path while the primary-key candidate set is
		// still bounded by the sparse-category threshold.
		query = query.Where(queryopt.Like("topics.title", q.Search))
	}
	var list []Entity
	query.Find(&list)
	list = slices.DeleteFunc(list, func(topic Entity) bool {
		if q.FilterStatus && (topic.Status != 1 || topic.ProcessStatus != 0) {
			return true
		}
		if q.UserId != 0 && topic.UserId != q.UserId {
			return true
		}
		return false
	})
	if q.FilterAudience {
		readable := make(map[uint64]struct{}, len(q.ReadableCategoryIds))
		for _, categoryID := range q.ReadableCategoryIds {
			readable[categoryID] = struct{}{}
		}
		list = slices.DeleteFunc(list, func(topic Entity) bool {
			_, ok := readable[topic.MainCategoryId]
			return !ok
		})
	}
	sortTopicPageEntities(list, q.Sort)
	start := min(q.Page*q.PageSize, len(list))
	end := min(start+q.PageSize+1, len(list))
	list = list[start:end]
	hasNext := len(list) > q.PageSize
	if hasNext {
		list = list[:q.PageSize]
	}
	return TopicPage{Page: q.Page + 1, PageSize: q.PageSize, Data: list, HasNext: hasNext}
}

func sortTopicPageEntities(list []Entity, sortKey string) {
	sort.Slice(list, func(i, j int) bool {
		left, right := list[i], list[j]
		switch sortKey {
		case "hot":
			if left.ReplyCount != right.ReplyCount {
				return left.ReplyCount > right.ReplyCount
			}
		case "popular":
			if left.ViewCount != right.ViewCount {
				return left.ViewCount > right.ViewCount
			}
		case "new":
			if !left.CreatedAt.Equal(right.CreatedAt) {
				return left.CreatedAt.After(right.CreatedAt)
			}
		default:
			if left.PinWeight != right.PinWeight {
				return left.PinWeight > right.PinWeight
			}
			if !left.UpdatedAt.Equal(right.UpdatedAt) {
				return left.UpdatedAt.After(right.UpdatedAt)
			}
		}
		return left.Id > right.Id
	})
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

func IncrementPostFast(topicId uint64, posters []Poster, lastPostID uint64, lastPostedAt time.Time) error {
	return IncrementPostFastWithDB(builder(), topicId, posters, lastPostID, lastPostedAt)
}

func IncrementPostFastWithDB(db *gorm.DB, topicId uint64, posters []Poster, lastPostID uint64, lastPostedAt time.Time) error {
	return db.Model(&Entity{}).Where("id = ?", topicId).Updates(map[string]any{
		"post_count":  gorm.Expr("post_count + 1"),
		"reply_count": gorm.Expr("reply_count + 1"),
		"posters":     jsonopt.Encode(posters),
		"last_post_id": gorm.Expr(
			"CASE WHEN last_posted_at IS NULL OR last_posted_at < ? OR (last_posted_at = ? AND last_post_id < ?) THEN ? ELSE last_post_id END",
			lastPostedAt, lastPostedAt, lastPostID, lastPostID,
		),
		"last_posted_at": gorm.Expr(
			"CASE WHEN last_posted_at IS NULL OR last_posted_at < ? THEN ? ELSE last_posted_at END",
			lastPostedAt, lastPostedAt,
		),
		"updated_at": time.Now(),
	}).Error
}

func DecrementPostFast(topicId uint64, posters []Poster, lastPostID uint64, lastPostedAt time.Time) error {
	return builder().Where("id = ?", topicId).Updates(map[string]any{
		"post_count":     gorm.Expr("CASE WHEN post_count > 0 THEN post_count - 1 ELSE 0 END"),
		"reply_count":    gorm.Expr("CASE WHEN reply_count > 0 THEN reply_count - 1 ELSE 0 END"),
		"posters":        jsonopt.Encode(posters),
		"last_post_id":   lastPostID,
		"last_posted_at": lastPostedAt,
	}).Error
}

func ReservePostSequence(topicId uint64) (uint64, error) {
	return ReservePostSequenceWithDB(builder(), topicId)
}

func ReservePostSequenceWithDB(db *gorm.DB, topicId uint64) (uint64, error) {
	result := db.Model(&Entity{}).
		Where("id = ?", topicId).
		Update("post_seq", gorm.Expr("post_seq + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	var postSeq uint64
	err := db.Model(&Entity{}).
		Select("post_seq").
		Where("id = ?", topicId).
		Scan(&postSeq).Error
	return postSeq, err
}

func applyPageSort(b *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "hot":
		return b.Order(queryopt.Desc("topics.reply_count")).Order(queryopt.Desc("topics.id"))
	case "popular":
		return b.Order(queryopt.Desc("topics.view_count")).Order(queryopt.Desc("topics.id"))
	case "new":
		return b.Order(queryopt.Desc("topics.created_at")).Order(queryopt.Desc("topics.id"))
	default:
		return b.Order(queryopt.Desc("topics.pin_weight")).Order(queryopt.Desc("topics.updated_at")).Order(queryopt.Desc("topics.id"))
	}
}
