package logging

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"gorm.io/gorm/utils"
	"log/slog"
	"time"

	gormLogger "gorm.io/gorm/logger"
)

// GormLogger 操作对象，实现 gormLogger.Interface
type GormLogger struct {
	Config        gormLogger.Config
	SlowThreshold time.Duration
}

// NewGormLogger 外部调用。实例化一个 GormLogger 对象，示例：
//
//	DB, err := gorm.Open(dbConfig, &gorm.Config{
//		Logger: logger.NewGormLogger(),
//	})
func NewGormLogger() *GormLogger {
	logLevel := gormLogger.Warn
	if debug {
		logLevel = gormLogger.Info
	}
	return &GormLogger{
		Config: gormLogger.Config{
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
		},
		SlowThreshold: 500 * time.Millisecond, // 慢查询阈值，单位为千分之一秒
	}
}

// LogMode 实现 gormLogger.Interface 的 LogMode 方法
func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *l
	newlogger.Config.LogLevel = level
	return &newlogger
}

// Info 实现 gormLogger.Interface 的 Info 方法
func (l *GormLogger) Info(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Info {
		slog.Info(fmt.Sprintf(format, args...))
	}
}

// Warn 实现 gormLogger.Interface 的 Warn 方法
func (l *GormLogger) Warn(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Warn {
		slog.Warn(fmt.Sprintf(format, args...))
	}
}

// Error 实现 gormLogger.Interface 的 Error 方法
func (l *GormLogger) Error(ctx context.Context, format string, args ...any) {
	if l.Config.LogLevel >= gormLogger.Error {
		slog.Error(fmt.Sprintf(format, args...))
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Config.LogLevel <= gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Config.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError):
		sql, rows := fc()
		row := cast.ToString(rows)
		if rows == -1 {
			row = "-"
		}
		slog.Warn("gormLogger",
			"fileWithLineNum", utils.FileWithLineNum(),
			"err", err,
			"elapsed", float64(elapsed.Nanoseconds())/1e6,
			"row", row,
			"sql", sql,
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.Config.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		row := cast.ToString(rows)
		if rows == -1 {
			row = "-"
		}
		slog.Warn("gormLogger",
			"fileWithLineNum", utils.FileWithLineNum(),
			"slowLog", slowLog,
			"elapsed", float64(elapsed.Nanoseconds())/1e6,
			"row", row,
			"sql", sql,
		)
	case l.Config.LogLevel == gormLogger.Info:
		sql, rows := fc()
		row := cast.ToString(rows)
		if rows == -1 {
			row = "-"
		}
		slog.Warn("gormLogger",
			"fileWithLineNum", utils.FileWithLineNum(),
			"elapsed", float64(elapsed.Nanoseconds())/1e6,
			"row", row,
			"sql", sql,
		)
	case debug:
		sql, rows := fc()
		slog.Debug("gormLogger",
			"sql", sql,
			"time", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
			"rows", rows,
		)
	}
}
