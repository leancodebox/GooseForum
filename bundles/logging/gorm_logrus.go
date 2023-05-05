package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
	gLogger "gorm.io/gorm/logger"
)

// GormLogger 操作对象，实现 gLogger.Interface
type GormLogger struct {
	LogrusLogger  *logrus.Logger
	SlowThreshold time.Duration
}

// NewGormLogger 外部调用。实例化一个 GormLogger 对象，示例：
//
//	DB, err := gorm.Open(dbConfig, &gorm.Config{
//		Logger: logger.NewGormLogger(),
//	})
func NewGormLogger() GormLogger {
	return GormLogger{
		LogrusLogger:  log,                    // 使用全局的 logger.Logger 对象
		SlowThreshold: 200 * time.Millisecond, // 慢查询阈值，单位为千分之一秒
	}
}

// LogMode 实现 gLogger.Interface 的 LogMode 方法
func (l GormLogger) LogMode(level gLogger.LogLevel) gLogger.Interface {
	return GormLogger{
		LogrusLogger:  l.LogrusLogger,
		SlowThreshold: l.SlowThreshold,
	}
}

// Info 实现 gLogger.Interface 的 Info 方法
func (l GormLogger) Info(ctx context.Context, str string, args ...any) {
	l.logger().Infof(str, args...)
}

// Warn 实现 gLogger.Interface 的 Warn 方法
func (l GormLogger) Warn(ctx context.Context, str string, args ...any) {
	l.logger().Warnf(str, args...)
}

// Error 实现 gLogger.Interface 的 Error 方法
func (l GormLogger) Error(ctx context.Context, str string, args ...any) {
	l.logger().Errorf(str, args...)
}

// Trace 实现 gLogger.Interface 的 Trace 方法
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 获取运行时间
	elapsed := time.Since(begin)
	// 获取 SQL 请求和返回条数
	sql, rows := fc()

	// 通用字段
	logFields := logrus.Fields{
		"sql":  sql,
		"time": fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
		"rows": rows,
	}

	// Gorm 错误
	if err != nil {
		// 记录未找到的错误使用 warning 等级
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.logger().Warn("Database ErrRecordNotFound", logFields)
		} else {
			// 其他错误使用 error 等级
			fields := logFields
			fields["err"] = err
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
func (l GormLogger) logger() *logrus.Logger {

	return l.LogrusLogger
}
