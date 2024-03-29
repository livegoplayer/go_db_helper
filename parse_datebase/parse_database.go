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
	"reflect"
	"regexp"
	"strconv"
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
	JsonTagStr       string `gorm:"-"`
	GormTagStr       string `gorm:"-"`
	ColumnPrimaryStr string `gorm:"-"`
}

type TableIndex struct {
	Table        string `gorm:"column:Table"`
	NonUnique    int64  `gorm:"column:Non_unique"`
	KeyName      string `gorm:"column:Key_name"`
	SeqInIndex   int64  `gorm:"column:Seq_in_index"`
	ColumnName   string `gorm:"column:Column_name"`
	Collation    string `gorm:"column:Collation"`
	Cardinality  int64  `gorm:"column:Cardinality"`
	SubPart      string `gorm:"column:Sub_part"`
	Packed       string `gorm:"column:Packed"`
	IndexType    string `gorm:"column:Index_type"`
	Comment      string `gorm:"column:Comment"`
	IndexComment string `gorm:"column:Index_comment"`
}

type Table struct {
	Name            string            `gorm:"column:Name"`
	TableName       string            `gorm:"-" json:"table_name"`
	PackageName     string            `gorm:"-" json:"package_name"`
	Prefix          string            `gorm:"column:prefix" json:"prefix"`
	Engine          string            `gorm:"column:Engine"`
	Version         string            `gorm:"column:Version"`
	RowFormat       string            `gorm:"column:Row_format"`
	Rows            string            `gorm:"column:Rows"`
	AvgRowLengt     string            `gorm:"column:Avg_row_length"`
	DataLength      string            `gorm:"column:Data_length"`
	MaxDataLength   string            `gorm:"column:Max_data_length"`
	IndexLength     string            `gorm:"column:Index_length"`
	DataFree        string            `gorm:"column:Data_free"`
	AutoIncrement   string            `gorm:"column:Auto_increment"`
	CreateTime      string            `gorm:"column:Create_time"`
	UpdateTime      string            `gorm:"column:Update_time"`
	CheckTime       string            `gorm:"column:Check_time"`
	Collation       string            `gorm:"column:Collation"`
	Checksum        string            `gorm:"column:Checksum"`
	CreateOptions   string            `gorm:"column:Create_options"`
	Comment         string            `gorm:"column:Comment"`
	DbName          string            `gorm:"-"`
	IsSplit         bool              `gorm:"-" json:"is_split"`
	Fields          []Field           `gorm:"-" json:"fields"`
	TableIndexs     TableIndexCollect `gorm:"-"`
	TableIndexSlice []Index           `gorm:"-"`
}

/**
@Collect
*/
type TableIndexCollect []TableIndex

type Index struct {
	Name        string
	Type        IndexType
	FieldsSlice []string
}

type IndexType string

const UNI IndexType = "UNIQUE"
const MUTI IndexType = "MUTI"

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
	tableIndexs := make([]TableIndex, 0)
	tables := make([]Table, 0)
	sql := "show table status "
	if len(Name) > 0 {
		sql += "where name = \"" + Name + "\""
	}
	mysql.NewBuild(&tables).TableName("disable").Raw(sql)

	for k, t := range tables {
		mysql.NewBuild(&fields).TableName("disable").Raw("show full fields from " + t.Name)
		mysql.NewBuild(&tableIndexs).TableName("disable").Raw(" SHOW INDEX FROM " + t.Name)

		tables[k].Fields = fields
		tables[k].Prefix = Prefix
		tables[k].TableIndexs = tableIndexs
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

		if t.TableIndexSlice == nil {
			t.TableIndexSlice = make([]Index, 0)
		}

		// 初始化
		for keyName, indexs := range t.TableIndexs.GroupByKeyName() {
			i := Index{
				Name: keyName,
			}

			i.FieldsSlice = make([]string, len(indexs))
			for _, index := range indexs {
				// 如果是unique
				if index.NonUnique == 0 {
					i.Type = UNI
				} else {
					i.Type = MUTI
				}
				i.FieldsSlice[index.SeqInIndex-1] = index.ColumnName
			}

			t.TableIndexSlice = append(t.TableIndexSlice, i)
		}

		newSlice := make([]Index, 0)
		newSlice = append(newSlice, t.TableIndexSlice...)
		for _, index := range t.TableIndexSlice {
			// 如果是独立的index
			if len(index.FieldsSlice) > 1 {
				slice := index
				slice.Type = MUTI
				slice.FieldsSlice = make([]string, 0)
				length := len(index.FieldsSlice) - 1
				for length > 0 {
					for i := 0; i < length; i++ {
						slice.FieldsSlice = append(slice.FieldsSlice, index.FieldsSlice[i])
					}
					length--
				}

				addFlg := false
				for _, v := range newSlice {
					if len(v.FieldsSlice) == len(slice.FieldsSlice) {
						for _, s := range v.FieldsSlice {
							if !IsExists(s, slice.FieldsSlice) {
								addFlg = true
								break
							}
						}
					} else {
						addFlg = true
					}
				}
				if addFlg {
					newSlice = append(newSlice, slice)
				}
			}
		}
		t.TableIndexSlice = newSlice

		for k, field := range t.Fields {
			for i, index := range t.TableIndexSlice {
				if index.Name == "PRIMARY" && IsExists(field.Field, index.FieldsSlice) {
					t.Fields[k].ColumnPrimaryStr = ";PRIMARY_KEY"
				} else if IsExists(field.Field, index.FieldsSlice) {
					t.Fields[k].ColumnPrimaryStr += ";" + index.Name + "_" + strconv.FormatInt(int64(i), 10)
					if index.Type == UNI {
						t.Fields[k].ColumnPrimaryStr += "_UNIQUE"
					} else {
						t.Fields[k].ColumnPrimaryStr += "_MULTI"
					}
				}
			}
		}

		for k, field := range t.Fields {
			t.Fields[k].GormTagStr = "column:" + field.Field
			t.Fields[k].JsonTagStr = field.Field
			if strings.HasPrefix(field.Default, "CURRENT_TIMESTAMP") {
				t.Fields[k].GormTagStr = "-"
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
	{{.Name}}  {{.Type}} ` + "`" + `gorm:"{{.GormTagStr}}{{.ColumnPrimaryStr}}" ` + `json:"{{.JsonTagStr}}"` + "`" + ` // {{.Comment}} {{end}} 
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

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func IsExists(val interface{}, array interface{}) bool {
	e, _ := InArray(val, array)
	return e
}
