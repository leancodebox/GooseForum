// Package closer stores process-wide shutdown callbacks.
package closer

import (
	"fmt"
	"log/slog"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Priority int

const (
	PriorityProducer Priority = 100
	PriorityFlush    Priority = 200
	PriorityCache    Priority = 300
	PriorityDefault  Priority = 500
	PriorityDatabase Priority = 900
	PriorityLogger   Priority = 1000
)

var (
	mu           sync.Mutex
	entries      []closerEntry
	closeTimeout = 10 * time.Second
	nextSeq      uint64
)

type closerEntry struct {
	f        func() error
	caller   string
	priority Priority
	seq      uint64
}

// Register adds f to the process shutdown callback list.
func Register(f func() error) {
	register(1, PriorityDefault, f)
}

// RegisterPriority adds f to the shutdown callback list with an explicit close phase.
func RegisterPriority(priority Priority, f func() error) {
	register(1, priority, f)
}

// Bind registers the Close method of c as a shutdown callback.
func Bind(c interface{ Close() error }) {
	register(1, PriorityDefault, c.Close)
}

// BindPriority registers the Close method of c with an explicit close phase.
func BindPriority(priority Priority, c interface{ Close() error }) {
	register(1, priority, c.Close)
}

func register(skip int, priority Priority, f func() error) {
	mu.Lock()
	defer mu.Unlock()

	_, file, line, ok := runtime.Caller(skip + 1)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	entries = append(entries, closerEntry{
		f:        f,
		caller:   caller,
		priority: priority,
		seq:      nextSeq,
	})
	nextSeq++
	slog.Info("closer: registered resource",
		"caller", caller,
		"priority", priority,
		"total", len(entries),
	)
}

// CloseAll runs all registered shutdown callbacks in reverse registration order.
func CloseAll() {
	mu.Lock()
	items := append([]closerEntry(nil), entries...)
	entries = nil
	mu.Unlock()

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].priority != items[j].priority {
			return items[i].priority < items[j].priority
		}
		return items[i].seq > items[j].seq
	})

	slog.Info("closer: starting to close all registered resources", "count", len(items))

	for i := range items {
		entry := items[i]
		slog.Info("closer: closing resource",
			"index", i,
			"priority", entry.priority,
			"registered_at", entry.caller,
		)
		if err := closeWithTimeout(entry); err != nil {
			slog.Error("closer: failed to close resource",
				"index", i,
				"priority", entry.priority,
				"error", err,
				"registered_at", entry.caller,
			)
		}
	}

	slog.Info("closer: all resources closed")
}

func closeWithTimeout(entry closerEntry) error {
	done := make(chan error, 1)
	go func() {
		done <- entry.f()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(closeTimeout):
		return fmt.Errorf("close timed out after %s", closeTimeout)
	}
}
