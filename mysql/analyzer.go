package mysql

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/livegoplayer/go_helper/utils"
	"runtime/debug"
	"strings"
	"time"
)

var isOpenExplain bool
var explainLevel = 0 // 0 无，1：警告，2：报错
var explainWhite string

func OpenExplain(level int, white string) {
	isOpenExplain = true
	explainLevel = level
	explainWhite = white
}

type ExpWarn uint

const (
	NotUseIndex ExpWarn = iota
	TooManyRows
	ScanFullTable
	ScanFullIndex
	UseTemporary
	UseFileSort
)

const MaxScanRows int64 = 10000

type Explain struct {
	Id           int64   `gorm:"column:id"             json:"id"`
	SelectType   string  `gorm:"column:select_type"    json:"select_type"`
	Table        string  `gorm:"column:table"          json:"table"`
	Partitions   string  `gorm:"column:partitions"     json:"partitions"`
	Type         string  `gorm:"column:type"           json:"type"`
	PossibleKeys string  `gorm:"column:possible_keys"  json:"possible_keys"`
	Key          string  `gorm:"column:key"            json:"key"`
	KeyLen       string  `gorm:"column:key_len"        json:"key_len"`
	Ref          string  `gorm:"column:ref"            json:"ref"`
	Rows         int64   `gorm:"column:rows"           json:"rows"`
	Filtered     float64 `gorm:"column:filtered"       json:"filtered"`
	Extra        string  `gorm:"column:Extra"          json:"Extra"`
}

var (
	ErrNotUseIndex   = errors.New("未使用索引")
	ErrTooManyRows   = errors.New("扫描行数过多")
	ErrScanFullTable = errors.New("scan full table")
	ErrScanFullIndex = errors.New("scan full index")
	ErrUseTemporary  = errors.New("use temporary")
	ErrUseFileSort   = errors.New("use file sort")
)

type Analyzer struct {
	Explains     []Explain
	preSql       string
	params       []interface{}
	warnIgnores  map[ExpWarn]bool
	tableIgnores []string
}

func NewAnalyzer(sql string, params []interface{}) *Analyzer {
	return &Analyzer{
		preSql: sql,
		params: params,
	}
}

// 忽略部分检测
func (alz *Analyzer) WarnIgnore(ews ...ExpWarn) *Analyzer {
	for _, ew := range ews {
		alz.warnIgnores[ew] = true
	}

	return alz
}

// 数据表白名单，加入白名单的表将不再检查任何问题
func (alz *Analyzer) TableIgnore(table ...string) *Analyzer {
	alz.tableIgnores = table

	return alz
}

// 是否忽略此类问题
func (alz *Analyzer) isIgnore(ew ExpWarn) bool {
	if ign, exist := alz.warnIgnores[ew]; exist && ign {
		return true
	}

	return false
}

// 执行 sql 分析
func (alz *Analyzer) Analyze() error {
	db := NewBuild(&alz.Explains).TableName("disable").Raw("EXPLAIN "+alz.preSql, alz.params...)
	if db.Error != nil {
		return db.Error
	}

	for _, exp := range alz.Explains {
		if utils.StrInSlice(alz.tableIgnores, exp.Table) {
			continue
		}

		if !alz.isIgnore(TooManyRows) && exp.Rows > MaxScanRows {
			return ErrTooManyRows
		}

		// if !regexp.MustCompile("^(update|UPDATE)").MatchString(alz.preSql) {
		if !alz.isIgnore(UseTemporary) && strings.Contains(exp.Extra, "Using temporary") {
			return ErrUseTemporary
		}
		// }

		if !alz.isIgnore(UseFileSort) && strings.Contains(exp.Extra, "Using filesort") {
			return ErrUseFileSort
		}

		if !alz.isIgnore(NotUseIndex) && (exp.Rows > 0 && exp.Key == "") {
			return ErrNotUseIndex
		}

		if !alz.isIgnore(ScanFullTable) && strings.ToUpper(exp.Type) == "ALL" {
			return ErrScanFullTable
		}

		if !alz.isIgnore(ScanFullIndex) && strings.ToUpper(exp.Type) == "INDEX" {
			return ErrScanFullIndex
		}

		// TODO:// 完善检查
		// fmt.Println(exp)
	}

	return nil
}

// gorm 插件
func AnalyzerCallback(scope *gorm.Scope) {
	// TODO:// 判断 ini 配置是否启用

	preSql := strings.TrimSpace(scope.SQL)
	if strings.ToUpper(preSql[:len("EXPLAIN")]) == "EXPLAIN" {
		return
	}

	alz := NewAnalyzer(scope.SQL, scope.SQLVars)
	err := alz.Analyze()

	// TODO:// panic 前 dump 出相关 sql 信息
	if err != nil {
		panic(err)
	}

	// 检查查询时间
}

func AnalyzerCallbackV2(sql string, vars []interface{}, dura time.Duration) {
	if !isOpenExplain {
		return
	}

	stack := debug.Stack() // 来源堆栈
	go func() {
		defer func() {
			if e := recover(); e != nil {
				Logger.Info("mysql-slow-error: " + string(stack) + e.(error).Error())
			}
		}()

		preSql := strings.TrimSpace(sql)
		if strings.ToUpper(preSql[:len("EXPLAIN")]) == "EXPLAIN" {
			return
		}

		if strings.ToUpper(preSql[:len("SHOW")]) == "SHOW" {
			return
		}

		alz := NewAnalyzer(sql, vars).TableIgnore(strings.Split(explainWhite, ",")...)
		err := alz.Analyze()
		if err == nil {
			return
		}
		id := time.Now().UnixNano()
		Logger.Error("慢数据查询:" + sql)
		if explainLevel == 2 {
			panic("mysql-slow-" + utils.AsString(id))
		}
	}()
}
