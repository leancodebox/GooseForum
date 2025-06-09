package logging

import (
	"context"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/asyncwrite"
	"github.com/leancodebox/GooseForum/app/bundles/fileopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	LogTypeStdout = "stdout"
	LogTypeFile   = "file"
)

func ErrIf(err error) bool {
	if err != nil {
		slog.Error(err.Error())
		return true
	}
	return false
}

var (
	debug      = setting.IsDebug()
	logType    = preferences.Get("log.type", LogTypeStdout)
	logPath    = preferences.Get("log.path", "./storage/logs/run.log")
	rolling    = preferences.GetBool("log.rolling", false)
	maxAge     = preferences.GetInt("log.maxAge", 30)
	maxSize    = preferences.GetInt("log.maxSize", 1024)
	maxBackUps = preferences.GetInt("log.maxBackUps", 1024)
)

var aw *asyncwrite.AsyncW
var once = new(sync.Once)

func init() {
	Init()
}

func Init() {
	once.Do(func() {
		var logOut io.Writer
		logOut = os.Stdout
		switch logType {
		default:
			slog.Info("Unknown Log Output Type")
		case LogTypeStdout:
		case LogTypeFile:
			if rolling {
				logOut = getAsyncFileIoRolling()
			} else {
				logOut = getFileIo()
			}
		}
		logLevel := slog.LevelInfo
		if debug {
			logLevel = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewJSONHandler(logOut, &slog.HandlerOptions{
			AddSource:   true,
			ReplaceAttr: replace,
			Level:       logLevel,
		})))
	})
}

func getFileIo() *os.File {
	logOut := os.Stdout
	if err := fileopt.IsExistOrCreate(logPath, ""); err != nil {
		slog.Info("Create log file error", "err", err)

	}
	logOut, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Info("Failed to log to file, using default stderr", "err", err)
	}
	return logOut
}

func getAsyncFileIoRolling() *asyncwrite.AsyncW {
	aw = asyncwrite.AsyncLumberjackBufIo(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackUps, // maxBackUps
		MaxAge:     maxAge,     //days
		LocalTime:  true,
		Compress:   false, // disabled by default
	})
	return aw
}

var rootDir = getRootDir()

func getRootDir() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "???"
	}

	for i := 0; i < 3; i++ {
		file = filepath.Dir(file)
	}
	return strings.ReplaceAll(file, "\\", "/")
}

func replace(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		if source, ok := a.Value.Any().(*slog.Source); ok {
			source.File = strings.TrimPrefix(source.File, rootDir+"/")
			a.Value = slog.StringValue(fmt.Sprintf("%v:%v %v", source.File, source.Line, source.Function))
		}
	case a.Key == slog.TimeKey:
		if item, ok := a.Value.Any().(time.Time); ok {
			a.Value = slog.StringValue(item.Format("2006-01-02 15:04:05.999999999"))
		}
	}
	return a
}

func Shutdown() {
	if aw == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	aw.Stop(ctx)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replace,
	})))
	slog.Info("logging 👋")
	aw = nil
}
