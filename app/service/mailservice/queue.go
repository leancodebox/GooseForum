package mailservice

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
)

const (
	MaxRetries    = 3               // 最大重试次数
	RetryInterval = time.Second * 5 // 重试间隔
	BatchSize     = 10              // 每次处理的任务数量
)

type EmailTask struct {
	To       string `json:"to"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Type     string `json:"type"` // activation, reset_password 等
}

func init() {
	// 启动邮件处理器
	StartEmailProcessor()
}

// AddToQueue 添加邮件任务到队列
func AddToQueue(task EmailTask) error {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化邮件任务失败: %v", err)
	}

	queueTask := &taskQueue.Entity{
		Type:     task.Type,
		Status:   taskQueue.StatusPending,
		TaskJson: string(taskJson),
	}

	return taskQueue.Create(queueTask)
}

// StartEmailProcessor 启动邮件处理器
func StartEmailProcessor() {
	go func() {
		for {
			// 获取待处理的任务
			tasks := taskQueue.GetPendingTasks(BatchSize)
			if len(tasks) == 0 {
				time.Sleep(time.Second * 5)
				continue
			}

			for _, task := range tasks {
				// 更新任务状态为处理中
				if err := taskQueue.UpdateStatus(task.Id, taskQueue.StatusRunning, nil); err != nil {
					slog.Error("更新任务状态失败", "error", err)
					continue
				}

				// 解析任务数据
				var emailTask EmailTask
				if err := json.Unmarshal([]byte(task.TaskJson), &emailTask); err != nil {
					slog.Error("解析任务数据失败", "error", err)
					taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err)
					continue
				}

				// 处理任务
				err := processEmailTask(emailTask)
				if err != nil {
					slog.Error("处理邮件任务失败",
						"type", emailTask.Type,
						"to", emailTask.To,
						"error", err,
					)

					// 处理重试逻辑
					if task.RetryCount < MaxRetries {
						taskQueue.IncrementRetryCount(task.Id)
						taskQueue.UpdateStatus(task.Id, taskQueue.StatusRetrying, err)
						time.Sleep(RetryInterval)
						continue
					}

					taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err)
					continue
				}

				// 更新任务状态为成功
				taskQueue.UpdateStatus(task.Id, taskQueue.StatusSuccess, nil)
				slog.Info("邮件发送成功",
					"type", emailTask.Type,
					"to", emailTask.To,
				)
			}
		}
	}()
}

// processEmailTask 处理邮件任务
func processEmailTask(task EmailTask) error {
	switch task.Type {
	case "activation":
		return SendActivationEmail(task.To, task.Username, task.Token)
	default:
		return fmt.Errorf("未知的邮件类型: %s", task.Type)
	}
}
