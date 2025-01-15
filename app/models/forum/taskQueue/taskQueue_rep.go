package taskQueue

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
)

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

// GetPendingTasks 获取待处理的任务
func GetPendingTasks(limit int) (tasks []*Entity) {
	builder().Where(queryopt.Eq("status", StatusPending)).
		Order("id asc").
		Limit(limit).
		Find(&tasks)
	return
}

// UpdateStatus 更新任务状态
func UpdateStatus(id uint64, status uint8, err error) error {
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": time.Now(),
	}
	if err != nil {
		updates["last_error"] = err.Error()
	}
	return builder().Where("id = ?", id).Updates(updates).Error
}

// IncrementRetryCount 增加重试次数
func IncrementRetryCount(id uint64) error {
	return builder().Exec("UPDATE task_queue SET retry_count = retry_count + 1 where id = ?", id).Error
}
