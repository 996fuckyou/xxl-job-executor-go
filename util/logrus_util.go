package util

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type LogFormatter struct {

}

func (log *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var data *bytes.Buffer
	var newLog string
	if entry.Buffer != nil {
		data = entry.Buffer
	} else {
		data = &bytes.Buffer{}
	}
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	if entry.HasCaller() {
		fileName := filepath.Base(entry.Caller.File)
		newLog = fmt.Sprintf("[%s] [%s] [%s:%d] %s\n", timestamp, entry.Level, fileName, entry.Caller.Line, entry.Message)
	} else {
		newLog = fmt.Sprintf("[%s] [%s] %s\n", timestamp, entry.Level, entry.Message)
	}
	data.WriteString(newLog)
	return data.Bytes(), nil
}
