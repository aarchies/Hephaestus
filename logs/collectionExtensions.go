package logs

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// AddLogsModule 添加日志模块
func AddLogsModule(level logrus.Level) {
	fmt.Println("enable logrus logging module")
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(level)
}
