package mysql

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/livegoplayer/go_helper/utils"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type wheres struct {
	tp       string
	column   string
	operator string
	value    interface{}
	values   []interface{}
	boolean  string
	sql      string
}

type joins struct {
	tp     string
	table  string
	one    string
	op     string
	two    string
	wheres []wheres
}

type order struct {
	column    string
	direction string
}

type Build struct {
	model         interface{}
	table         string
	wheres        []wheres
	bindings      []interface{}
	joins         []joins
	selects       []string
	group         string
	sets          utils.H
	orders        []order
	offset        int
	limit         int
	err           error
	notPanic      bool   // 当sql执行不成功时，是否抛出错误
	effectRow     int64  // 执行sql影响的行数
	isSetSplitVal bool   // 是否设置过分表值，因为splitVal有可能设为初始值0
	splitVal      int64  // 分表值
	con           string // 指定使用的连接
	isOnWrite     bool   // 是否强制读主库
	unscoped      bool
	operate       *operat
}
type raw struct {
	sql string
}

func Raw(sql string) *raw {
	return &raw{sql}
}

// 构建where语句的回调
type WhereCb func(build *Build)

type tabler interface {
	TableName() string
}

func NewBuild(model interface{}) *Build {
	build := new(Build)
	build.selects = []string{}

	if str, ok := model.(string); ok {
		// 指定表名
		build.TableName(str)
	} else {
		build.model = model
	}
	build.wheres = []wheres{}
	build.sets = make(utils.H)
	build.bindings = []interface{}{}
	build.operate = newOperat()
	return build
}

// 复制当前的内容出一个新的build对象
func (build *Build) Clone() *Build {
	return &Build{
		model:         build.model,
		table:         build.table,
		wheres:        build.wheres,
		bindings:      build.bindings,
		joins:         build.joins,
		selects:       build.selects,
		group:         build.group,
		sets:          build.sets,
		orders:        build.orders,
		offset:        build.offset,
		limit:         build.limit,
		err:           build.err,
		notPanic:      build.notPanic,
		effectRow:     build.effectRow,
		isSetSplitVal: build.isSetSplitVal,
		splitVal:      build.splitVal,
		con:           build.con,
		isOnWrite:     build.isOnWrite,
		unscoped:      build.unscoped,
	}
}

func (build *Build) Unscoped() *Build {
	build.unscoped = true
	return build
}

func (build *Build) Select(columns ...interface{}) *Build {
	for _, str := range columns {
		switch str.(type) {
		case string:
			s := str.(string)
			if s[0:1] == "`" {
				build.selects = append(build.selects, s)
			} else {
				build.selects = append(build.selects, "`"+s+"`")
			}
		default:
			if r, ok := str.(*raw); ok {
				build.selects = append(build.selects, r.sql)
			}
		}
	}
	return build
}

func (build *Build) Where(column string, args ...interface{}) *Build {
	var operator string
	var value interface{}
	if len(args) == 1 {
		operator = "="
		value = args[0]
	} else if len(args) > 1 {
		operator = args[0].(string)
		value = args[1]
	}
	build.addWheres(wheres{operator: operator, tp: "Basic", value: value, column: column, boolean: "and"})
	return build
}

func (build *Build) WhereMap(datas map[string]interface{}) *Build {
	for key, val := range datas {
		build.Where(key, val)
	}
	return build
}

func (build *Build) OrWhere(column string, args ...interface{}) *Build {
	var operator string
	var value interface{}
	if len(args) == 1 {
		operator = "="
		value = args[0]
	} else {
		operator = args[0].(string)
		value = args[1]
	}
	build.addWheres(wheres{operator: operator, tp: "Basic", value: value, column: column, boolean: "or"})
	return build
}

func (build *Build) OrWhereCb(cb WhereCb) *Build {
	build.addWheres(wheres{tp: "cb", boolean: "or", value: cb})
	return build
}

func (build *Build) WhereCb(cb WhereCb) *Build {
	build.addWheres(wheres{tp: "cb", boolean: "and", value: cb})
	return build
}

func (build *Build) WhereRaw(sql string, bindings ...interface{}) *Build {
	if sql == "" {
		return build
	}
	build.addWheres(wheres{tp: "raw", sql: sql, boolean: "and", values: bindings})
	return build
}

