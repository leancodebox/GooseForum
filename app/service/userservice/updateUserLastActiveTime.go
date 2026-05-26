package userservice

import (
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

const (
	userActivityTickInterval = 5 * time.Second
	userActivityIdleFlush    = 15 * time.Second
	userActivityMaxDelay     = 45 * time.Second
	userActivityQueueSize    = 1024
)

// UserActivityTask tracks a pending last-active-time write.
type UserActivityTask struct {
	UserID         uint64
	LastActiveTime time.Time
	CreatedAt      time.Time
}

type userActivityRequest struct {
	UserID     uint64
	ActiveTime time.Time
}

// UserActivityManager batches user activity writes.
type UserActivityManager struct {
	mu        sync.RWMutex
	ticker    *time.Ticker
	closed    bool
	closeCh   chan struct{}
	requestCh chan userActivityRequest
	flushFn   func([]*UserActivityTask)
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
			ticker:    time.NewTicker(userActivityTickInterval),
			closeCh:   make(chan struct{}),
			requestCh: make(chan userActivityRequest, userActivityQueueSize),
			flushFn:   flushUserActivityTasks,
		}
		manager.start()
		closer.Register(CloseUpdateUserLastActiveTime)
	})
	return manager
}

// UpdateUserActivity queues a non-blocking user activity update.
func (m *UserActivityManager) UpdateUserActivity(userID uint64) {
	m.UpdateUserActivityAt(userID, time.Now())
}

func (m *UserActivityManager) UpdateUserActivityAt(userID uint64, activeTime time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return
	}

	select {
	case m.requestCh <- userActivityRequest{UserID: userID, ActiveTime: activeTime}:
	default:
	}
}

// handleUserActivityRequest merges activity updates by user ID.
func handleUserActivityRequest(tasks map[uint64]*UserActivityTask, req userActivityRequest) {
	now := time.Now()
	if req.ActiveTime.IsZero() {
		req.ActiveTime = now
	}
	if task, exists := tasks[req.UserID]; exists {
		if req.ActiveTime.After(task.LastActiveTime) {
			task.LastActiveTime = req.ActiveTime
		}
	} else {
		tasks[req.UserID] = &UserActivityTask{
			UserID:         req.UserID,
			LastActiveTime: req.ActiveTime,
			CreatedAt:      now,
		}
	}
}

// start launches the activity queue worker.
func (m *UserActivityManager) start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		tasks := make(map[uint64]*UserActivityTask)
		for {
			select {
			case <-m.ticker.C:
				tasks = m.processExpiredTasks(tasks)
			case req := <-m.requestCh:
				handleUserActivityRequest(tasks, req)
			case <-m.closeCh:
				m.drain(tasks)
				m.flushTasks(mapValues(tasks))
				return
			}
		}
	}()
}

// processExpiredTasks flushes tasks old enough to write.
func (m *UserActivityManager) processExpiredTasks(tasks map[uint64]*UserActivityTask) map[uint64]*UserActivityTask {
	now := time.Now()
	var expiredTasks []*UserActivityTask
	activeTasks := make(map[uint64]*UserActivityTask, len(tasks))

	for userID, task := range tasks {
		if now.Sub(task.LastActiveTime) > userActivityIdleFlush || now.Sub(task.CreatedAt) > userActivityMaxDelay {
			expiredTasks = append(expiredTasks, task)
			continue
		}
		activeTasks[userID] = task
	}

	if len(expiredTasks) > 0 {
		m.flushTasks(expiredTasks)
	}
	return activeTasks
}

func (m *UserActivityManager) drain(tasks map[uint64]*UserActivityTask) {
	for {
		select {
		case req := <-m.requestCh:
			handleUserActivityRequest(tasks, req)
		default:
			return
		}
	}
}

// flushTasks writes activity tasks to storage.
func (m *UserActivityManager) flushTasks(tasks []*UserActivityTask) {
	if m.flushFn == nil {
		return
	}
	m.flushFn(tasks)
}

func flushUserActivityTasks(tasks []*UserActivityTask) {
	for _, task := range tasks {
		userStatistics.UpdateUserActivity(task.UserID, task.LastActiveTime)
	}
}

func mapValues(tasks map[uint64]*UserActivityTask) []*UserActivityTask {
	values := make([]*UserActivityTask, 0, len(tasks))
	for _, task := range tasks {
		values = append(values, task)
	}
	return values
}

// Close stops the manager and flushes remaining activity updates.
func (m *UserActivityManager) Close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true

	m.ticker.Stop()
	close(m.closeCh)
	m.mu.Unlock()

	m.wg.Wait()
}

// UpdateUserActivity queues a global user activity update.
func UpdateUserActivity(userID uint64) {
	if userID == 0 {
		return
	}
	GetUserActivityManager().UpdateUserActivity(userID)
}

func UpdateUserActivityAt(userID uint64, activeTime time.Time) {
	if userID == 0 {
		return
	}
	GetUserActivityManager().UpdateUserActivityAt(userID, activeTime)
}

// CloseUpdateUserLastActiveTime stops the global activity manager.
func CloseUpdateUserLastActiveTime() error {
	if manager != nil {
		manager.Close()
	}
	return nil
}
