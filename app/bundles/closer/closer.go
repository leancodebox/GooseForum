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

// Register 注册一个关闭函数
func Register(f func() error) {
	register(1, f)
}

// Bind 绑定一个实现了 Close() error 接口的对象
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

// CloseAll 执行所有已注册的关闭逻辑
// 通常在 main 函数退出前或捕获到系统信号时调用
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()

	slog.Info("closer: starting to close all registered resources", "count", len(entries))

	// 倒序执行（先初始化的后关闭，符合依赖关系）
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

	// 执行完后清空，防止重复执行
	entries = nil
	slog.Info("closer: all resources closed")
}
