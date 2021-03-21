package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
)

type mysqlConfig struct {
	MaxIdleConns int
	MaxOpenConns int
	Config       mysql.Config
	IsWrite      bool
}

var MysqlConnects []*mysqlConfig

var Model *gorm.DB

var models map[string]map[string]*gorm.DB

const (
	// 数据库的读写性质
	ReadOnly = "readonly"
	Write    = "write"
	Proxy    = "proxy"

	DefaultCon   = "qeeniao_daodao" // 默认使用连接的key名
	DaodaoCon    = "qeeniao_daodao"
	SpendEarnCon = "qeeniao_spendearn"
)

func InitMysql(debug bool) {
	models = make(map[string]map[string]*gorm.DB)

	for _, item := range MysqlConnects {
		model, err := gorm.Open("mysql", item.Config.FormatDSN())
		if err != nil {
			panic(err)
		}
		model.DB().SetMaxIdleConns(item.MaxIdleConns)
		model.DB().SetMaxOpenConns(item.MaxOpenConns)
		model.LogMode(debug)
		model.SetLogger(&logger{})

		if _, ok := models[item.Config.DBName]; !ok {
			models[item.Config.DBName] = make(map[string]*gorm.DB)
		}
		if item.IsWrite {
			models[item.Config.DBName][Write] = model
		} else {
			models[item.Config.DBName][Proxy] = model
		}
	}
	for _, v := range models {
		if _, ok := v[Write]; !ok {
			v[Write] = v[Proxy]
		}
	}
	Model = models[DefaultCon][Proxy]
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
