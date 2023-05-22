package logging

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/leancodebox/goose/fileopt"
	"github.com/leancodebox/goose/preferences"
	"github.com/sirupsen/logrus"
)

const (
	LogTypeStdout = "stdout"
	LogTypeFile   = "file"
)

type Entry struct {
	level   logrus.Level
	message string
}

var (
	log        = logrus.StandardLogger()
	logChannel = make(chan *Entry, 1024*1024)
	wg         sync.WaitGroup
)

func std() *logrus.Logger {
	return log
}

func Info(args ...interface{}) {
	sendLog(logrus.InfoLevel, fmt.Sprint(args...))
}

func Printf(format string, args ...interface{}) {
	sendLog(logrus.InfoLevel, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	sendLog(logrus.InfoLevel, fmt.Sprintln(args...))
}

func Error(args ...interface{}) {
	sendLog(logrus.ErrorLevel, fmt.Sprint(args...))
}

func ErrIf(err error) bool {
	if err != nil {
		sendLog(logrus.ErrorLevel, err.Error())
		return true
	}
	return false
}

// Send log entry to the channel
func sendLog(level logrus.Level, msg string) {

	if caller, success := getCaller(3); success {
		msg = caller + ":" + msg
	}

	entry := &Entry{
		level:   level,
		message: msg,
	}
	logChannel <- entry
}

func getCaller(depth int) (string, bool) {
	pc, file, line, ok := runtime.Caller(depth) // 1 表示跳过当前函数的调用帧
	if ok {
		f := runtime.FuncForPC(pc)
		if f != nil {
			funcName := f.Name()
			return fmt.Sprintf("[%s:%s:%d] message", funcName, filepath.Base(file), line), true
		}
	}
	return "", false
}

func processLogEntries() {
	defer wg.Done()
	for entry := range logChannel {
		std().Log(entry.level, entry.message)
	}
}

var (
	logType = preferences.Get("log.type")
	logPath = preferences.Get("log.path", "./storage/logs/thh.log")
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

	wg.Add(1)
	go processLogEntries()
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
	msg = fmt.Sprintf("[%-7s] %v %s\n",
		timestamp, entry.Level.String(), entry.Message)
	b.WriteString(msg)
	return b.Bytes(), nil
}

func Shutdown() {
	close(logChannel)
	wg.Wait()
	fmt.Println("logging 👋")
}
