package imUserChatConfigs

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"gorm.io/gorm"
)

const conversationAccessTTL = 2 * time.Minute

var conversationAccessCache = localcache.Cache[bool]{MaxEntries: 4096}

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

func CanAccessConversation(userId, convId uint64) bool {
	if userId == 0 || convId == 0 {
		return false
	}
	return conversationAccessCache.GetOrLoad(conversationAccessCacheKey(userId, convId), func() (bool, error) {
		var entity Entity
		err := builder().Select("id").Where("user_id = ? AND conv_id = ? AND is_deleted = 0", userId, convId).First(&entity).Error
		return err == nil && entity.Id != 0, nil
	}, conversationAccessTTL)
}

func InvalidateConversationAccess(userId, convId uint64) {
	if userId == 0 || convId == 0 {
		return
	}
	conversationAccessCache.Delete(conversationAccessCacheKey(userId, convId))
}

func conversationAccessCacheKey(userId, convId uint64) string {
	return fmt.Sprintf("chat:conversation:access:%d:%d", userId, convId)
}

func GetUserConfigs(userId uint64) []Entity {
	var entities []Entity
	builder().Where("user_id = ? AND is_deleted = 0", userId).Order("updated_at DESC").Find(&entities)
	return entities
}

func HasUnread(userId uint64) bool {
	var entity Entity
	return builder().
		Select("id").
		Where("user_id = ? AND is_deleted = 0 AND unread_count > 0", userId).
		Limit(1).
		First(&entity).Error == nil && entity.Id != 0
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
	InvalidateConversationAccess(userId, convId)
}