func (build *Build) WhereNull(column string) *Build {
	return build.WhereRaw(column + " is NULL")
}

func (build *Build) WhereNotNull(column string) *Build {
	return build.WhereRaw(column + " IS NOT NULL")
}

func (build *Build) OrWhereNull(column string) *Build {
	build.addWheres(wheres{tp: "raw", sql: column + " is NULL", boolean: "or"})
	return build
}

func (build *Build) OrWhereNotNull(column string) *Build {
	build.addWheres(wheres{tp: "raw", sql: column + " is NOT NULL", boolean: "or"})
	return build
}

func (build *Build) WhereIn(column string, values interface{}) *Build {
	build.addWheres(wheres{tp: "In", value: values, column: column, boolean: "and"})
	return build
}

func (build *Build) WhereNotIn(column string, values interface{}) *Build {
	build.addWheres(wheres{tp: "NotIn", value: values, column: column, boolean: "and"})
	return build
}

func (build *Build) join(table string, one string, op string, two string, joinType string) *Build {
	build.joins = append(build.joins, joins{tp: joinType, table: table, one: one, op: op, two: two})
	return build
}

func (build *Build) LeftJoin(table string, one string, args ...interface{}) *Build {
	var op, two string
	if len(args) == 0 {
		op = "="
		two = one
	} else if len(args) == 1 {
		op = "="
		two = args[0].(string)
	} else {
		op = args[0].(string)
		two = args[1].(string)
	}
	build.join(table, one, op, two, "left")
	return build
}

func (build *Build) RightJoin(table string, one string, args ...interface{}) *Build {
	var op, two string
	if len(args) == 0 {
		op = "="
		two = one
	} else if len(args) == 1 {
		op = "="
		two = args[0].(string)
	} else {
		op = args[0].(string)
		two = args[1].(string)
	}
	build.join(table, one, op, two, "right")
	return build
}

func (build *Build) InnerJoin(table string, one string, args ...interface{}) *Build {
	var op, two string
	if len(args) == 0 {
		op = "="
		two = one
	} else if len(args) == 1 {
		op = "="
		two = args[0].(string)
	} else {
		op = args[0].(string)
		two = args[1].(string)
	}
	build.join(table, one, op, two, "inner")
	return build
}

func (build *Build) OrderBy(column string, direction string) *Build {
	build.orders = append(build.orders, order{column: column, direction: strings.ToLower(direction)})
	return build
}

func (build *Build) OrderDescBy(column string) *Build {
	return build.OrderBy(column, "desc")
}

func (build *Build) OrderAscBy(column string) *Build {
	return build.OrderBy(column, "asc")
}

func (build *Build) Group(column string) *Build {
	build.group = column
	return build
}

func (build *Build) Inc(col string, num int) *Build {
	build.operate.cumulateInc(col, num)
	return build
}

func (build *Build) Set(col string, val interface{}) *Build {
	build.operate.cumulateSet(col, val)
	return build
}

func (build *Build) Skip(lines int) *Build {
	build.offset = lines
	return build
}

func (build *Build) Offset(lines int) *Build {
	return build.Skip(lines)
}

func (build *Build) Limit(lines int) *Build {
	build.limit = lines
	return build
}

func (build *Build) Take(lines int) *Build {
	return build.Limit(lines)
}

