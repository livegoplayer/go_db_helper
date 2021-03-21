package mysql

import "github.com/sirupsen/logrus"

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
}
