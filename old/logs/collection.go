package logs

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func LogMode(mode string, isfile bool) {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.FullTimestamp = true

	logrus.SetFormatter(customFormatter) // 设置日志输出格式，默认值为 logrus.TextFormatter
	logrus.SetReportCaller(false)        // 设置日志是否记录被调用的位置，默认值为 false
	logrus.SetLevel(logrus.DebugLevel)   // 设置日志级别，默认值为 logrus.InfoLevel

	switch mode {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	if isfile {
		logFile, _ := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		mw := io.MultiWriter(os.Stdout, logFile)
		logrus.SetOutput(mw)
	}
}
