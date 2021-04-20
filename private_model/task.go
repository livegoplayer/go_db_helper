package private_model

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func Write(text, path string) {
	_ = ioutil.WriteFile(path, []byte(text), 0644)
}

// 替换包名后，将模板文件写入目标文件夹中
func CpTemplate(templateDir, targetPath, packageName string) {
	fs, _ := ioutil.ReadDir(filepath.FromSlash(templateDir))

	for _, file := range fs {
		str, _ := ioutil.ReadFile(filepath.FromSlash(path.Join(templateDir, file.Name())))
		afterStr := strings.Replace(string(str), "package template", "package "+packageName, 1)
		Write(afterStr, path.Join(targetPath, "lib_auto_generate_"+file.Name()))
	}
}

// 将文件中声明的方法，全部变为私有
func ToPrivate(str string) string {
	matchStr := str

	white := []string{"Get", "First", "GetOne", "Count", "DoneOperate", "Delete"}

	reg := regexp.MustCompile(`\) [A-Z][a-zA-Z]+\(`)
	arr := reg.FindAllStringSubmatch(matchStr, -1)
	for _, s := range arr {
		beforeMethod := s[0]
		ignore := false
		for _, w := range white {
			if strings.Contains(beforeMethod, w) {
				ignore = true
			}
		}
		if ignore {
			continue
		}

		// 第三个字符为方法的首字符
		first := beforeMethod[2:3]
		btStr := []byte(beforeMethod)
		low := strings.ToLower(first)
		btStr[2] = []byte(low)[0]

		matchStr = strings.Replace(matchStr, beforeMethod, string(btStr), 1)
	}
	return matchStr
}

type DefineField struct {
	StructKey string
	Key       string
	Type      string
	IsPrimary bool
	IsUnique  bool
	IsMulti   bool
	Number    bool
}

type Fields struct {
	All        []DefineField
	Number     []DefineField
	Pluck      []DefineField
	PluckUni   []DefineField
	UniIndex   []DefineField
	MultiIndex []DefineField
	Map        []DefineField
}

type Func struct {
	Name    string // 自己的名字
	Proxy   string // 执行的build方法名
	Argus   string
	ToBuild string
}

type Render struct {
	Funcs     []Func
	TypeName  string
	QueryName string
	Driver    string
	Fields    Fields
}

// 渲染
func (t Render) Render(tmp string) string {
	var doc bytes.Buffer
	tm, err := template.New("code").Parse(tmp)
	if err != nil {
		panic(err)
	}
	_ = tm.Execute(&doc, t)
	html := doc.String()
	return html
}

func MethodNameToPrivate(str string) string {
	btStr := []byte(str)
	btStr[0] = []byte(strings.ToLower(str[0:1]))[0]
	return string(btStr)
}

type Filter func(name string) bool

func defaultFilter(name string) bool {
	return !strings.Contains(name, "_") && name[0] <= 'Z' && name[0] >= 'A'
}

type Task struct {
	FromDirPath     string // 解析的结构体，所在的文件夹
	BuildFilePath   string
	ignoreMethod    []string // 自动解析出来的方法，需要跳过的内容
	PackageName     string   // 包名
	WriteDirPath    string   // 生成的代码，写入的路劲
	DriverName      string   // 驱动名
	IsPrivate       bool     // 生产的方法，是否是私有
	ModelFilterFunc Filter
}

func (mt Task) topLine() string {
	return "package " + mt.PackageName + "\n"
}

// 代理了build的方法集
func (mt Task) ProxyTemplate() string {
	return `
{{ $name := .TypeName }}
{{ $queryName := .QueryName }}
{{range .Funcs}}
func (m *{{$queryName}}) {{.Name}}({{.Argus}}) *{{$queryName}}{
	m.GetBuild().{{.Proxy}}({{.ToBuild}})
	return m
}
{{end}}

`
}

// 用于构建查询语句的模板, 公有的
func (mt Task) PublicBuildQueryTemplate() string {
	return `


{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}(args ...interface{}) *{{$queryName}}{
	return m.Where("{{.Key}}", args...)
}
{{end}}


{{range .Fields.All}}
func (m *{{$queryName}}) KSet{{.StructKey}}(val interface{}) *{{$queryName}}{
	return m.Set("{{.Key}}", val)
}
{{end}}

{{range .Fields.Number}}
func (m *{{$queryName}}) KInc{{.StructKey}}(num int) *{{$queryName}}{
	return m.Inc("{{.Key}}", num)
}
{{end}}


{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}In(values interface{}) *{{$queryName}}{
	return m.WhereIn("{{.Key}}", values)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}NotIn(values interface{}) *{{$queryName}}{
	return m.WhereNotIn("{{.Key}}", values)
}
{{end}}
`
}

