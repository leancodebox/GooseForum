package imConversations

import (
	"time"

	"github.com/samber/lo"
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

func UpdateLastMsg(id uint64, content string) {
	builder().Where("id = ?", id).Updates(map[string]any{
		"last_msg_content": content,
		"last_msg_time":    time.Now(),
	})
}

func GetByIds(ids []uint64) (entities []*Entity) {
	if len(ids) == 0 {
		return
	}
	builder().Where("id IN ?", ids).Find(&entities)
	return
}

func GetMapByIds(ids []uint64) map[uint64]*Entity {
	return lo.KeyBy(GetByIds(ids), func(v *Entity) uint64 {
		return v.Id
	})
}
