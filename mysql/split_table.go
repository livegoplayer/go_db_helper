package mysql

import (
	"strconv"
)

//用于反射判断是否实现了分表功能
type splitInterface interface {
	BaseName()
	setSplitVal(int64)
	load()
	splitIdx(int64) int64
	getSplitCol() string
	isCanSplit() bool
	TableName() string
}

type SplitTable struct {
	spType        string `gorm:"-"    json:"-"`
	spNum         int    `gorm:"-"    json:"-"`
	spColVal      int64  `gorm:"-"    json:"-"`
	spColName     string `gorm:"-"    json:"-"`
	isSetVal      bool   `gorm:"-"    json:"-"`
	BaseTableName string `gorm:"-"    json:"-"`
}

//加载配置，并且返回根据配置生成的分表表名, 如需使用，必须外部先调用setSplitVal
func (sp *SplitTable) TableName() string {
	sp.load()
	if !sp.isSetVal {
		panic("调用分表未设置分表key table:" + sp.BaseTableName)
	}
	return sp.BaseTableName + "_" + strconv.FormatInt(sp.spColVal%int64(sp.spNum), 10)
}

//获取分表的序列
func (sp *SplitTable) splitIdx(val int64) int64 {
	return val % int64(sp.spNum)
}

func (sp *SplitTable) setSplitVal(val int64) {
	sp.isSetVal = true
	sp.spColVal = val
}

func (sp *SplitTable) isCanSplit() bool {
	return sp.isSetVal
}

func (sp *SplitTable) getSplitCol() string {
	return sp.spColName
}

//将配置表中的数据读入结构中
func (sp *SplitTable) load() {
	cfg, ok := GetSplitData()[sp.BaseTableName]
	if !ok {
		panic("分表配置有问题 table:" + sp.BaseTableName)
	}
	sp.spType = cfg.SpType
	sp.spNum = cfg.SpNum
	sp.spColName = cfg.SpCol
}

func SplitTableName(baseName string, val int64) string {
	split := SplitTable{BaseTableName: baseName}
	split.load()
	split.setSplitVal(val)
	return split.TableName()
}
