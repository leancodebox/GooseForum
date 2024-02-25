package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// GormLogger 操作对象，实现 gormLogger.Interface
type GormLogger struct {
	Config        GormLoggerConfig
	SlowThreshold time.Duration
}

type GormLoggerConfig struct {
	LogLevel gormLogger.LogLevel
}

// NewGormLogger 外部调用。实例化一个 GormLogger 对象，示例：
//
//	DB, err := gorm.Open(dbConfig, &gorm.Config{
//		Logger: logger.NewGormLogger(),
//	})
func NewGormLogger() GormLogger {
	logLevel := gormLogger.Warn
	if debug {
		logLevel = gormLogger.Info
	}
	return GormLogger{
		Config: GormLoggerConfig{
			LogLevel: logLevel,
		},
		SlowThreshold: 500 * time.Millisecond, // 慢查询阈值，单位为千分之一秒
	}
}

// LogMode 实现 gormLogger.Interface 的 LogMode 方法
func (l GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := l
	newlogger.Config.LogLevel = level
	return &newlogger
}

// Info 实现 gormLogger.Interface 的 Info 方法
func (l GormLogger) Info(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Info {
		slog.Info(fmt.Sprintf(format, args...))
	}
}

// Warn 实现 gormLogger.Interface 的 Warn 方法
func (l GormLogger) Warn(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Warn {
		slog.Warn(fmt.Sprintf(format, args...))
	}
}

// Error 实现 gormLogger.Interface 的 Error 方法
func (l GormLogger) Error(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Error {
		slog.Error(fmt.Sprintf(format, args...))
	}
}

// Trace 实现 gormLogger.Interface 的 Trace 方法
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Config.LogLevel <= gormLogger.Silent {
		if debug {
			elapsed := time.Since(begin)
			sql, rows := fc()
			slog.Debug("Database Query", slog.Group(
				"gorm",
				"sql", sql,
				"time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
				"rows", rows,
			))
		}
		return
	}
	// 获取运行时间
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound)):
		slog.Warn("Database ErrRecordNotFound", slog.Group(
			"gorm",
			"sql", sql,
			"time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
			"rows", rows,
		))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		slog.Warn("Database Slow Log", slog.Group(
			"gorm",
			"sql", sql,
			"time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
		))
	}

}
