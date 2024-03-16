package fluentbit

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

type FluentBitLogFormat struct{}

func (f *FluentBitLogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	time := fmt.Sprintf("[%s]", entry.Time.Format("2006/01/02 15:04:05"))
	b.WriteString(time)

	level := fmt.Sprintf(" [%5s] ", entry.Level.String())
	b.WriteString(level)

	if i, ok := entry.Data["interface"]; ok {
		b.WriteString(fmt.Sprintf("[%s] ", i))
	}

	if entry.Message != "" {
		b.WriteString(entry.Message)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func GetLogger(name string, level logrus.Level) *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(new(FluentBitLogFormat))
	return logger.WithFields(logrus.Fields{"interface": name})
}
