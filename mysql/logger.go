package mysql

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var (
	Logger *logrus.Logger
)

// 初始化日志打印函数
func SetLogger(logger *logrus.Logger) {
	if logger != nil {
		Logger = logger
	}
}

func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{})
	Logger.SetOutput(os.Stdout)
}

type logger struct{}

func (logger) Print(v ...interface{}) {
	// 日志部分
	Logger.Debug(v...)
	if len(v) < 5 {
		// 此时，无法进行慢查询的检测
		return
	}
	// 慢sql检测
	AnalyzerCallbackV2(v[3].(string), v[4].([]interface{}), v[2].(time.Duration))
}
