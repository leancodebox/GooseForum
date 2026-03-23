package imUserChatConfigs

import (
	"time"

	"gorm.io/gorm"
)

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
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

func GetConfig(userId, peerId uint64) *Entity {
	var entity Entity
	err := builder().Where("user_id = ? AND peer_id = ?", userId, peerId).First(&entity).Error
	if err != nil {
		return nil
	}
	return &entity
}

func GetUserConfigs(userId uint64) []Entity {
	var entities []Entity
	builder().Where("user_id = ? AND is_deleted = 0", userId).Order("updated_at DESC").Find(&entities)
	return entities
}

func IncrUnread(convId, userId uint64) {
	builder().Where("conv_id = ? AND user_id = ?", convId, userId).Updates(map[string]any{
		"unread_count": gorm.Expr("unread_count + ?", 1),
		"updated_at":   time.Now(),
	})
}

func Touch(convId, userId uint64) {
	builder().Where("conv_id = ? AND user_id = ?", convId, userId).Update("updated_at", time.Now())
}

func ClearUnread(convId, userId uint64) {
	builder().Where("conv_id = ? AND user_id = ?", convId, userId).Update("unread_count", 0)
}

func DeleteConfig(convId, userId uint64) {
	builder().Where("conv_id = ? AND user_id = ?", convId, userId).Update("is_deleted", 1)
}
