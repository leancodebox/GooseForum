package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger 基于 log/slog 的 GORM 日志记录器
// 相比 GormLogger，提供更好的结构化日志和性能
type GormLogger struct {
	logger        *slog.Logger
	config        gormLogger.Config
	slowThreshold time.Duration
}

// GormLoggerConfig 配置选项
type GormLoggerConfig struct {
	Logger                    *slog.Logger
	LogLevel                  gormLogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	Colorful                  bool
}

// NewGormLogger 创建新的 GormLogger 实例
// 如果 config 为 nil，将使用默认配置
func NewGormLogger(config *GormLoggerConfig) *GormLogger {
	if config == nil {
		config = &GormLoggerConfig{}
	}

	// 设置默认值
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 200 * time.Millisecond
	}
	if config.LogLevel == 0 {
		if debug {
			config.LogLevel = gormLogger.Info
		} else {
			config.LogLevel = gormLogger.Warn
		}
	}

	return &GormLogger{
		logger: config.Logger,
		config: gormLogger.Config{
			LogLevel:                  config.LogLevel,
			IgnoreRecordNotFoundError: true, // config.IgnoreRecordNotFoundError,
			ParameterizedQueries:      config.ParameterizedQueries,
			Colorful:                  config.Colorful,
		},
		slowThreshold: config.SlowThreshold,
	}
}

// NewGormLoggerWithDefault 使用默认配置创建 GormLogger
func NewGormLoggerWithDefault() *GormLogger {
	return NewGormLogger(nil)
}

// LogMode 实现 gormLogger.Interface 的 LogMode 方法
func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.config.LogLevel = level
	return &newLogger
}

// Info 实现 gormLogger.Interface 的 Info 方法
func (l *GormLogger) Info(ctx context.Context, format string, args ...any) {
	if l.config.LogLevel >= gormLogger.Info {
		l.logger.InfoContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Warn 实现 gormLogger.Interface 的 Warn 方法
func (l *GormLogger) Warn(ctx context.Context, format string, args ...any) {
	if l.config.LogLevel >= gormLogger.Warn {
		l.logger.WarnContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Error 实现 gormLogger.Interface 的 Error 方法
func (l *GormLogger) Error(ctx context.Context, format string, args ...any) {
	if l.config.LogLevel >= gormLogger.Error {
		l.logger.ErrorContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Trace 实现 gormLogger.Interface 的 Trace 方法
// 这是核心方法，用于记录 SQL 执行信息
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.config.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建基础日志参数
	baseArgs := []any{
		"source", utils.FileWithLineNum(),
		"elapsed_ms", float64(elapsed.Nanoseconds()) / 1e6,
		"sql", sql,
	}

	// 添加行数信息
	if rows >= 0 {
		baseArgs = append(baseArgs, "rows", rows)
	} else {
		baseArgs = append(baseArgs, "rows", "unknown")
	}

	// 根据不同情况记录日志
	switch {
	case err != nil && l.config.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.config.IgnoreRecordNotFoundError):
		// 错误日志
		errorArgs := append(baseArgs, "error", err)
		l.logger.ErrorContext(ctx, "GORM SQL Error", errorArgs...)

	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.config.LogLevel >= gormLogger.Warn:
		// 慢查询日志
		slowArgs := append(baseArgs,
			"slow_threshold", l.slowThreshold,
			"is_slow", true,
		)
		l.logger.WarnContext(ctx, "GORM Slow SQL", slowArgs...)

	case l.config.LogLevel >= gormLogger.Info:
		// 普通信息日志
		l.logger.InfoContext(ctx, "GORM SQL", baseArgs...)

	default:
		// Debug 模式下的详细日志
		if debug {
			l.logger.DebugContext(ctx, "GORM SQL Debug", baseArgs...)
		}
	}
}

// WithLogger 返回一个使用指定 logger 的新实例
func (l *GormLogger) WithLogger(logger *slog.Logger) *GormLogger {
	newLogger := *l
	newLogger.logger = logger
	return &newLogger
}

// WithSlowThreshold 设置慢查询阈值
func (l *GormLogger) WithSlowThreshold(threshold time.Duration) *GormLogger {
	newLogger := *l
	newLogger.slowThreshold = threshold
	return &newLogger
}

// GetSlowThreshold 获取当前慢查询阈值
func (l *GormLogger) GetSlowThreshold() time.Duration {
	return l.slowThreshold
}

// GetLogLevel 获取当前日志级别
func (l *GormLogger) GetLogLevel() gormLogger.LogLevel {
	return l.config.LogLevel
}
