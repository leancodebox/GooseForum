package userservice

import (
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

// UserActivityTask tracks a pending last-active-time write.
type UserActivityTask struct {
	UserID         uint64
	LastActiveTime time.Time
	CreatedAt      time.Time
}

// UserActivityManager batches user activity writes.
type UserActivityManager struct {
	tasks     map[uint64]*UserActivityTask
	mu        sync.RWMutex
	ticker    *time.Ticker
	closed    bool
	closeCh   chan struct{}
	requestCh chan uint64
	wg        sync.WaitGroup
}

var (
	manager *UserActivityManager
	once    sync.Once
)

// GetUserActivityManager returns the singleton activity manager.
func GetUserActivityManager() *UserActivityManager {
	once.Do(func() {
		manager = &UserActivityManager{
			tasks:     make(map[uint64]*UserActivityTask),
			ticker:    time.NewTicker(5 * time.Second),
			closeCh:   make(chan struct{}),
			requestCh: make(chan uint64, 1024),
		}
		manager.start()
		closer.Register(CloseUpdateUserLastActiveTime)
	})
	return manager
}

// UpdateUserActivity queues a non-blocking user activity update.
func (m *UserActivityManager) UpdateUserActivity(userID uint64) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	select {
	case m.requestCh <- userID:
	default:
	}
}

// handleUserActivityRequest merges activity updates by user ID.
func (m *UserActivityManager) handleUserActivityRequest(userID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	now := time.Now()
	if task, exists := m.tasks[userID]; exists {
		task.LastActiveTime = now
	} else {
		m.tasks[userID] = &UserActivityTask{
			UserID:         userID,
			LastActiveTime: now,
			CreatedAt:      now,
		}
	}
}

// start launches the activity queue worker.
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

// processExpiredTasks flushes tasks old enough to write.
func (m *UserActivityManager) processExpiredTasks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expiredTasks []*UserActivityTask

	for userID, task := range m.tasks {
		if now.Sub(task.LastActiveTime) > 15*time.Second || now.Sub(task.CreatedAt) > 45*time.Second {
			expiredTasks = append(expiredTasks, task)
			delete(m.tasks, userID)
		}
	}

	if len(expiredTasks) > 0 {
		m.flushTasks(expiredTasks)
	}
}

// flushTasks writes activity tasks to storage.
func (m *UserActivityManager) flushTasks(tasks []*UserActivityTask) {
	for _, task := range tasks {
		userStatistics.UpdateUserActivity(task.UserID, task.LastActiveTime)
	}
}

// Close stops the manager and flushes remaining activity updates.
func (m *UserActivityManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}
	m.closed = true

	m.ticker.Stop()
	close(m.closeCh)
	m.wg.Wait()

	close(m.requestCh)
	for userID := range m.requestCh {
		now := time.Now()
		if task, exists := m.tasks[userID]; exists {
			task.LastActiveTime = now
		} else {
			m.tasks[userID] = &UserActivityTask{
				UserID:         userID,
				LastActiveTime: now,
				CreatedAt:      now,
			}
		}
	}

	for _, task := range m.tasks {
		userStatistics.UpdateUserActivity(task.UserID, task.LastActiveTime)
	}
	m.tasks = nil
}

// UpdateUserActivity queues a global user activity update.
func UpdateUserActivity(userID uint64) {
	if userID == 0 {
		return
	}
	GetUserActivityManager().UpdateUserActivity(userID)
}

// CloseUpdateUserLastActiveTime stops the global activity manager.
func CloseUpdateUserLastActiveTime() error {
	if manager != nil {
		manager.Close()
	}
	return nil
}
