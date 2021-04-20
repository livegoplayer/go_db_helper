package private_model

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

func (mt *MysqlTask) parseFunction() []Func {
	str, err := ioutil.ReadFile(mt.BuildFilePath)
	if err != nil {
		panic(err)
	}
	reg := regexp.MustCompile("func \\(build \\*Build\\)\\s+([A-Z]\\S+)\\((.+)?\\) \\*Build \\{")
	x := reg.FindAllStringSubmatch(string(str), -1)
	funcs := make([]Func, 0)
	for _, s := range x {
		if IsExists(s[1], mt.ignoreMethod) {
			// 手动忽略的方法
			continue
		}
		if s[2] == "" {
			funcs = append(funcs, Func{Name: s[1], Proxy: s[1], Argus: "", ToBuild: ""})
		} else {
			argus := strings.Split(s[2], ",")
			toBuilds := make([]string, len(argus))
			for j, a := range argus {
				arrs := strings.Split(strings.Trim(a, " "), " ")
				if arrs[1][0:3] == "..." {
					toBuilds[j] = arrs[0] + "..."
				} else {
					toBuilds[j] = arrs[0]
				}
			}
			funcs = append(funcs, Func{Name: s[1], Proxy: s[1], Argus: s[2], ToBuild: strings.Join(toBuilds, ",")})
		}
	}

	return funcs
}

func (mt *MysqlTask) parseField(fileTxt string) (DefineFields, []Index) {
	// \s+ 用于去除任何空白字符，其后的 \w+? 用于获取字段名，如果存在使用 // 单行注释字段该字段将会被略过
	reg := regexp.MustCompile("@\\s+(\\w+?)\\s+(\\S+?)\\s+`gorm:\"(\\S+?)\"[^`]+?json:\"(\\S+?)\"`.*?@")
	x := reg.FindAllStringSubmatch(fileTxt, -1)
	names := make(DefineFields, 0)
	indexs := make([]Index, 0)
	indexMap := make(map[string][]string, 0)
	for _, field := range x {
		// field[0] Raw Str
		// field[1] Struct Field Name
		// field[2] Struct Field Type
		// field[3] gorm tag
		// field[4] json tag

		// 获取 column name
		if len(field) > 3 {
			// 存在 column
			if pos := strings.Index(field[3], "column:"); pos >= 0 {
				begin := pos + len("column:")
				// 区分 gorm:"column:theater_id;PRIMARY_KEY"
				if sp := strings.Index(field[3], ";"); sp >= 0 && sp > begin {
					str := strings.Trim(field[3][sp:], ";")
					i := strings.Split(str, ";")
					f := field[3][begin:sp]

					for _, v := range i {
						if indexMap[v] == nil {
							indexMap[v] = make([]string, 0)
						}
						if !IsExists(f, indexMap[v]) {
							indexMap[v] = append(indexMap[v], f)
						}
					}

					field[3] = field[3][begin:sp]
				} else {
					field[3] = field[3][begin:]
				}
			} else {
				continue
			}
		} else {
			continue
		}

		isNum := IsExists(field[2], []string{"int64", "int", "float64", "float32"})
		names = append(names, DefineField{StructKey: field[1], Key: field[3], Type: field[2], Number: isNum})
	}

	indexs = ParseIndex(indexMap, names)

	return names, indexs
}

func NewMysqlTask(targetPath, PackageName string) *MysqlTask {
	return &MysqlTask{
		Task: Task{
			FromDirPath:     filepath.FromSlash(targetPath),
			BuildFilePath:   filepath.Join(BuildFileDir, "build.go"),
			PackageName:     PackageName,
			WriteDirPath:    filepath.FromSlash(targetPath),
			DriverName:      "mysql",
			ignoreMethod:    []string{"clone", "Clone", "TableName"},
			ModelFilterFunc: defaultFilter,
		},
	}
}

type MysqlTask struct {
	Task
}

func (mt *MysqlTask) Run() {
	funcs := mt.parseFunction()
	if mt.IsPrivate {
		// 如果是渲染私有的方式，替换方法名
		for i, item := range funcs {
			if strings.Contains(strings.ToLower(item.Name), "where") {
				funcs[i].Name = MethodNameToPrivate(item.Name)
				continue
			}
		}
	}
	allTypes, fileds := mt.getAllTypeName()

	workTypes := make([]string, 0)
	workFileds := make([]File, 0)
	for i := range allTypes {
		workTypes = append(workTypes, allTypes[i])
		workFileds = append(workFileds, fileds[i])

		if (i > 0 && i%5 == 0) || i == len(allTypes)-1 {
			query := mt.topLine() +
				"import \"reflect\"\n" +
				"import \"" + GetBuildPath() + "\"\n"

			for i, tname := range workTypes {
				query += mt.renderQuery(funcs, tname, workFileds[i])
			}

			Write(query, filepath.Join(mt.WriteDirPath, "lib_auto_generate_query.go"))

			workTypes = make([]string, 0)
			workFileds = make([]File, 0)
		}
	}
}

func (mt *MysqlTask) renderQuery(funcs []Func, typeName string, filed File) string {
	nums := make(DefineFields, 0)
	for _, i := range filed.Files {
		if i.Number {
			nums = append(nums, i)
		}
	}
	t := Render{funcs, typeName, typeName + "Query", mt.DriverName, Fields{
		All:      filed.Files,
		Pluck:    filed.Files,
		PluckUni: filed.Files,
		Map:      filed.Files,
		Number:   nums,
	}}

	for _, v := range filed.IndexMap {
		if v.IsMulti {
			t.Fields.MutiIndex = append(t.Fields.MutiIndex, v)
		}

		if v.IsUnique {
			t.Fields.UniIndex = append(t.Fields.UniIndex, v)
		}
	}

	code := t.Render(mt.ExecTemplate())
	if mt.IsPrivate {
		code = ToPrivate(code)
	}
	code += t.Render(mt.ProxyTemplate())
	if mt.IsPrivate {
		code += t.Render(mt.PrivateBuildQueryTemplate())
	} else {
		code += t.Render(mt.PublicBuildQueryTemplate())
	}

	code += t.Render(mt.BasePublicBuildQueryTemplate())
	return code
}

