package dbHelper

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DbName string

//这里存放的是一些封装
type MysqlConfig struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       int64  `json:"port"`
	logMode    bool   `json:"log_mode"`
	MaxOpenCon int    `json:"max_open_con"`
	MaxIdleCon int    `json:"max_idle_con"`
	Dbname     string `json:"dbname"`
}

var mysqlConfig *MysqlConfig
var _db_list map[DbName]*gorm.DB

func InitDbHelper(mysqlCfg *MysqlConfig) *gorm.DB {
	//初始化全局sql连接

	mysqlConfig = mysqlCfg
	mysqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Dbname)
	//连接MYSQL
	db, err := gorm.Open("mysql", mysqlDsn)
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
	//打开调试模式
	db.LogMode(mysqlCfg.logMode)

	//设置数据库连接池参数
	db.DB().SetMaxOpenConns(mysqlCfg.MaxOpenCon) //设置数据库连接池最大连接数
	db.DB().SetMaxIdleConns(mysqlCfg.MaxIdleCon) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。

	return db
}

func InitDbList(mysqlCfg map[DbName]*MysqlConfig) {
	//初始化全局sql连接
	for dbName, config := range mysqlCfg {
		db := InitDbHelper(config)
		_db_list[dbName] = db
	}
}

//获取gorm db对象，其他包需要执行数据库查询的时候，只要通过tools.getDB()获取db对象即可。
//不用担心协程并发使用同样的db对象会共用同一个连接，db对象在调用他的方法的时候会从数据库连接池中获取新的连接
func GetDBByName(name DbName) *gorm.DB {
	if m, ok := _db_list[name]; ok {
		return m
	}
	return nil
}

func GetDBList() map[DbName]*gorm.DB {
	return _db_list
}
