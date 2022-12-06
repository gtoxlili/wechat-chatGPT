package util

import (
	r "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	"runtime"
)

type LogFormatter struct {
	ChildFormatter r.Formatter
}

func DefaultLogFormatter() *LogFormatter {
	return &LogFormatter{
		ChildFormatter: r.Formatter{
			ChildFormatter: &logrus.TextFormatter{
				DisableColors:   runtime.GOOS == "windows",
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			},
			Line:    true,
			Package: true,
			File:    true,
		},
	}
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level < logrus.InfoLevel {
		return f.ChildFormatter.Format(entry)
	} else {
		return f.ChildFormatter.ChildFormatter.Format(entry)
	}
}
