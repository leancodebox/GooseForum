// Package closer stores process-wide shutdown callbacks.
package closer

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
)

var (
	mu      sync.Mutex
	entries []closerEntry
)

type closerEntry struct {
	f      func() error
	caller string
}

// Register adds f to the process shutdown callback list.
func Register(f func() error) {
	register(1, f)
}

// Bind registers the Close method of c as a shutdown callback.
func Bind(c interface{ Close() error }) {
	register(1, c.Close)
}

func register(skip int, f func() error) {
	mu.Lock()
	defer mu.Unlock()

	_, file, line, ok := runtime.Caller(skip + 1)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	entries = append(entries, closerEntry{
		f:      f,
		caller: caller,
	})
	slog.Info("closer: registered resource",
		"caller", caller,
		"total", len(entries),
	)
}

// CloseAll runs all registered shutdown callbacks in reverse registration order.
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()

	slog.Info("closer: starting to close all registered resources", "count", len(entries))

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		slog.Info("closer: closing resource",
			"index", i,
			"registered_at", entry.caller,
		)
		if err := entry.f(); err != nil {
			slog.Error("closer: failed to close resource",
				"index", i,
				"error", err,
				"registered_at", entry.caller,
			)
		}
	}

	entries = nil
	slog.Info("closer: all resources closed")
}
