package posts

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
)

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func SaveNoUpdate(entity *Entity) error {
	return builder().Omit("updated_at").Save(entity).Error
}

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func Get(id uint64) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetMaxId() uint64 {
	var entity Entity
	builder().Order(queryopt.Desc("id")).Limit(1).First(&entity)
	return entity.Id
}

func GetByIds(ids []uint64) (entities []*Entity) {
	if len(ids) == 0 {
		return
	}
	builder().Where("id in ?", ids).Find(&entities)
	return
}

func GetMapByIds(ids []uint64) map[uint64]*Entity {
	list := GetByIds(ids)
	result := make(map[uint64]*Entity, len(list))
	for _, item := range list {
		if item != nil {
			result[item.Id] = item
		}
	}
	return result
}

func UpdateProcessStatus(id uint64, processStatus int8) error {
	return builder().Where(queryopt.Eq("id", id)).Update("process_status", processStatus).Error
}

func DeleteEntity(entity *Entity) int64 {
	return builder().Delete(entity).RowsAffected
}

func GetFirstPageByTopicId(topicId uint64) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Limit(20).
		Order(queryopt.Asc("post_no")).
		Order(queryopt.Asc("id")).
		Find(&entities)
	return
}

func GetByTopicPostNoAsc(topicId uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Limit(limit).
		Order(queryopt.Asc("post_no")).
		Order(queryopt.Asc("id")).
		Find(&entities)
	return
}

func GetByTopicPostNoDesc(topicId uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Limit(limit).
		Order(queryopt.Desc("post_no")).
		Order(queryopt.Desc("id")).
		Find(&entities)
	reversePosts(entities)
	return
}

func GetByTopicPostNoAfter(topicId uint64, postNo uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Gt("post_no", postNo)).
		Limit(limit).
		Order(queryopt.Asc("post_no")).
		Order(queryopt.Asc("id")).
		Find(&entities)
	return
}

func GetByTopicPostNoBefore(topicId uint64, postNo uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Lt("post_no", postNo)).
		Limit(limit).
		Order(queryopt.Desc("post_no")).
		Order(queryopt.Desc("id")).
		Find(&entities)
	reversePosts(entities)
	return
}

func GetByTopicPostNoAtOrAfter(topicId uint64, postNo uint64) (entity Entity, ok bool) {
	err := builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Ge("post_no", postNo)).
		Order(queryopt.Asc("post_no")).
		Order(queryopt.Asc("id")).
		First(&entity).Error
	return entity, err == nil
}

func GetByTopicPostNoAtOrBefore(topicId uint64, postNo uint64) (entity Entity, ok bool) {
	err := builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Le("post_no", postNo)).
		Order(queryopt.Desc("post_no")).
		Order(queryopt.Desc("id")).
		First(&entity).Error
	return entity, err == nil
}

func GetLastByTopicID(topicID uint64) (entity Entity, ok bool) {
	err := builder().
		Where(queryopt.Eq("topic_id", topicID)).
		Order(queryopt.Desc("post_no")).
		Order(queryopt.Desc("id")).
		First(&entity).Error
	return entity, err == nil
}

func GetMaxPostNoByTopicId(topicId uint64) uint64 {
	var entity Entity
	err := builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Order(queryopt.Desc("post_no")).
		Order(queryopt.Desc("id")).
		Limit(1).
		First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0
	}
	return entity.PostNo
}

func GetByTopicIdAfter(topicId uint64, id uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Gt("id", id)).
		Limit(limit).
		Order(queryopt.Asc("id")).
		Find(&entities)
	return
}

func GetByTopicIdBefore(topicId uint64, id uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq("topic_id", topicId)).
		Where(queryopt.Lt("id", id)).
		Limit(limit).
		Order(queryopt.Desc("id")).
		Find(&entities)
	reversePosts(entities)
	return
}

func reversePosts(entities []*Entity) {
	for i, j := 0, len(entities)-1; i < j; i, j = i+1, j-1 {
		entities[i], entities[j] = entities[j], entities[i]
	}
}
