package dbHelper

import (
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	m "github.com/livegoplayer/go_db_helper/mysql"
	"github.com/livegoplayer/go_helper/utils"
	"github.com/sirupsen/logrus"
)

type DbName string
type DbType string

const (
	WRITE DbType = "write"
	READ  DbType = "read"
)

//这里存放的是一些封装
type MysqlConfig struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	Host                 string `json:"host"`
	Port                 int64  `json:"port"`
	Collation            string `json:"collation"`
	LogMode              bool   `json:"log_mode"`
	Net                  string `json:"net"`
	AllowNativePasswords bool   `json:"allow_native_passwords"`
	MaxOpenCon           int    `json:"max_open_con"`
	MaxIdleCon           int    `json:"max_idle_con"`
	Dbname               string `json:"dbname"`
	IsWrite              bool   `json:"is_write"`
}

var mysqlConfig *MysqlConfig
var _db_list map[DbName]map[DbType]*gorm.DB
var DefaultDbName string

func InitDbHelper(mysqlCfg *MysqlConfig) *gorm.DB {
	//初始化全局sql连接
	mcfg := &mysql.Config{
		User:                 mysqlCfg.Username,
		Passwd:               mysqlCfg.Password,
		Addr:                 mysqlCfg.Host + ":" + utils.AsString(mysqlCfg.Port),
		Collation:            mysqlCfg.Collation,
		Net:                  mysqlCfg.Net,
		AllowNativePasswords: mysqlCfg.AllowNativePasswords,
		DBName:               mysqlCfg.Dbname,
	}
	mysqlDsn := mcfg.FormatDSN()

	//连接MYSQL
	db, err := gorm.Open("mysql", mysqlDsn)
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
	//打开调试模式
	db.LogMode(mysqlCfg.LogMode)

	//设置数据库连接池参数
	db.DB().SetMaxOpenConns(mysqlCfg.MaxOpenCon) //设置数据库连接池最大连接数
	db.DB().SetMaxIdleConns(mysqlCfg.MaxIdleCon) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
	return db
}

func InitDbList(mysqlCfg map[DbName]*MysqlConfig, defaultDbName string) {
	//初始化全局sql连接
	for _, dbConfig := range mysqlCfg {
		if _db_list[DbName(dbConfig.Dbname)] == nil {
			_db_list[DbName(dbConfig.Dbname)] = make(map[DbType]*gorm.DB, 0)
		}
		db := InitDbHelper(dbConfig)

		if dbConfig.IsWrite {
			_db_list[DbName(dbConfig.Dbname)][WRITE] = db
		} else {
			_db_list[DbName(dbConfig.Dbname)][READ] = db
		}

		if defaultDbName != "" {
			DefaultDbName = defaultDbName
		}
	}
}

//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过tools.getDB()获取db对象即可。
//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接
func GetDBByName(name DbName) map[DbType]*gorm.DB {
	if m, ok := _db_list[name]; ok {
		return m
	}
	return nil
}

func GetDBList() map[DbName]map[DbType]*gorm.DB {
	return _db_list
}

func SetLogger(logger *logrus.Logger) {
	if logger != nil {
		m.SetLogger(logger)
	}
}
