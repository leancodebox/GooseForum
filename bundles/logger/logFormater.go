package logger

import (
	"bytes"
	"fmt"
	"github.com/leancodebox/goose/luckrand"
	"github.com/sirupsen/logrus"
)

// TextFormatter formats logs into text
type TextFormatter struct {
}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	trace := luckrand.MyTrace()
	b.WriteString(fmt.Sprintf("%v %v", trace.GetNextTrace(), entry.Message))
	b.WriteByte('\n')
	return b.Bytes(), nil
}