// 用于构建查询语句的模板, 私有的
func (mt Task) PrivateBuildQueryTemplate() string {
	return `


{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}(args ...interface{}) *{{$queryName}}{
	return m.where("{{.Key}}", args...)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) kSet{{.StructKey}}(val interface{}) *{{$queryName}}{
	return m.Set("{{.Key}}", val)
}
{{end}}

{{range .Fields.Number}}
func (m *{{$queryName}}) kInc{{.StructKey}}(num int) *{{$queryName}}{
	return m.Inc("{{.Key}}", num)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}In(values interface{}) *{{$queryName}}{
	return m.whereIn("{{.Key}}", values)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}NotIn(values interface{}) *{{$queryName}}{
	return m.whereNotIn("{{.Key}}", values)
}
{{end}}
`
}

// 一些常用的
func (mt Task) BasePublicBuildQueryTemplate() string {
	return `
{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

func Save(f *{{$name}}) *{{$name}} {
	New{{$queryName}}().save(f)
	return f
}

func Update{{$name}}All(p *{{$name}}) int64 {
	build := New{{$queryName}}()
	return build.update(p)
}

func Fetch{{$name}}All() {{$name}}Collect {
	build := New{{$queryName}}()
	return build.Get()
}
{{range .Fields.UniIndex}}
func Update{{$name}}By{{.StructKey}}s(x []{{.Type}}, p *{{$name}}) int64 {
	build := New{{$queryName}}()

	if len(x) == 0 {
		return 0
	}

	if len(x) == 1 {
		build.kWhe{{.StructKey}}(x[0])
	}else{
		build.kWhe{{.StructKey}}In(x)
	}

	return build.update(p)
}

func Update{{$name}}By{{.StructKey}}sWhatEver (x []{{.Type}}, p *{{$name}}) int64 {
	build := New{{$queryName}}()

	if len(x) == 1 {
		build.kWhe{{.StructKey}}(x[0])
	}else if len(x) > 1{
		build.kWhe{{.StructKey}}In(x)
	}

	return build.update(p)
}

func Update{{$name}}By{{.StructKey}} (x {{.Type}}, p *{{$name}}) int64 {
	build := New{{$queryName}}()
	build.kWhe{{.StructKey}}(x)
	return build.update(p)
}

func CheckExistBy{{.StructKey}} (m {{.Type}}) bool {
	cnt := New{{$queryName}}().kWhe{{.StructKey}}(m).Count()
	return cnt > 0
}

func GetOneBy{{.StructKey}} (m {{.Type}}) *{{$name}} {
	return New{{$queryName}}().kWhe{{.StructKey}}(m).GetOne()
}

func GetFirstBy{{.StructKey}} (m {{.Type}}) *{{$name}} {
	return New{{$queryName}}().kWhe{{.StructKey}}(m).First()
}
{{end}}{{range .Fields.MultiIndex}}
func FetchBy{{.StructKey}} (m {{.Type}}) {{$name}}Collect {
	return New{{$queryName}}().kWhe{{.StructKey}}(m).Get()
}

func GetFirstBy{{.StructKey}} (m {{.Type}}) *{{$name}} {
	return New{{$queryName}}().kWhe{{.StructKey}}(m).First()
}

func GetOneBy{{.StructKey}} (m {{.Type}}) *{{$name}} {
	return New{{$queryName}}().kWhe{{.StructKey}}(m).GetOne()
}

func Update{{$name}}By{{.StructKey}} (x {{.Type}}, p *{{$name}}) int64 {
	build := New{{$queryName}}()
	build.kWhe{{.StructKey}}(x)
	return build.update(p)
}

func Update{{$name}}By{{.StructKey}}s(x []{{.Type}}, p *{{$name}}) int64 {
	build := New{{$queryName}}()

	if len(x) == 0 {
		return 0
	}

	if len(x) == 1 {
		build.kWhe{{.StructKey}}(x[0])
	}else{
		build.kWhe{{.StructKey}}In(x)
	}

	return build.update(p)
}
{{end}}
`
}
