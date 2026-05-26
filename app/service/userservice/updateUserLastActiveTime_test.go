package userservice

import (
	"testing"
	"time"
)

func TestProcessExpiredTasksRebuildsActiveMap(t *testing.T) {
	now := time.Now()
	oldTask := &UserActivityTask{UserID: 1, LastActiveTime: now.Add(-20 * time.Second), CreatedAt: now.Add(-20 * time.Second)}
	activeTask := &UserActivityTask{UserID: 2, LastActiveTime: now, CreatedAt: now}
	tasks := map[uint64]*UserActivityTask{
		1: oldTask,
		2: activeTask,
	}

	var flushed []*UserActivityTask
	manager := &UserActivityManager{
		flushFn: func(tasks []*UserActivityTask) {
			flushed = append(flushed, tasks...)
		},
	}

	next := manager.processExpiredTasks(tasks)

	if len(next) != 1 || next[2] != activeTask {
		t.Fatalf("active tasks = %#v, want only user 2", next)
	}
	if len(flushed) != 1 || flushed[0] != oldTask {
		t.Fatalf("flushed tasks = %#v, want old task", flushed)
	}
}

func TestUserActivityManagerFlushesPendingOnClose(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	var flushed []*UserActivityTask
	manager := &UserActivityManager{
		ticker:    ticker,
		closeCh:   make(chan struct{}),
		requestCh: make(chan userActivityRequest, 8),
		flushFn: func(tasks []*UserActivityTask) {
			flushed = append(flushed, tasks...)
		},
	}
	manager.start()

	manager.UpdateUserActivity(7)
	manager.UpdateUserActivity(7)
	manager.UpdateUserActivity(9)
	manager.Close()

	if len(flushed) != 2 {
		t.Fatalf("flushed task count = %d, want 2", len(flushed))
	}
	seen := map[uint64]bool{}
	for _, task := range flushed {
		seen[task.UserID] = true
	}
	if !seen[7] || !seen[9] {
		t.Fatalf("flushed users = %#v, want users 7 and 9", seen)
	}
}

func TestUserActivityManagerUsesEventActiveTime(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	activeTime := time.Now().Add(-time.Minute)
	var flushed []*UserActivityTask
	manager := &UserActivityManager{
		ticker:    ticker,
		closeCh:   make(chan struct{}),
		requestCh: make(chan userActivityRequest, 8),
		flushFn: func(tasks []*UserActivityTask) {
			flushed = append(flushed, tasks...)
		},
	}
	manager.start()

	manager.UpdateUserActivityAt(7, activeTime)
	manager.Close()

	if len(flushed) != 1 {
		t.Fatalf("flushed task count = %d, want 1", len(flushed))
	}
	if !flushed[0].LastActiveTime.Equal(activeTime) {
		t.Fatalf("last active time = %v, want %v", flushed[0].LastActiveTime, activeTime)
	}
}

func TestUserActivityManagerKeepsLatestActiveTime(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	newer := time.Now()
	older := newer.Add(-time.Minute)
	var flushed []*UserActivityTask
	manager := &UserActivityManager{
		ticker:    ticker,
		closeCh:   make(chan struct{}),
		requestCh: make(chan userActivityRequest, 8),
		flushFn: func(tasks []*UserActivityTask) {
			flushed = append(flushed, tasks...)
		},
	}
	manager.start()

	manager.UpdateUserActivityAt(7, newer)
	manager.UpdateUserActivityAt(7, older)
	manager.Close()

	if len(flushed) != 1 {
		t.Fatalf("flushed task count = %d, want 1", len(flushed))
	}
	if !flushed[0].LastActiveTime.Equal(newer) {
		t.Fatalf("last active time = %v, want latest %v", flushed[0].LastActiveTime, newer)
	}
}
