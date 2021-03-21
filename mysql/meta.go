package mysql

import (
	"strconv"
	"strings"
)

type Field struct {
	Field   string `gorm:"column:Field"`
	Type    string `gorm:"column:Type"`
	Null    string `gorm:"column:Null"`
	Key     string `gorm:"column:Key"`
	Default string `gorm:"column:Default"`
	Extra   string `gorm:"column:Extra"`
}

type Index struct {
	Table        string `gorm:"column:Table"`
	NonUnique    int    `gorm:"column:Non_unique"`
	KeyName      string `gorm:"column:Key_name"`
	SeqInIndex   int    `gorm:"column:Seq_in_index"`
	ColumnName   string `gorm:"column:Column_name"`
	Collation    string `gorm:"column:Collation"`
	Cardinality  int    `gorm:"column:Cardinality"`
	SubPart      string `gorm:"column:Sub_part"`
	Packed       string `gorm:"column:Packed"`
	Null         string `gorm:"column:Null"`
	IndexType    string `gorm:"column:Index_type"`
	Comment      string `gorm:"column:Comment"`
	IndexComment string `gorm:"column:Index_comment"`
}

type Meta struct {
	Table            string
	AllFieldsDefault map[string]string
	AllFields        []string
	RequiredFields   []string
	UniqueFields     []string
	PrimaryFields    []string
	SplitColumn      string
	SplitType        string
	Sync             int
}

type MetaTable struct {
	Autokid     int64  `gorm:"column:autokid"`
	Tablename   string `gorm:"column:table_name"`   // 表名
	SplitColumn string `gorm:"column:split_column"` // 分表字段
	SplitType   string `gorm:"column:split_type"`   // 分表类型
	Sync        int    `gorm:"column:sync"`         // 支持同步
}

func (MetaTable) TableName() string {
	return "meta_table"
}

type splitCfg struct {
	SpType string // 分表类型
	SpNum  int    //
	SpCol  string // 分表依据字段
}

var splits *map[string]*splitCfg

func GetSplitData() map[string]*splitCfg {
	GetMetaData() //如果没有初始化，则初始化
	return *splits
}

//申明为指针，若需要定时更新时，更细指针即可，不会导致多线程访问全局变量的问题
var metas *map[string]*Meta

func GetMetaData() map[string]*Meta {
	if metas == nil {
		readMeta()
	}
	return *metas
}

func newMeta() *Meta {
	meta := new(Meta)
	meta.AllFieldsDefault = make(map[string]string)
	meta.AllFields = make([]string, 0)
	meta.RequiredFields = make([]string, 0)
	meta.UniqueFields = make([]string, 0)
	meta.PrimaryFields = make([]string, 0)
	return meta
}

// 获取某张表的所有字段、必填（非空）字段、唯一索引、主键
func selectFields(tableName string, backName string) ([]Field, []Field, []Index, []Index) {
	fields := make([]Field, 0)
	requiredFields := make([]Field, 0)
	uniqueIndexs := make([]Index, 0)
	primaryIndexs := make([]Index, 0)

	NewBuild(&fields).Raw("DESC " + tableName)
	NewBuild(&requiredFields).Raw("show columns from `" + tableName + "` WHERE `Null` = 'NO' AND `Default` IS NULL AND `Extra` <> 'auto_increment'")
	NewBuild(&uniqueIndexs).Raw("show INDEX FROM `" + tableName + "` WHERE `Non_unique` = 0 AND `Key_name` <> 'PRIMARY'")
	NewBuild(&primaryIndexs).Raw("show INDEX FROM `" + tableName + "` WHERE `Key_name` = 'PRIMARY'")

	/*	if len(fields) == 0 {
			NewBuild(&fields).Raw("DESC " + backName)
		}
		if len(requiredFields) == 0 {
			NewBuild(&requiredFields).Raw("show columns from `" + backName + "` WHERE `Null` = 'NO' AND `Default` IS NULL AND `Extra` <> 'auto_increment'")
		}
		if len(uniqueIndexs) == 0 {
			NewBuild(&uniqueIndexs).Raw("show INDEX FROM `" + backName + "` WHERE `Non_unique` = 0 AND `Key_name` <> 'PRIMARY'")
		}
		if len(primaryIndexs) == 0 {
			NewBuild(&primaryIndexs).Raw("show INDEX FROM `" + backName + "` WHERE `Key_name` = 'PRIMARY'")
		}*/

	return fields, requiredFields, uniqueIndexs, primaryIndexs
}

func readMeta() {
	metaTables := make([]MetaTable, 0)
	NewBuild(&metaTables).Get()
	store := make(map[string]*Meta)
	splitStore := make(map[string]*splitCfg)
	for _, metaTable := range metaTables {
		meta := newMeta()
		meta.Table = metaTable.Tablename
		meta.SplitColumn = metaTable.SplitColumn
		meta.SplitType = metaTable.SplitType
		meta.Sync = metaTable.Sync

		getTableName := metaTable.Tablename
		if metaTable.SplitColumn != "" {
			getTableName = metaTable.Tablename + "_0"
		}
		fields, requiredFields, uniqueIndexs, primaryIndexs := selectFields(getTableName, metaTable.Tablename)

		for _, field := range fields {
			meta.AllFields = append(meta.AllFields, field.Field)
			if field.Default == "CURRENT_TIMESTAMP" {
				continue
			}
			if field.Extra == "auto_increment" {
				continue
			}
			meta.AllFieldsDefault[field.Field] = field.Default
		}

		for _, field := range requiredFields {
			meta.RequiredFields = append(meta.RequiredFields, field.Field)
		}
		for _, field := range uniqueIndexs {
			meta.UniqueFields = append(meta.UniqueFields, field.ColumnName)
		}
		for _, field := range primaryIndexs {
			meta.PrimaryFields = append(meta.PrimaryFields, field.ColumnName)
		}
		store[metaTable.Tablename] = meta

		//如果该表需要分表，将配置加入分表配置列表中
		if meta.SplitColumn == "" {
			continue
		}
		r := strings.Split(meta.SplitType, ":")
		num, _ := strconv.Atoi(r[1])
		splitStore[meta.Table] = &splitCfg{
			SpType: r[0],
			SpNum:  num,
			SpCol:  meta.SplitColumn,
		}
	}
	splits = &splitStore
	metas = &store
}
