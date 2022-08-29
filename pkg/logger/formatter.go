package logger

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type formatter struct{}

// 日志颜色
var colors = map[logrus.Level]func(a ...interface{}) string{
	logrus.WarnLevel:  color.New(color.FgYellow).Add(color.Bold).SprintFunc(),
	logrus.PanicLevel: color.New(color.BgHiRed).Add(color.Bold).SprintFunc(),
	logrus.FatalLevel: color.New(color.BgRed).Add(color.Bold).SprintFunc(),
	logrus.ErrorLevel: color.New(color.FgRed).Add(color.Bold).SprintFunc(),
	logrus.InfoLevel:  color.New(color.FgCyan).Add(color.Bold).SprintFunc(),
	logrus.DebugLevel: color.New(color.FgWhite).Add(color.Bold).SprintFunc(),
}

// 不同级别前缀与时间的间隔，保持宽度一致
var spaces = map[logrus.Level]int{
	logrus.WarnLevel:  1,
	logrus.PanicLevel: 3,
	logrus.ErrorLevel: 3,
	logrus.FatalLevel: 3,
	logrus.InfoLevel:  4,
	logrus.DebugLevel: 3,
}

func GetLevelText(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel:
		return "Panic"
	case logrus.FatalLevel:
		return "Fatal"
	case logrus.ErrorLevel:
		return "Error"
	case logrus.WarnLevel:
		return "Warning"
	case logrus.InfoLevel:
		return "Info"
	case logrus.DebugLevel:
		return "Debug"
	default:
		return "Unknown"
	}
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	c := color.New()
	c.EnableColor()

	s := c.Sprintf(
		"%s%s %s %s\n",
		colors[entry.Level]("["+GetLevelText(entry.Level)+"]"),
		strings.Repeat(" ", spaces[entry.Level]),
		time.Now().Format("2006-01-02 15:04:05"),
		entry.Message,
	)

	return []byte(s), nil
}