func (build *Build) First() bool {
	g := build.newGorm().Where(build.whereQuery(), build.bindings...)
	g = build.joinQuery(g)
	g = build.findQuery(g)
	g = g.First(build.model)
	build.dealError(g)
	if g.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (build *Build) Get() int64 {
	g := build.newGorm().Where(build.whereQuery(), build.bindings...)
	g = build.joinQuery(g)
	g = build.findQuery(g)
	g = g.Find(build.model)
	build.dealError(g)
	return g.RowsAffected
}

func (build *Build) Count() int64 {
	return int64(build.math("*", "Count"))
}

func (build *Build) Sum(col string) float64 {
	return build.math(col, "Sum")
}

func (build *Build) Max(col string) float64 {
	return build.math(col, "MAX")
}

func (build *Build) math(col string, method string) float64 {
	type scan struct {
		Ret float64
	}
	s := new(scan)
	g := build.newGorm().Select(method+"("+col+") as ret").Where(build.whereQuery(), build.bindings...)
	g = build.joinQuery(g)
	g = build.findQuery(g)
	g = g.Scan(s)
	build.dealError(g)
	return s.Ret
}

func (build *Build) Update(h utils.H) int64 {
	g := build.newGorm().Where(build.whereQuery(), build.bindings...).Updates(h)
	build.dealError(g)
	return g.RowsAffected
}

func (build *Build) Save() {
	g := build.newGorm().Save(build.model)
	build.dealError(g)
}

// 支持分表
func (build *Build) Insert(argu interface{}) {
	h, ok := argu.([]map[string]interface{})
	if !ok {
		x, ok := argu.(map[string]interface{})
		if !ok {
			panic("数据类型错误")
		}
		h = []map[string]interface{}{x}
	}
	if len(h) == 0 {
		return
	}
	ele := h[0]
	keys := make([]string, len(ele))
	i := 0
	for k := range ele {
		keys[i] = k
		i++
	}
	if build.table != "" {
		// 已经指定表名了
		bindings := formatBindings(h, keys)
		build.Exec(build.insertSql(keys, len(h)), bindings...)
		return
	}
	split, tab := build.isSplit()
	sort.Strings(keys)
	if tab != nil {
		// 非分表
		bindings := formatBindings(h, keys)
		build.Exec(build.insertSql(keys, len(h)), bindings...)
		return
	}
	split.BaseName()
	split.load()

	// 分拣数据
	splitData := make(map[int64][]map[string]interface{})
	for _, item := range h {
		v, ok := item[split.getSplitCol()]
		if !ok {
			panic("插入数据缺少分表值")
		}
		idx := split.splitIdx(utils.AsInt64(v))
		_, ok = splitData[idx]
		if !ok {
			splitData[idx] = make([]map[string]interface{}, 0)
		}
		splitData[idx] = append(splitData[idx], item)
	}
	for idx, list := range splitData {
		bindings := formatBindings(list, keys)
		b := NewBuild(build.model).SplitBy(idx)
		b.Exec(b.insertSql(keys, len(list)), bindings...)
	}
}

func (build *Build) Delete() int64 {
	softer := build.softDeleted()
	if softer == nil {
		g := build.newGorm().Where(build.whereQuery(), build.bindings...).Delete(&build.model)
		build.dealError(g)
		return g.RowsAffected
	}
	col, _, delVal := softer.SoftDeleted()
	return build.Update(utils.H{col: delVal})
}

// 将刚刚积累的操作，一次性执行到数据库
func (build *Build) DoneOperate() int64 {
	op := build.operate.raw()
	if len(op) == 0 {
		return 0
	}
	return build.Update(op)
}

func (build *Build) Increment(column string, amount int) int64 {
	g := build.newGorm().Where(build.whereQuery(), build.bindings...).
		Update(column, gorm.Expr(column+" + ?", amount))
	build.dealError(g)
	return g.RowsAffected
}

func (build *Build) Raw(sql string, values ...interface{}) *gorm.DB {
	g := build.newCon().Model(build.model).Raw(sql, values...).Scan(build.model)
	build.dealError(g)
	return g
}

func (build *Build) Exec(sql string, values ...interface{}) *gorm.DB {
	g := build.newCon().Model(build.model).Exec(sql, values...)
	build.dealError(g)
	return g
}

func (build *Build) ForUpdate() *Build {
	build.sets["gorm:query_option"] = "FOR UPDATE"
	return build
}

// 当需要分表，但是在where中并没有调用分表key的查询，则需要手动执行该函数判断分表
func (build *Build) SplitBy(val int64) *Build {
	build.isSetSplitVal = true
	build.splitVal = val
	return build
}

// 指定表名
func (build *Build) TableName(name string) *Build {
	build.table = name
	return build
}

// 指定model类
func (build *Build) ModelType(v interface{}) *Build {
	build.model = v
	return build
}

// 调用此函数后，数据库的所有错误都不再抛出，需要使用者主动调用Error获取错误处理
func (build *Build) DisablePanic() *Build {
	build.notPanic = true
	return build
}

// 获取sql执行的错误, 需要先调用DisablePanic
func (build *Build) Error() error {
	return build.err
}

// 用给定的值构建出sql, 和绑定值
func (build *Build) BuildInsertSql(argu interface{}) (string, []interface{}) {
	h, ok := argu.([]map[string]interface{})
	if !ok {
		x, ok := argu.(map[string]interface{})
		if !ok {
			panic("数据类型错误")
		}
		h = []map[string]interface{}{x}
	}
	if len(h) == 0 {
		return "", []interface{}{}
	}
	ele := h[0]
	keys := make([]string, len(ele))
	i := 0
	for k := range ele {
		keys[i] = k
		i++
	}
	bindings := formatBindings(h, keys)
	sql := build.insertSql(keys, len(h))
	return sql, bindings
}

// 使用配置中IsWrite的连接，若不存在，则使用proxy连接
func (build *Build) OnWrite() *Build {
	build.isOnWrite = true
	return build
}

// 获取一个通用数据库句柄
func (build *Build) DB() *sql.DB {
	return build.newCon().DB()
}

// Private func 以下为私有接口

func (build *Build) addWheres(whe wheres) {
	build.wheres = append(build.wheres, whe)
}

func (build *Build) addBinding(values interface{}) {
	build.bindings = append(build.bindings, values)
}

func (build *Build) newCon() *gorm.DB {
	var g *gorm.DB
	gid := getSlow()
	con := build.getConName()
	h := getHandle(gid, con).trans
	if h == nil {
		if build.isOnWrite {
			if _, ok := models[con][Write]; !ok {
				g = models[con][Write]
			} else {
				g = models[con][Proxy]
			}
		} else {
			g = models[con][Proxy]
		}
	} else {
		g = h.db
	}
	return g
}

func (build *Build) newGorm() *gorm.DB {
	g := build.newCon().Model(build.model).Table(build.tableName()).Unscoped()
	for key, set := range build.sets {
		g = g.Set(key, set)
	}

	// 开发环境注册 Analyzer 插件
	// if config.Development {
	//    g.Callback().Query().Register("Analyzer", AnalyzerCallback)
	//    // g.Callback().Query().Before("gorm:query").Register("Analyzer", AnalyzerCallback)
	// }

	return g
}

func (build *Build) whereSql(whe wheres) (string, interface{}, bool) {
	switch whe.tp {
	case "Basic":
		return "`" + whe.column + "` " + whe.operator + " ?", whe.value, false
	case "raw":
		return whe.sql, whe.values, true
	case "In":
		return "`" + whe.column + "` in (?)", whe.value, false
	case "NotIn":
		return "`" + whe.column + "` not in (?)", whe.value, false
	case "cb":
		b := NewBuild(build.model)
		whe.value.(WhereCb)(b)
		return b.whereQuery(), b.bindings, true
	}
	panic("操作类型错误")
}

// 不太好用
func (build *Build) joinQuery(g *gorm.DB) *gorm.DB {
	name := build.tableName()
	for _, join := range build.joins {
		var o, t, query string
		if !strings.Contains(join.one, ".") && !strings.Contains(strings.ToLower(name), " as ") {
			o = name + "." + join.one
		} else {
			o = join.one
		}
		if !strings.Contains(join.two, ".") && !strings.Contains(strings.ToLower(join.table), " as ") {
			t = join.table + "." + join.two
		} else {
			t = join.two
		}
		query += join.tp + " join " + join.table + " on " + o + " " + join.op + " " + t
		g = g.Joins(query)
	}
	return g
}

func (build *Build) whereQuery() string {
	query := ""

	// 添加软删除的查询条件
	if !build.unscoped && build.softDeleted() != nil {
		softer := build.softDeleted()
		col, liveVal, _ := softer.SoftDeleted()
		if liveVal == nil {
			build.WhereNull(col)
		} else {
			build.Where(col, liveVal)
		}

	}

	for idx, whe := range build.wheres {
		sql, val, isMultiBind := build.whereSql(whe)
		boolean := whe.boolean
		if idx == 0 {
			boolean = ""
		}
		query += boolean + " (" + sql + ") "
		if val != nil {
			if isMultiBind {
				for _, v := range val.([]interface{}) {
					build.addBinding(v)
				}
			} else {
				build.addBinding(val)
			}
		}
	}
	return query
}

func (build *Build) findQuery(g *gorm.DB) *gorm.DB {
	if len(build.selects) != 0 {
		g = g.Select(build.selects)
	}
	if build.limit != 0 {
		g = g.Limit(build.limit)
	}

	if build.offset != 0 {
		g = g.Offset(build.offset)
	}
	if build.group != "" {
		g = g.Group(build.group)
	}
	for _, order := range build.orders {
		g = g.Order(order.column + " " + order.direction)
	}
	return g
}

func (build *Build) insertSql(keys []string, lines int) string {
	keysSql := strings.Join(keys, "`,`")
	valueChar := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		valueChar[i] = "?"
	}
	valueSql := strings.Join(valueChar, ",")
	valuesStr := make([]string, lines)
	for k := 0; k < lines; k++ {
		valuesStr[k] = "(" + valueSql + ")"
	}
	valuesSql := strings.Join(valuesStr, ",")
	return "INSERT " + build.tableName() + " (`" + keysSql + "`) VALUES " + valuesSql
}

