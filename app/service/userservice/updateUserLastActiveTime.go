package userservice

import (
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

// UserActivityTask 用户活跃时间更新任务
type UserActivityTask struct {
	UserID         uint64    // 用户ID
	LastActiveTime time.Time // 最后活跃时间
	CreatedAt      time.Time // 任务创建时间
}

// UserActivityManager 用户活跃时间管理器
type UserActivityManager struct {
	tasks     map[uint64]*UserActivityTask // 任务队列，key为用户ID
	mu        sync.RWMutex                 // 读写锁
	ticker    *time.Ticker                 // 定时器
	closed    bool                         // 是否已关闭
	closeCh   chan struct{}                // 关闭信号
	requestCh chan uint64                  // 请求缓冲通道
	wg        sync.WaitGroup               // 等待组
}

var (
	manager *UserActivityManager
	once    sync.Once
)

// GetUserActivityManager 获取用户活跃时间管理器单例
func GetUserActivityManager() *UserActivityManager {
	once.Do(func() {
		manager = &UserActivityManager{
			tasks:     make(map[uint64]*UserActivityTask),
			ticker:    time.NewTicker(5 * time.Second), // 每5秒检查一次
			closeCh:   make(chan struct{}),
			requestCh: make(chan uint64, 1000), // 1000缓冲区
		}
		manager.start()
	})
	return manager
}

// UpdateUserActivity 更新用户活跃时间
func (m *UserActivityManager) UpdateUserActivity(userID uint64) {
	// 检查是否已关闭
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	// 非阻塞发送到channel
	select {
	case m.requestCh <- userID:
		// 成功发送到channel
	default:
		// channel满了，直接丢弃（可以根据需要调整策略）
		// 在高并发场景下，偶尔丢失一次活跃时间更新是可以接受的
	}
}

// handleUserActivityRequest 处理用户活跃时间更新请求
func (m *UserActivityManager) handleUserActivityRequest(userID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	now := time.Now()
	// 检查是否已存在该用户的任务
	if task, exists := m.tasks[userID]; exists {
		// 更新最后活跃时间为当前时间
		task.LastActiveTime = now
	} else {
		// 创建新任务
		m.tasks[userID] = &UserActivityTask{
			UserID:         userID,
			LastActiveTime: now,
			CreatedAt:      now,
		}
	}
}

// start 启动管理器
func (m *UserActivityManager) start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.ticker.C:
				m.processExpiredTasks()
			case userID := <-m.requestCh:
				m.handleUserActivityRequest(userID)
			case <-m.closeCh:
				return
			}
		}
	}()
}

// processExpiredTasks 处理过期任务
func (m *UserActivityManager) processExpiredTasks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expiredTasks []*UserActivityTask

	for userID, task := range m.tasks {
		// 检查刷入条件
		// 1. 活跃时间距今超过15秒
		// 2. 任务创建时间距今超过45秒（强制刷入）
		if now.Sub(task.LastActiveTime) > 15*time.Second || now.Sub(task.CreatedAt) > 45*time.Second {
			expiredTasks = append(expiredTasks, task)
			delete(m.tasks, userID)
		}
	}

	// 批量刷入数据库
	if len(expiredTasks) > 0 {
		m.flushTasks(expiredTasks)
	}
}

// flushTasks 批量刷入任务到数据库
func (m *UserActivityManager) flushTasks(tasks []*UserActivityTask) {
	for _, task := range tasks {
		// 异步刷入，避免阻塞主流程
		userStatistics.UpdateUserActivity(task.UserID, task.LastActiveTime)
	}
}

// Close 关闭管理器，刷入所有剩余任务
func (m *UserActivityManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}
	m.closed = true

	// 停止定时器和关闭信号
	m.ticker.Stop()
	close(m.closeCh)
	m.wg.Wait()

	// 处理剩余请求并刷入所有任务
	close(m.requestCh)
	for userID := range m.requestCh {
		now := time.Now()
		// 检查是否已存在该用户的任务
		if task, exists := m.tasks[userID]; exists {
			// 更新最后活跃时间为当前时间
			task.LastActiveTime = now
		} else {
			// 创建新任务
			m.tasks[userID] = &UserActivityTask{
				UserID:         userID,
				LastActiveTime: now,
				CreatedAt:      now,
			}
		}
	}

	// 同步刷入所有剩余任务
	for _, task := range m.tasks {
		userStatistics.UpdateUserActivity(task.UserID, task.LastActiveTime)
	}
	m.tasks = nil
}

// UpdateUserActivity 全局函数，更新用户活跃时间
func UpdateUserActivity(userID uint64) {
	if userID == 0 {
		return
	}
	GetUserActivityManager().UpdateUserActivity(userID)
}

// CloseUpdateUserLastActiveTime 关闭用户活跃时间更新服务
func CloseUpdateUserLastActiveTime() {
	if manager != nil {
		manager.Close()
	}
}
