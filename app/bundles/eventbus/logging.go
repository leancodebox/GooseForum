package eventbus

import (
	"log/slog"
	"maps"

	"github.com/ThreeDotsLabs/watermill"
)

type dynamicSlogLogger struct {
	fields watermill.LogFields
}

func (l *dynamicSlogLogger) Error(msg string, err error, fields watermill.LogFields) {
	slog.Error(msg, append(l.prepareArgs(fields), "err", err)...)
}

func (l *dynamicSlogLogger) Info(msg string, fields watermill.LogFields) {
	slog.Info(msg, l.prepareArgs(fields)...)
}

func (l *dynamicSlogLogger) Debug(msg string, fields watermill.LogFields) {
	slog.Debug(msg, l.prepareArgs(fields)...)
}

func (l *dynamicSlogLogger) Trace(msg string, fields watermill.LogFields) {
	slog.Debug(msg, append(l.prepareArgs(fields), "trace", true)...)
}

func (l *dynamicSlogLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newFields := make(watermill.LogFields, len(l.fields)+len(fields))
	maps.Copy(newFields, l.fields)
	maps.Copy(newFields, fields)
	return &dynamicSlogLogger{fields: newFields}
}

func (l *dynamicSlogLogger) prepareArgs(fields watermill.LogFields) []any {
	args := make([]any, 0, (len(l.fields)+len(fields))*2)
	for k, v := range l.fields {
		args = append(args, k, v)
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}
