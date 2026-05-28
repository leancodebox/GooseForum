package mailservice

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
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

var emailProcessor = struct {
	once   sync.Once
	stopCh chan struct{}
	wg     sync.WaitGroup
}{
	stopCh: make(chan struct{}),
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

	if err = taskQueue.Create(queueTask); err != nil {
		slog.Debug("邮件任务写入队列失败", "type", task.Type, "to", task.To, "err", err)
		return err
	}
	slog.Debug("邮件任务写入队列成功", "id", queueTask.Id, "type", task.Type, "to", task.To)
	return nil
}

// StartEmailProcessor starts the background email queue worker.
func StartEmailProcessor() {
	emailProcessor.once.Do(func() {
		closer.Register(StopEmailProcessor)
		emailProcessor.wg.Add(1)
		go func() {
			defer emailProcessor.wg.Done()
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			if !processPendingEmailTasks(emailProcessor.stopCh) {
				return
			}

			for {
				select {
				case <-ticker.C:
					if !processPendingEmailTasks(emailProcessor.stopCh) {
						return
					}
				case <-emailProcessor.stopCh:
					return
				}
			}
		}()
	})
}

// StopEmailProcessor stops the background email queue worker.
func StopEmailProcessor() error {
	select {
	case <-emailProcessor.stopCh:
	default:
		close(emailProcessor.stopCh)
	}
	emailProcessor.wg.Wait()
	return nil
}

func processPendingEmailTasks(stopCh <-chan struct{}) bool {
	for {
		select {
		case <-stopCh:
			return false
		default:
		}

		tasks := taskQueue.GetPendingTasks(BatchSize)
		if len(tasks) == 0 {
			return true
		}
		slog.Debug("邮件队列拉取任务", "count", len(tasks))

		for _, task := range tasks {
			slog.Debug("邮件队列开始处理任务", "id", task.Id, "type", task.Type, "status", task.Status, "retryCount", task.RetryCount)
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
					"id", task.Id,
					"type", emailTask.Type,
					"to", emailTask.To,
					"retryCount", task.RetryCount,
					"error", err,
				)

				if task.RetryCount < MaxRetries {
					taskQueue.IncrementRetryCount(task.Id)
					taskQueue.UpdateStatus(task.Id, taskQueue.StatusRetrying, err)
					select {
					case <-time.After(RetryInterval):
					case <-stopCh:
						return false
					}
					continue
				}

				taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err)
				continue
			}

			taskQueue.UpdateStatus(task.Id, taskQueue.StatusSuccess, nil)
			slog.Info("邮件发送成功",
				"id", task.Id,
				"type", emailTask.Type,
				"to", emailTask.To,
			)
		}
	}
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