// 根据model判断是否是分表, 两个返回值中，只有一个有值
func (build *Build) isSplit() (splitInterface, tabler) {
	reflectType := reflect.ValueOf(build.model).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	tab, ok := reflect.New(reflectType).Interface().(splitInterface)
	if !ok {
		return nil, reflect.New(reflectType).Interface().(tabler)
	}
	return tab, nil
}

// tableName 返回表名，如果查询的是分表，但是却又没有设置分表key的值，则抛出异常
func (build *Build) tableName() string {
	if build.table != "" {
		return build.table
	}
	tab, other := build.isSplit()
	if other != nil {
		return other.TableName()
	}
	tab.BaseName()
	tab.load()

	// 如果外部已经手动设置过分表值，则直接设置
	if build.isSetSplitVal {
		tab.setSplitVal(build.splitVal)
		return tab.TableName()
	}
	for _, whe := range build.wheres {
		if whe.tp != "Basic" {
			continue
		}
		if whe.operator != "=" {
			continue
		}
		if whe.column != tab.getSplitCol() {
			continue
		}
		var v int64
		switch whe.value.(type) {
		case int:
			v = int64(whe.value.(int))
		case int64:
			v = whe.value.(int64)
		case float64:
			v = int64(whe.value.(float64))
		case string:
			id, err := strconv.ParseInt(whe.value.(string), 10, 64)
			if err != nil {
				panic("分表值类型错误," + err.Error())
			}
			v = id
		default:
			panic("分表值类型错误")
		}
		tab.setSplitVal(v)
		return tab.TableName()
	}
	panic("分表查询未设置分表值")
}

