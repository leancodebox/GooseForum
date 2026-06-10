package recovery

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// LogPanic writes a structured panic log with a full stack trace.
func LogPanic(scope string, panicValue any, attrs ...any) {
	fields := []any{
		"scope", scope,
		"panic", fmt.Sprint(panicValue),
		"panic_type", fmt.Sprintf("%T", panicValue),
		"stack", string(debug.Stack()),
	}
	fields = append(fields, attrs...)
	slog.Error("panic recovered", fields...)
}

// Recover logs and swallows a panic inside deferred cleanup.
func Recover(scope string, attrs ...any) {
	if panicValue := recover(); panicValue != nil {
		LogPanic(scope, panicValue, attrs...)
	}
}