type File struct {
	Files    DefineFields
	IndexMap []Index
}

func (mt *MysqlTask) getAllTypeName() ([]string, []File) {
	files, _ := ioutil.ReadDir(mt.FromDirPath)
	tables := make([]string, 0)
	fields := make([]File, 0)
	for _, f := range files {

		if f.IsDir() {
			continue
		}
		if !mt.ModelFilterFunc(f.Name()) {
			continue
		}
		str, err := ioutil.ReadFile(filepath.Join(mt.FromDirPath, f.Name()))
		if err != nil {
			panic(err)
		}
		name := f.Name()[:len(f.Name())-3]
		if !regexp.MustCompile("\\n\\s+" + mt.DriverName + ".Query\\s*\\n").MatchString(string(str)) {
			continue
		}
		if !regexp.MustCompile(".*type " + name + "Query struct").MatchString(string(str)) {
			continue
		}

		reg := regexp.MustCompile("\\n")
		modelStr := reg.ReplaceAllString(string(str), "@@")
		// } 后增加判断换行以防止误判 interface{} 类型为 struct 结尾
		targetStructStr := regexp.MustCompile("type " + name + " struct {.+?}@@").FindStringSubmatch(modelStr)
		if len(targetStructStr) == 0 {
			continue
		}

		fs := File{}
		fs.Files, fs.IndexMap = mt.parseField(targetStructStr[0])
		fields = append(fields, fs)
		tables = append(tables, name)
	}

	return tables, fields
}

// 执行后获得结果的方法模板
func (mt *MysqlTask) ExecTemplate() string {
	return `
/**
@Collect
 */
type {{.TypeName}}Collect []{{.TypeName}}

func New{{.QueryName}}() *{{.QueryName}} {
	s := {{.QueryName}}{}
    s.SetBuild({{.Driver}}.NewBuild(&s))
	i, ok := reflect.ValueOf(&s).Interface().(BeforeHook)
	if ok {
		i.Before()
	}
	return &s
}{{ $name := .TypeName }}{{ $queryName := .QueryName }}

type _{{$queryName}}ColsStruct struct{
{{range .Fields.All}}{{.StructKey}} string
{{end}}
}
func Get{{$queryName}}Cols() *_{{$queryName}}ColsStruct {
	return &_{{$queryName}}ColsStruct{
{{range .Fields.All}}{{.StructKey}} : "{{.Key}}",
{{end}}
	}
}

func (m *{{.QueryName}}) First() *{{.TypeName}} {
	s := make([]{{.TypeName}}, 0)
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	m.GetBuild().ModelType(&s).Limit(1).Get()
	if len(s) > 0{
		return &s[0]
	}
	return &{{.TypeName}}{}
}

func (m *{{.QueryName}}) GetOne() *{{.TypeName}} {
	s := make([]{{.TypeName}}, 0)
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	m.GetBuild().ModelType(&s).Limit(1).Get()
	if len(s) > 0{
		return &s[0]
	}
	return nil
}

func (m *{{.QueryName}}) Get() {{.TypeName}}Collect {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := make([]{{.TypeName}}, 0)
	m.GetBuild().ModelType(&s).Get()
	return s
}

func (m *{{.QueryName}}) Clone() *{{.QueryName}} {
	nm := New{{.QueryName}}()
	nm.SetBuild(m.GetBuild().Clone())
	return nm
}

func (m *{{.QueryName}}) Count() int64 {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Count()
}

func (m *{{.QueryName}}) Sum(col string) float64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Sum(col)
}

func (m *{{.QueryName}}) Max(col string) float64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Max(col)
}

func (m *{{.QueryName}}) DoneOperate() int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).DoneOperate()
}

func (m *{{.QueryName}}) Update(h *{{.TypeName}}) int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Update(h)
}

func (m *{{.QueryName}}) Delete() int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Delete()
}

func (m *{{.QueryName}}) Save(x *{{.TypeName}}) {
    m.GetBuild().ModelType(x).Save()
}

func (m *{{.QueryName}}) Error() error {
    return m.GetBuild().ModelType(m).Error()
}

//支持分表
func (m *{{.QueryName}}) Insert(argu interface{}) {
	s := {{.TypeName}}{}
	m.GetBuild().ModelType(&s).Insert(argu)
}

func (m *{{.QueryName}}) Increment(column string, amount int) int64 {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Increment(column, amount)
}
`
}

func IsExists(val interface{}, array interface{}) bool {
	e, _ := InArray(val, array)
	return e
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

func ParseIndex(m map[string][]string, fields DefineFields) []Index {
	list := make([]Index, 0)
	for key, f := range m {
		i := Index{}
		if strings.HasSuffix(key, "PRIMARY_KEY") {
			i.IsUnique = true
		}

		if strings.HasSuffix(key, "_UNIQUE") {
			i.IsUnique = true
		}

		if strings.HasSuffix(key, "_MULTI") {
			i.IsMulti = true
		}

		fieldsMap := fields.KeyByKey()
		for _, name := range f {
			if field, ok := fieldsMap[name]; ok {
				i.Fields = append(i.Fields, field)
			}
		}

		list = append(list, i)
	}

	return list
}