func (build *Build) dealError(g *gorm.DB) {
	if g.Error == nil {
		return
	}
	if g.Error.Error() == "record not found" {
		// 忽略没有找到数据的错误
		return
	}
	if !build.notPanic {
		panic(g.Error.Error())
	} else {
		build.err = g.Error
	}
}

// 工具函数
func formatBindings(list []map[string]interface{}, keys []string) []interface{} {
	bindings := make([]interface{}, len(list)*len(keys))
	m := 0
	for _, item := range list {
		for _, k := range keys {
			v, ok := item[k]
			if !ok {
				panic("插入数据缺少key:" + k)
			}
			bindings[m] = v
			m++
		}
	}
	return bindings
}

// 获取连接名
func (build *Build) getConName() string {
	type conInterface interface {
		Connect() string
	}

	if build.con != "" {
		return build.con
	}

	if build.model == nil {
		return DefaultCon
	}
	reflectType := reflect.ValueOf(build.model).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	_, ok := reflect.New(reflectType).Interface().(conInterface)
	if ok {
		return reflect.New(reflectType).Interface().(conInterface).Connect()
	}
	return DefaultCon
}

type softer interface {
	SoftDeleted() (string, interface{}, interface{})
}

// 获取软删除的设置
func (build *Build) softDeleted() softer {
	if !reflect.ValueOf(build.model).IsValid() {
		return nil
	}
	reflectType := reflect.ValueOf(build.model).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	s, ok := reflect.New(reflectType).Interface().(softer)
	if ok {
		return s
	}
	return nil
}
