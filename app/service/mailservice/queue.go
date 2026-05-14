package mailservice

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
)

const (
	MaxRetries    = 3
	RetryInterval = time.Second * 5
	BatchSize     = 10
)

type EmailTask struct {
	To       string `json:"to"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Type     string `json:"type"`
}

func init() {
	StartEmailProcessor()
}

// AddToQueue stores an email task for background processing.
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

// StartEmailProcessor starts the background email queue worker.
func StartEmailProcessor() {
	go func() {
		for {
			tasks := taskQueue.GetPendingTasks(BatchSize)
			if len(tasks) == 0 {
				time.Sleep(time.Second * 5)
				continue
			}

			for _, task := range tasks {
				if err := taskQueue.UpdateStatus(task.Id, taskQueue.StatusRunning, nil); err != nil {
					slog.Error("更新任务状态失败", "error", err)
					continue
				}

				var emailTask EmailTask
				if err := json.Unmarshal([]byte(task.TaskJson), &emailTask); err != nil {
					slog.Error("解析任务数据失败", "error", err)
					taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err)
					continue
				}

				err := processEmailTask(emailTask)
				if err != nil {
					slog.Error("处理邮件任务失败",
						"type", emailTask.Type,
						"to", emailTask.To,
						"error", err,
					)

					if task.RetryCount < MaxRetries {
						taskQueue.IncrementRetryCount(task.Id)
						taskQueue.UpdateStatus(task.Id, taskQueue.StatusRetrying, err)
						time.Sleep(RetryInterval)
						continue
					}

					taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err)
					continue
				}

				taskQueue.UpdateStatus(task.Id, taskQueue.StatusSuccess, nil)
				slog.Info("邮件发送成功",
					"type", emailTask.Type,
					"to", emailTask.To,
				)
			}
		}
	}()
}

// processEmailTask dispatches an email task by type.
func processEmailTask(task EmailTask) error {
	switch task.Type {
	case "activation":
		return SendActivationEmail(task.To, task.Username, task.Token)
	case "reset_password":
		return SendPasswordResetEmail(task.To, task.Username, task.Token)
	default:
		return fmt.Errorf("未知的邮件类型: %s", task.Type)
	}
}
