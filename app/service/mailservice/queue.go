package mailservice

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	paniclog "github.com/leancodebox/GooseForum/app/bundles/recovery"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"gorm.io/gorm"
)

const (
	MaxRetries    = 3
	RetryInterval = time.Second * 5
	BatchSize     = 10

	EmailTypeActivation             = "activation"
	EmailTypePasswordReset          = "reset_password"
	EmailTypeRegistrationInvitation = "registration_invitation"
)

type EmailTask struct {
	To       string `json:"to"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Type     string `json:"type"`
	Locale   string `json:"locale,omitempty"`
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
	_, err := AddToQueueWithDB(dbconnect.Connect(), task)
	return err
}

// AddToQueueWithDB stores an email task in the supplied transaction or connection.
func AddToQueueWithDB(db *gorm.DB, task EmailTask) (uint64, error) {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return 0, fmt.Errorf("序列化邮件任务失败: %w", err)
	}

	queueTask := &taskQueue.Entity{
		Type:     task.Type,
		Status:   taskQueue.StatusPending,
		TaskJson: string(taskJson),
	}

	if db == nil {
		return 0, fmt.Errorf("邮件任务写入队列失败: 数据库连接为空")
	}
	if err = db.Create(queueTask).Error; err != nil {
		slog.Debug("邮件任务写入队列失败", "type", task.Type, "to", task.To, "err", err)
		return 0, err
	}
	slog.Debug("邮件任务写入队列成功", "id", queueTask.Id, "type", task.Type, "to", task.To)
	return queueTask.Id, nil
}

// StartEmailProcessor starts the background email queue worker.
func StartEmailProcessor() {
	emailProcessor.once.Do(func() {
		closer.RegisterPriority(closer.PriorityProducer, StopEmailProcessor)
		emailProcessor.wg.Go(func() {
			defer paniclog.Recover("mail_processor")
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
		})
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
			claimed, err := taskQueue.Claim(task.Id)
			if err != nil {
				slog.Error("更新任务状态失败", "error", err)
				continue
			}
			if !claimed {
				slog.Debug("邮件任务已被其他 worker 领取", "id", task.Id)
				continue
			}

			var emailTask EmailTask
			if err := json.Unmarshal([]byte(task.TaskJson), &emailTask); err != nil {
				slog.Error("解析任务数据失败", "error", err)
				if updateErr := taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err); updateErr != nil {
					slog.Error("更新任务状态失败", "id", task.Id, "error", updateErr)
				}
				continue
			}

			err = processEmailTask(emailTask)
			if err != nil {
				slog.Error("处理邮件任务失败",
					"id", task.Id,
					"type", emailTask.Type,
					"to", emailTask.To,
					"retryCount", task.RetryCount,
					"error", err,
				)

				if task.RetryCount < MaxRetries {
					if updateErr := taskQueue.IncrementRetryCount(task.Id); updateErr != nil {
						slog.Error("更新任务重试次数失败", "id", task.Id, "error", updateErr)
					}
					if updateErr := taskQueue.UpdateStatus(task.Id, taskQueue.StatusRetrying, err); updateErr != nil {
						slog.Error("更新任务状态失败", "id", task.Id, "error", updateErr)
					}
					select {
					case <-time.After(RetryInterval):
					case <-stopCh:
						return false
					}
					continue
				}

				if updateErr := taskQueue.UpdateStatus(task.Id, taskQueue.StatusFailed, err); updateErr != nil {
					slog.Error("更新任务状态失败", "id", task.Id, "error", updateErr)
				}
				continue
			}

			if err := taskQueue.UpdateStatus(task.Id, taskQueue.StatusSuccess, nil); err != nil {
				slog.Error("更新任务状态失败", "id", task.Id, "error", err)
				continue
			}
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
	case EmailTypeActivation:
		return SendActivationEmail(task.To, task.Username, task.Token, task.Locale)
	case EmailTypePasswordReset:
		return SendPasswordResetEmail(task.To, task.Username, task.Token, task.Locale)
	case EmailTypeRegistrationInvitation:
		return SendRegistrationInvitationEmail(task.To, task.Token, task.Locale)
	default:
		return fmt.Errorf("未知的邮件类型: %s", task.Type)
	}
}
