package taskQueue

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

// GetPendingTasks 获取待处理的任务
func GetPendingTasks(limit int) (tasks []*Entity) {
	builder().Where(queryopt.In("status", []int{StatusPending, StatusRetrying})).
		Order("id asc").
		Limit(limit).
		Find(&tasks)
	return
}

// Claim atomically moves a pending or retrying task to running. The boolean
// distinguishes a stale task read by another worker from a database error.
func Claim(id uint64) (bool, error) {
	result := builder().Where("id = ? AND status IN ?", id, []int{StatusPending, StatusRetrying}).Updates(map[string]any{
		"status":       StatusRunning,
		"processed_at": time.Now(),
	})
	return result.RowsAffected == 1, result.Error
}

// UpdateStatus 更新任务状态
func UpdateStatus(id uint64, status uint8, err error) error {
	updates := map[string]any{
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

// Retry resets a failed task so an authorized caller can explicitly requeue it.
func Retry(id uint64, taskType string) (bool, error) {
	result := builder().Where("id = ? AND type = ? AND status = ?", id, taskType, StatusFailed).Updates(map[string]any{
		"status":       StatusPending,
		"retry_count":  0,
		"last_error":   "",
		"processed_at": time.Time{},
	})
	return result.RowsAffected == 1, result.Error
}
