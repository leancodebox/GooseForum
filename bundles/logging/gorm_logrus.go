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
	SlogLogger    *slog.Logger
	SlowThreshold time.Duration
}

// NewGormLogger 外部调用。实例化一个 GormLogger 对象，示例：
//
//	DB, err := gorm.Open(dbConfig, &gorm.Config{
//		Logger: logger.NewGormLogger(),
//	})
func NewGormLogger() GormLogger {
	return GormLogger{
		SlogLogger:    log,                    // 使用全局的 logger.Logger 对象
		SlowThreshold: 200 * time.Millisecond, // 慢查询阈值，单位为千分之一秒
	}
}

// LogMode 实现 gormLogger.Interface 的 LogMode 方法
func (l GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return GormLogger{
		SlogLogger:    l.SlogLogger,
		SlowThreshold: l.SlowThreshold,
	}
}

// Info 实现 gormLogger.Interface 的 Info 方法
func (l GormLogger) Info(ctx context.Context, format string, args ...any) {
	l.logger().Info(fmt.Sprintf(format, args...))
}

// Warn 实现 gormLogger.Interface 的 Warn 方法
func (l GormLogger) Warn(ctx context.Context, format string, args ...any) {
	l.logger().Warn(fmt.Sprintf(format, args...))
}

// Error 实现 gormLogger.Interface 的 Error 方法
func (l GormLogger) Error(ctx context.Context, format string, args ...any) {
	l.logger().Error(fmt.Sprintf(format, args...))
}

// Trace 实现 gormLogger.Interface 的 Trace 方法
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 获取运行时间
	elapsed := time.Since(begin)
	// 获取 SQL 请求和返回条数
	sql, rows := fc()

	// 通用字段
	logFields := slog.Group(
		"gorm",
		"sql", sql,
		"time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
		"rows", rows,
	)

	// Gorm 错误
	if err != nil {
		// 记录未找到的错误使用 warning 等级
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.logger().Warn("Database ErrRecordNotFound", logFields)
		} else {
			// 其他错误使用 error 等级
			fields := logFields
			l.logger().Error("Database Error", fields)
		}
	}

	// 慢查询日志
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.logger().Warn("Database Slow Log", logFields)
	}

	// 记录所有 SQL 请求
	l.logger().Debug("Database Query", logFields)
}

// logger 内用的辅助方法，
func (l GormLogger) logger() *slog.Logger {
	return l.SlogLogger
}
