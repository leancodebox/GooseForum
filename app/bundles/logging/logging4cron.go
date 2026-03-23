package logging

import (
	"fmt"
	"log/slog"
)

type CronLogging struct {
}

func (itself CronLogging) Printf(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...))
}

func (itself CronLogging) Info(msg string, keysAndValues ...any) {
	slog.Info(msg, keysAndValues...)
}

// Error logs an error condition.
func (itself CronLogging) Error(err error, msg string, keysAndValues ...any) {
	slog.Error("error", "msg", msg, "err", err, keysAndValues)
}
