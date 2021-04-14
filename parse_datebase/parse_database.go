package parse_datebase

import (
	"bytes"
	"flag"
	"github.com/livegoplayer/go_db_helper/mysql"
	mysqlHelper "github.com/livegoplayer/go_db_helper/mysql"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/***
原理: show full fields from tablename; 可以看到表的所有字段的详细信息
     show  create  table  tablename; 可以查看表注释
*/

var OUTPUT_PATH_NAME = ""

type Field struct {
	Field            string `gorm:"column:Field"`
	Type             string `gorm:"column:Type"`
	Collation        string `gorm:"column:Collation"`
	Null             string `gorm:"column:Null"`
	Key              string `gorm:"column:Key"`
	Default          string `gorm:"column:Default"`
	Extra            string `gorm:"column:Extra"`
	Privileges       string `gorm:"column:Privileges"`
	Comment          string `gorm:"column:Comment"`
	Name             string `gorm:"-"`
	Tagstr           string `gorm:"-"`
	ColumnPrimaryStr string `gorm:"-"`
}

type Table struct {
	Name          string  `gorm:"column:Name"`
	TableName     string  `gorm:"-" json:"table_name"`
	PackageName   string  `gorm:"-" json:"package_name"`
	Prefix        string  `gorm:"column:prefix" json:"prefix"`
	Engine        string  `gorm:"column:Engine"`
	Version       string  `gorm:"column:Version"`
	RowFormat     string  `gorm:"column:Row_format"`
	Rows          string  `gorm:"column:Rows"`
	AvgRowLengt   string  `gorm:"column:Avg_row_length"`
	DataLength    string  `gorm:"column:Data_length"`
	MaxDataLength string  `gorm:"column:Max_data_length"`
	IndexLength   string  `gorm:"column:Index_length"`
	DataFree      string  `gorm:"column:Data_free"`
	AutoIncrement string  `gorm:"column:Auto_increment"`
	CreateTime    string  `gorm:"column:Create_time"`
	UpdateTime    string  `gorm:"column:Update_time"`
	CheckTime     string  `gorm:"column:Check_time"`
	Collation     string  `gorm:"column:Collation"`
	Checksum      string  `gorm:"column:Checksum"`
	CreateOptions string  `gorm:"column:Create_options"`
	Comment       string  `gorm:"column:Comment"`
	DbName        string  `gorm:"-"`
	IsSplit       bool    `gorm:"-" json:"is_split"`
	Fields        []Field `gorm:"-" json:"fields"`
}

// too 从这里可以修改默认的数据库
func (t *Table) Connect() string {
	return string(mysqlHelper.DefaultDbName)
}

var Name string
var Prefix string
var Path string

func init() {
	flag.StringVar(&Name, "table", "", "table")
	flag.StringVar(&Prefix, "prefix", "", "prefix")
	flag.StringVar(&OUTPUT_PATH_NAME, "out", "", "out")
	flag.Parse()
}

func Parse(APPRoot string) {
	if OUTPUT_PATH_NAME == "" {
		OUTPUT_PATH_NAME = "repository_template_model"
	}

	fields := make([]Field, 0)
	tables := make([]Table, 0)
	sql := "show table status "
	if len(Name) > 0 {
		sql += "where name = \"" + Name + "\""
	}
	mysql.NewBuild(&tables).TableName("disable").Raw(sql)

	for k, t := range tables {
		mysql.NewBuild(&fields).TableName("disable").Raw("show full fields from " + t.Name)
		tables[k].Fields = fields
		tables[k].Prefix = Prefix
	}

	currentDirPath := APPRoot
	Path = filepath.Join(currentDirPath, OUTPUT_PATH_NAME)
	// 新建目录
	_ = os.Mkdir(Path, os.ModeDir)
	err := os.Chmod(Path, os.ModePerm)
	if err != nil {
		panic(err)
	}

	maxKeyLength := 0
	for _, t := range tables {
		t.DbName = t.Connect()
		for i, field := range t.Fields {
			tp := mysql.FieldTypeToGolangType(field.Type)
			if tp == "" {
				panic("无法识别类型, key:" + field.Field)
			}
			t.Fields[i].Type = tp
			t.Fields[i].Name = UnderLineToCamel(field.Field)
			if len(field.Field) > maxKeyLength {
				maxKeyLength = len(field.Field)
			}
		}

		for k, field := range fields {
			if field.Key == "PRI" {
				fields[k].ColumnPrimaryStr = ";PRIMARY_KEY"
			}
		}

		name := Name
		if name == "" {
			name = t.Name
		}
		regs := regexp.MustCompile("^(.+)_[0-9]+$").FindStringSubmatch(name)
		if len(regs) == 0 {
			t.TableName = UnderLineToCamel(GetNameWithoutPrefix(name, t.Prefix))
			t.PackageName = GetNameWithoutPrefix(name, t.Prefix)
			t.Name = GetNameWithoutPrefix(name, t.Prefix)
			t.IsSplit = false
		} else {
			t.TableName = UnderLineToCamel(GetNameWithoutPrefix(regs[1], t.Prefix))
			t.PackageName = GetNameWithoutPrefix(regs[1], t.Prefix)
			t.Name = GetNameWithoutPrefix(regs[1], t.Prefix)
			t.IsSplit = true
		}

		var doc bytes.Buffer
		tm, err := template.New("create_model").Parse(temp)
		if err != nil {
			panic(err)
		}
		err = tm.Execute(&doc, t)
		if err != nil {
			panic(err.Error())
		}
		html := doc.String()
		html = regexp.MustCompile("&#34;").ReplaceAllString(html, "\"")

		pdir := filepath.Join(Path, t.PackageName)
		_ = os.Mkdir(pdir, os.ModeDir)
		err = os.Chmod(pdir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		p := filepath.Join(pdir, t.TableName+".go")
		// 新建目录
		err = ioutil.WriteFile(p, []byte(html), 0644)
		if err != nil {
			panic(err)
		}
	}

}

var temp = `package {{.PackageName}}

import "github.com/livegoplayer/go_db_helper/mysql"

const PREFIX = "{{.Prefix}}"

// {{.Comment}}
type {{.TableName}} struct {   {{range .Fields}}
	{{.Name}}  {{.Type}} ` + "`" + `gorm:"column:{{.Field}}{{.ColumnPrimaryStr}}" ` + `json:"{{.Field}}"` + "`" + ` // {{.Comment}} {{end}} 
}
{{if .IsSplit}}
func (t * {{.TableName}}) BaseName() {
	t.BaseTableName = PREFIX + "{{.Name}}"
}
{{else}}
func ({{.TableName}}) TableName() string {
	return PREFIX + "{{.Name}}"
}
{{end}}
type {{.TableName}}Query struct {
	mysql.Query
}

func ({{.TableName}}) Connect() string {
	return "{{.DbName}}"
}
`

func GetNameWithoutPrefix(name, prefix string) string {
	str := strings.TrimPrefix(name, prefix)
	return str
}

// 将下划线风格的单词变为驼峰命名的单词
func UnderLineToCamel(line string) string {
	words := strings.Split(line, "_")
	n := ""
	for _, w := range words {
		n += strings.ToUpper(w[0:1]) + w[1:]
	}
	return n
}
