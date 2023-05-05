package logging

import (
	"bytes"
	"fmt"
	"github.com/leancodebox/goose/fileopt"
	"github.com/leancodebox/goose/luckrand"
	"github.com/leancodebox/goose/preferences"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	LogTypeStdout = "stdout"
	LogTypeFile   = "file"
)

var log = logrus.StandardLogger()

// 防止外部调用，本包的目的是为了做一个隔离。目的是为了以后logger替换无感知。不可以直接暴露std。否则后续替换将无法进行
func std() *logrus.Logger {
	return log
}

func Info(args ...any) {
	std().Info(args...)
}

func Printf(format string, args ...interface{}) {
	std().Printf(format, args...)
}

func Println(args ...interface{}) {
	std().Println(args...)
}

func Error(args ...interface{}) {
	std().Error(args...)
}

func ErrIf(err error) bool {
	if err != nil {
		std().Error(err)
		return true
	}
	return false
}

var (
	logType = preferences.Get("log.type")
	logPath = preferences.Get("log.path", "./storage/log/app.log")
	debug   = preferences.GetBool("app.debug", true)
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&LogFormatter{})

	log.Out = os.Stdout
	if debug {
		log.Level = logrus.TraceLevel
	}

	switch logType {
	default:
		log.Info("Unknown Log Output Type")
	case LogTypeStdout:
	case LogTypeFile:
		// You could set this to any `io.Writer` such as a file
		if err := fileopt.FilePutContents(logPath, []byte(""), true); err != nil {
			log.Info(err)
			return
		}

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Info("Failed to log to file, using default stderr")
			return
		}
		log.Out = file
	}
}

type LogFormatter struct{}

func (m *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	var msg string
	trace := luckrand.MyTrace()
	if entry.HasCaller() {
		fName := filepath.Base(entry.Caller.File)
		msg = fmt.Sprintf("[%-7s] %v [%s] [%s:%d %s] %s\n",
			timestamp, entry.Level.String(), trace.GetNextTrace(), fName, entry.Caller.Line, entry.Caller.Function, entry.Message)
	} else {
		msg = fmt.Sprintf("[%v] [%s] [%s] %s\n", trace.GetNextTrace(), timestamp, entry.Level, entry.Message)
	}

	b.WriteString(msg)
	return b.Bytes(), nil
}
