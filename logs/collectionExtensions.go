package logs

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

func SetLogsModule(level logrus.Level) {
	fmt.Println("enable loggers logging module")
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(level)
}
