package parse_collect

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type Stru struct {
	PackageName string
	packagePath string
	FilePath    string
	FileDir     string
	StruName    string
	Imports     []*ast.ImportSpec
	ImportStru  StruImports
	StruType    *StruType
	CommentType []Comment
}

type StruType struct {
	NewName string
	Type    *FieldType
}

type PackagePath string
type PackageName string
type TypeName string

type StruImports []Im

type Im struct {
	alias    string
	RealName string
	path     string
	DirPath  string
}

func (i *Im) ToString() string {
	return i.path
}

func (i *Im) IsEmpty() bool {
	return i.path == ""
}

func (i *Im) IsBasePro() bool {
	if i.path == "" {
		panic("请先初始化")
	}

	return IsBasePro(AppRoot, i.path)
}

func (i *Im) parseDirPath() string {
	i.DirPath = ParseDirPathByImport(AppRoot, i.path)
	return i.DirPath
}

func (i *Im) Equal(j *Im) bool {
	return i.path == j.path
}

type FieldType struct {
	PackageName   string       // 当前变量的类型所属包名
	Extends       []FieldType  // 继承的结构体变量
	Im            Im           // 当前变量的所属包信息
	FieldName     string       // 当前变量名
	FieldTypeName string       // 当前变量的类型
	FieldRefType  FieldRefType // 当前变量的类型的基础类型名
	FieldStrut    ast.Ident    // 当前变量的ast对象存储
	TagName       TagNameType  // 当前变量的tag( 如果有的话 )
	ChildList     []FieldType  // 当前变量如果是结构体，他的子变量列表, 也有可能是map类型，存放map的key和value // todo
	InPkg         bool         // 是否在当前包
}

func (ft *FieldType) getTypeName() string {
	typeName := ft.FieldTypeName
	if !ft.InPkg && typeName != "" {
		if ft.PackageName != "" {
			if ft.FieldTypeName[0] == '*' {
				typeName = "*" + ft.PackageName + "." + ft.FieldTypeName[1:]
			} else {
				typeName = ft.PackageName + "." + ft.FieldTypeName
			}
		}
	}

	if ft.FieldRefType == Star {
		typeName = ft.FieldRefType.ToString() + typeName
	}

	if ft.FieldRefType == Map {
		typeName = ft.FieldTypeName
	}

	if ft.FieldRefType == SLICE {
		typeName = "[]" + typeName
	}

	if typeName == "" {
		typeName = ft.FieldRefType.ToString()
	}

	return typeName
}

type TagNameType string

func (t TagNameType) ToString() string {
	reg := regexp.MustCompile(`.*?json:"(.*?)".*`)
	m := reg.FindAllStringSubmatch(string(t), -1)
	if len(m) > 0 {
		if len(m[0]) > 1 {
			return m[0][1]
		}
	}
	return ""
}

type FieldRefType string

func (t FieldRefType) ToString() string {
	return string(t)
}

const SLICE FieldRefType = "array"
const INT FieldRefType = "int"
const INT8 FieldRefType = "int8"
const INT16 FieldRefType = "int16"
const INT32 FieldRefType = "int32"
const INT64 FieldRefType = "int64"
const UINT FieldRefType = "uint"
const UINT8 FieldRefType = "uint8"
const UINT16 FieldRefType = "uint16"
const UINT32 FieldRefType = "uint32"
const UINT64 FieldRefType = "uint64"
const FLOAT32 FieldRefType = "float32"
const FLOAT64 FieldRefType = "float64"
const STRING FieldRefType = "string"
const BOOL FieldRefType = "bool"
const STRUCT FieldRefType = "struct"
const Chan FieldRefType = "chan"
const Interface FieldRefType = "interface{}"
const Map FieldRefType = "map"
const Star FieldRefType = "*"

// 一般是超过两次的类型定义, 且都在别的包 如下
// type a m.b
// type b n.c
const Complex FieldRefType = "complex"

func (f FieldType) ToString() string {
	return f.PackageName + "." + f.FieldName
}

// 解析内部结构, 因为ast对不同文件中的内容并没有提供跨文件变量引用解析，所以这里是一个手动实现的逻辑
func (ft *FieldType) parseInnerStruct() {
	// 当前只支持这两种类型
	if ft.FieldRefType == SLICE || ft.FieldRefType == STRUCT || ft.FieldRefType == "" {
		// 首先找到引用导入
		if ft.Im.IsEmpty() {
			panic("未找到引用导入：" + ft.PackageName + ":" + ft.FieldName)
		}

		// 在预加载的内存中加载import
		imPkg := ft.Im.MatchPkg()

		if imPkg == nil {
			panic("未找到引用导入：" + ft.PackageName + ":" + ft.FieldName)
		}

		// 找到导入的变量, 直接赋值
		if childStru, ok := imPkg[TypeName(ft.FieldTypeName)]; ok {
			// 匹配extends
			ft.ChildList = childStru.StruType.Type.ChildList
			for k, v := range childStru.StruType.Type.Extends {
				v.parseInnerStruct()
				x := v
				ft.MergeChild(v.ChildList)
				childStru.StruType.Type.Extends[k] = x
			}
			ft.Extends = childStru.StruType.Type.Extends
			// 补全extends的缺失变量
			ft.setFieldRefType(childStru.StruType.Type.FieldRefType)
		}

		// 补全因为调用其他包而缺失的reftype
		for k, v := range ft.ChildList {
			t := v
			if v.FieldRefType != "" {
				continue
			}
			// 在预加载的内存中加载import
			z := v.Im.MatchPkg()
			// 找到导入的变量, 直接赋值
			if ct, ok := z[TypeName(v.FieldTypeName)]; ok {
				// 如果还是调用的别的包的变量，直接放弃
				if ct.StruType.Type.FieldRefType == "" {
					t.FieldRefType = Complex
				}
				t.FieldRefType = ct.StruType.Type.FieldRefType
			}
			ft.ChildList[k] = t
		}

	}

	return
}

// 解析packagePath
func (s *Stru) parsePackagePath(appRoot string) {
	// 解析出import的包名
	if s.FileDir == "" {
		panic("请先解析出FileDir再调用本函数")
	}

	// 这里是本项目下的包
	s.packagePath = ParseImportByDirPath(appRoot, s.FileDir)
}

// 解析imports
func (s *Stru) parseImports() {
	list := make(StruImports, 0)

	// 解析所有的imports的type定义 并且存入内存
	for _, v := range s.Imports {
		item := Im{}

		if v.Name != nil {
			item.alias = v.Name.Name
		}

		if v.Path != nil {

			item.path = strings.Trim(v.Path.Value, "\"")

			// 忽略C
			if item.path == "C" {
				continue
			}
			item.parseDirPath()
			strus, err := item.parsePathTypes()
			if len(strus) > 0 {
				item.RealName = strus[0].PackageName
			}
			// 存入imports ,避免重复获取
			if err == nil {
				if len(strus) > 0 {
					setPathStrus(strus)
				}
			}
		}

		list = append(list, item)
	}

	// 本包的所有type定义也存入内存
	item := s.getSelfIm()
	strus, err := item.parsePathTypes()
	// 存入imports ,避免重复获取
	if err == nil {
		if len(strus) > 0 {
			setPathStrus(strus)
		}
	}

	list = append(list, item)

	// 把解析完成的结果存入结构体，给下面匹配使用
	s.ImportStru = list

	// 匹配所有选中的type定义以及其子类型的导入Im，这里简单通过ImportStru匹配即可
	s.matchTypePkgName()
}

// 获取自己所在包的Im
func (s *Stru) getSelfIm() Im {
	item := Im{
		alias:    "",
		path:     s.packagePath,
		RealName: s.PackageName,
	}

	item.parseDirPath()
	// 存入imports ,避免重复获取
	return item
}

// 解析结构体变量类型的导入
func (s *Stru) matchTypePkgName() {
	// 如果本类型的切片类型在本包，直接赋值
	if s.StruType.Type.InPkg {
		selfIm := s.getSelfIm()
		s.StruType.Type.Im = selfIm
	}

	for _, v := range s.ImportStru {
		// 如果不在本包，进行匹配
		t := s.StruType.Type
		if !s.StruType.Type.InPkg {
			if v.match(t.PackageName) {
				t.Im = v
			}
		}

		// 对结构体子类型做一遍，如果没有，就不做，子类型没有在不在本包的区分，
		// 因为s.ImportStru里面包含本包内容，这里的本包内容是直接匹配的
		// 注意点是val.PackageName 一定要存在
		for k, val := range t.ChildList {
			x := val
			if v.match(x.PackageName) {
				x.Im = v
			}
			t.ChildList[k] = x
		}

		// 对结构体继承的变量再做一遍匹配, 把父类型的childlist 加入当前并且根据tag去重,
		for k, val := range t.Extends {
			x := val
			// 对结构体子类型做一遍，如果没有，就不做，子类型没有在不在本包的区分，
			// 因为s.ImportStru里面包含本包内容，这里的本包内容是直接匹配的
			// 注意点是val.PackageName 一定要存在
			for kx, vx := range val.ChildList {
				x := vx
				if v.match(x.PackageName) {
					x.Im = v
				}
				val.ChildList[kx] = x
			}

			if v.match(x.PackageName) {
				x.Im = v
				t.Extends[k] = x
			}
		}
	}
}

func (ft *FieldType) MergeChild(mergeTypes []FieldType) {
	for _, unit := range mergeTypes {
		for _, child := range ft.ChildList {
			if unit.TagName != "" {
				if unit.TagName.ToString() == child.TagName.ToString() {
					continue
				}
			}
		}
		unit.InPkg = false
		ft.ChildList = append(ft.ChildList, unit)
	}
}

func (i Im) match(pkgName string) bool {
	if i.alias == "" {
		if i.RealName == pkgName {
			return true
		}
	} else {
		if i.alias == pkgName {
			return true
		}
	}

	return false
}

var AppRoot = ""
var GORoot = ""

// 索引缓存
var Imports ImportCache

type ImportCacheTypeMap map[TypeName]Stru
type ImportCachePackageMap map[PackageName]ImportCacheTypeMap
type ImportCachePathMap map[PackagePath]ImportCachePackageMap
type ImportCache ImportCachePathMap

func MatchType(pkgPath, pkgName, name string) *Stru {
	if Imports != nil {
		if _, ok := Imports[PackagePath(pkgPath)]; ok {
			if _, ok := Imports[PackagePath(pkgPath)][PackageName(pkgName)]; ok {
				if _, ok := Imports[PackagePath(pkgPath)][PackageName(pkgName)][TypeName(name)]; ok {
					t := Imports[PackagePath(pkgPath)][PackageName(pkgName)][TypeName(name)]
					return &t
				}
			}
		}
	}

	return nil
}

func MatchPkg(pkgPath, pkgName string) ImportCacheTypeMap {
	if Imports != nil {
		if _, ok := Imports[PackagePath(pkgPath)]; ok {
			if _, ok := Imports[PackagePath(pkgPath)][PackageName(pkgName)]; ok {
				return Imports[PackagePath(pkgPath)][PackageName(pkgName)]
			}
		}
	}

	return nil
}

func MatchPath(pkgPath string) ImportCachePackageMap {
	if Imports != nil {
		if _, ok := Imports[PackagePath(pkgPath)]; ok {
			return Imports[PackagePath(pkgPath)]
		}
	}

	return nil
}

func (im *Im) MatchPkg() ImportCacheTypeMap {
	pkgName := im.RealName
	return MatchPkg(im.path, pkgName)
}

func (im *Im) MatchPath() ImportCachePackageMap {
	return MatchPath(im.path)
}

func Parse(AppROOT string) {
	if AppRoot == "" {
		AppRoot = AppROOT
	}

	Imports = make(ImportCache, 0)
	DirPath := AppRoot
	// 获取了所有的collectstru对象列表
	// csl := getAllCollimpectsByPath(DirPath)
	dps := []string{DirPath}

	// 清空以前生成的文件
	cleanDirs(dps)
	strus := getAllCollectsByPathAst(dps)

	for _, v := range strus {
		print(v.FilePath + ":" + v.StruName + "\n")
	}

	t := NewMainTask(strus)
	cts := t.GetChildTasks()

	for _, ct := range cts {
		ct.Run()
	}
}

func (ft *FieldType) parseCollect() {
	// 切片和struck需要丰富
	if ft.FieldRefType == SLICE || ft.FieldRefType == STRUCT {
		ft.parseInnerStruct()
	}
}

func (s *Stru) parseCollect() {
	s.StruType.Type.parseCollect()
}

// 解析所有的含有以下注解的类型
/**
 *  @Collect()
 */
// type myCollect []int
func getAllCollectsByPathAst(dirPaths []string) []Stru {
	var struList []Stru
	if len(dirPaths) == 0 {
		return []Stru{}
	}

	for _, dirPath := range dirPaths {
		strus, err := parseDir(dirPath)
		if err != nil {
			panic(err)
		}

		for _, stru := range strus {
			if IsExists(stru, struList) {
				continue
			}
			stru.parseCollect()
			struList = append(struList, stru)
		}
	}

	return struList
}

// 删除文件夹下的所有以前生成的文件
func cleanDirs(dirPaths []string) {
	for _, dirPath := range dirPaths {
		err := cleanDir(dirPath)
		if err != nil {
			panic(err)
		}
	}
}

// 清除文件夹下所有以前生成的文件
func cleanDir(dirPath string) (err error) {
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if err != nil {
				return err
			}

			if strings.HasPrefix(info.Name(), "lib_auto_generate_") && strings.HasSuffix(info.Name(), "collect.go") {
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}

		}
		return nil
	})
	return err
}

func parseDir(dirPath string) (strULS []Stru, err error) {
	// 获取当前目录下的所有collect
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if err != nil {
				return err
			}

			strS := parseFile(path)

			for _, v := range strS {
				strULS = append(strULS, v)
			}
		}
		return nil
	})

	return strULS, err
}

func parseFile(path string) (strULS []Stru) {
	pkgs := getCommentsNodeInFile(path)

	for pkgName, p := range pkgs {
		for filePath, f := range p.Files {
			for _, decl := range f.Decls {
				// 这里解析变量
				if fdecl, ok := decl.(*ast.GenDecl); ok {
					if fdecl.Doc != nil {
						items, err := parseReturnCollectionType(fdecl.Doc)
						// 如果有解析出特定注释
						if len(items) > 0 {
							if len(fdecl.Specs) > 0 {
								// 如果是变量
								if ts, ok := fdecl.Specs[0].(*ast.TypeSpec); ok {
									ft := FieldType{
										PackageName: pkgName,
									}
									ident := ts.Name

									item := Stru{
										PackageName: pkgName,
										FilePath:    filePath,
										Imports:     f.Imports,
										FileDir:     path,
										StruName:    ts.Name.Name,
										StruType: &StruType{
											NewName: ts.Name.Name,
											Type:    &ft,
										},
										CommentType: items,
									}
									item.parsePackagePath(AppRoot)
									ft.parseTopIdent(*ident)
									item.StruType.Type = &ft
									item.parseImports()
									strULS = append(strULS, item)
								}
							}
						}

						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
	}

	return
}

func parseFileTypes(path string) (strULS []Stru) {
	pkgs := getTypeNodeInFile(path)

	for pkgName, p := range pkgs {
		for filePath, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				ret, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}
				ft := FieldType{
					PackageName: pkgName,
				}
				ident := ret.Name

				item := Stru{
					PackageName: pkgName,
					FilePath:    filePath,
					Imports:     f.Imports,
					FileDir:     path,
					StruName:    ret.Name.Name,
					StruType: &StruType{
						NewName: ret.Name.Name,
						Type:    &ft,
					},
				}
				item.parsePackagePath(AppRoot)
				ft.parseTopIdent(*ident)

				item.StruType.Type = &ft
				item.parseImports()
				strULS = append(strULS, item)

				return true
			})
		}
	}

	return
}

func setPathStrus(strus []Stru) {
	for _, v := range strus {
		p := PackagePath(v.packagePath)
		importPackageName := PackageName(v.PackageName)
		typeName := TypeName(v.StruName)
		if _, ok := Imports[p]; ok {
			if _, ok := Imports[p][importPackageName]; ok {
				if _, ok := Imports[p][importPackageName][typeName]; ok {
					continue
				}
			}
		}

		if _, ok := Imports[p]; !ok {
			Imports[p] = make(ImportCachePackageMap, 0)
		}

		if _, ok := Imports[p][importPackageName]; !ok {
			Imports[p][importPackageName] = make(map[TypeName]Stru, 0)
		}

		Imports[p][importPackageName][typeName] = v
	}
}

func setPathLock(path string) {
	p := PackagePath(path)
	if !getPathLock(path) {
		Imports[p] = make(ImportCachePackageMap, 0)
	}
}

func getPathLock(path string) bool {
	p := PackagePath(path)
	if _, ok := Imports[p]; ok {
		return true
	}

	return false
}

// 解析出一个文件目录下的所有types
func (im *Im) parsePathTypes() (strULS []Stru, err error) {
	// 一次解析一个目录
	if getPathLock(im.path) == true {
		// 获取一下已经被解析的stru
		struMap := im.MatchPath()
		for _, v := range struMap {
			for _, v := range v {
				strULS = append(strULS, v)
			}
		}
		return strULS, errors.New("该目录已经被解析过或者正在被解析")
	}
	// fmt.Print("开始解析 " + im.DirPath + " \n")
	setPathLock(im.path)

	strS := parseFileTypes(im.DirPath)
	for _, v := range strS {
		strULS = append(strULS, v)
	}

	return strULS, err
}

// 注释类型
type CommentType string

const ReturnCollectComment = "collect" // 返回集合

type Comment struct {
	// 参数
	Params map[string]interface{}
	Type   CommentType
}

// 解析注释 支持以下三种写法,
/**
  @Collect //这里也支持打注释
*/
/**
  @Collect() //括号里支持传递参数，这里也支持打注释，括号里不要出现 "(" ")" "\n" 这些字符
*/
/**
@Collect(dasdasd)  //括号里支持传递参数，这里也支持打注释，括号里不要出现 "(" ")" "\n" 这些字符
*/
func parseReturnCollectionType(f *ast.CommentGroup) (items []Comment, err error) {
	for _, d := range f.List {
		params := make(map[string]interface{})
		reg := regexp.MustCompile(`.*?@Collect`)
		text := reg.FindAllStringSubmatch(d.Text, -1)
		// 如果成功匹配
		if len(text) > 0 {
			reg := regexp.MustCompile(`.*?@Collect\(([^)(]*)\)`)
			t := reg.FindAllStringSubmatch(d.Text, -1)
			if len(t) > 0 {
				text = t
				// todo text[0][1] 这个字符串支持解析
				// params =
				//fmt.Printf("%s\n", text[0][1])
			} else {
				//fmt.Printf("%s\n", text[0])
			}
		}

		if len(text) > 0 {
			items = append(items, Comment{
				Type:   ReturnCollectComment,
				Params: params,
			})
		}
	}

	return
}

func getCommentsNodeInFile(path string) map[string]*ast.Package {
	pkgs := make(map[string]*ast.Package, 0)
	fset := token.NewFileSet()
	pkgFolder, err := parser.ParseDir(fset, path, validFile, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for k, p := range pkgFolder {
		pkgs[k] = p
	}
	return pkgs
}

func getTypeNodeInFile(path string) map[string]*ast.Package {
	pkgs := make(map[string]*ast.Package, 0)
	fset := token.NewFileSet()
	pkgFolder, err := parser.ParseDir(fset, path, validFile, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for k, p := range pkgFolder {
		pkgs[k] = p
	}

	return pkgs
}

func validFile(info os.FileInfo) bool {
	return !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
}

func getFieldTypeListByAstFields(fields *ast.FieldList, pkgName string) ([]FieldType, []FieldType) {
	fieldList := make([]FieldType, 0)
	preExtends := make([]FieldType, 0)
	for _, field := range fields.List {
		ft := FieldType{
			PackageName: pkgName,
		}

		// 代表这里是一个继承声明
		if len(field.Names) == 0 {
			// 调用本包的
			if ident, ok := field.Type.(*ast.Ident); ok {
				extend := getPreInsertExtend(ident, pkgName)
				if extend != nil {
					preExtends = append(preExtends, *extend)
				}
			}

			// 调用别的包的
			if se, ok := field.Type.(*ast.SelectorExpr); ok {
				extend := getPreInsertExtendOther(se, pkgName)
				if extend != nil {
					preExtends = append(preExtends, *extend)
				}
			}
		} else {
			ft.parseTopIdent(*field.Names[0])
			fieldList = append(fieldList, ft)
		}

	}

	return fieldList, preExtends
}

func (ft *FieldType) parseTopIdent(ident ast.Ident) {
	// 第一步获取有数据的obj
	if ident.Obj == nil {
		return
	}

	// 第二步获取有数据的Decl
	if ident.Obj.Decl == nil {
		return
	}

	// 拿到能拿到的部分，剩下的需要解析
	ft.FieldName = ident.Name
	ft.FieldStrut = ident
	ft.parseDecl(ident.Obj.Decl)

	return
}

func (ft *FieldType) parseStructType(d ast.Expr) {
	if ty, ok := d.(*ast.StructType); ok {
		fieldList, preExtends := getFieldTypeListByAstFields(ty.Fields, ft.PackageName)
		if ft.PackageName == "" {
			return
		}
		ft.ChildList = fieldList
		ft.Extends = preExtends
	}
}

func (ft *FieldType) parseIdent(d ast.Expr) {
	if t, ok := d.(*ast.Ident); ok {
		ft.FieldTypeName = t.Name
		// 如果是基础类型
		if ft.HasBaseTypeName() {
			ft.PackageName = ""
			// 如果不是array的子类型
			if ft.FieldRefType != SLICE {
				ft.FieldRefType = FieldRefType(t.Name)
			}
			ft.InPkg = false
		} else {
			ft.setFieldRefType(STRUCT)
		}
		ft.InPkg = true
	}
}

func (ft *FieldType) parseChanType(d ast.Expr) {
	if ty, ok := d.(*ast.ChanType); ok {
		if t, ok := ty.Value.(*ast.Ident); ok {
			ft.FieldTypeName = "chan" + " " + t.Name
			ft.PackageName = ""
		} else {
			temp := FieldType{}
			temp.parseDeclType(ty.Value)
			ft.FieldTypeName = "chan" + " " + temp.getTypeName()
		}
	}
}

func (ft *FieldType) parseInterfaceType(d ast.Expr) {
	if _, ok := d.(*ast.InterfaceType); ok {
		ft.FieldTypeName = "interface{}"
		ft.PackageName = ""
	}
}

// parsemaptype 目前只支持 value的导入，不支持key 如果需要支持需要把im变量变成数组
func (ft *FieldType) parseMapType(d ast.Expr) {
	if t, ok := d.(*ast.MapType); ok {
		key := ""
		value := ""
		if k, ok := t.Key.(*ast.Ident); ok {
			key = k.Name
		} else {
			temp := FieldType{}
			temp.parseDeclType(t.Key)
			key = temp.getTypeName()
		}
		if v, ok := t.Value.(*ast.Ident); ok {
			value = v.Name
		} else {
			temp := FieldType{}
			temp.parseDeclType(t.Value)
			value = temp.getTypeName()
			ft.PackageName = temp.PackageName
		}
		ft.FieldTypeName = "map[" + key + "]" + value
	}
}

func (ft *FieldType) parseDecl(decl interface{}) {
	switch decl.(type) {
	case *ast.Field, *ast.TypeSpec:
		// 如果是结构体的成员变量，需要额外解析一个tagName
		if decl, ok := decl.(*ast.Field); ok {
			if decl.Tag != nil {
				ft.TagName = TagNameType(decl.Tag.Value)
			}
			ft.parseDeclType(decl.Type)
		}
		// 如果是简单的类型定义
		if decl, ok := decl.(*ast.TypeSpec); ok {
			ft.parseDeclType(decl.Type)
		}
	case *ast.ValueSpec, *ast.ImportSpec:
		// do nothing
		fmt.Print("未记录类型")
		return
	}
}

func (ft *FieldType) parseDeclType(t ast.Expr) {
	// 判断类型
	switch t.(type) {
	// 如果是数组，可以继续往下解析
	case *ast.ArrayType:
		if t, ok := t.(*ast.ArrayType); ok {
			ft.setFieldRefType(SLICE)
			ft.parseDeclTypeArrayType(t)
		}
		return
	// 代表它是一个基础的类型 或者本包的类型
	case *ast.Ident:
		ft.parseIdent(t)

		return
	// 代表它是一个结构体
	case *ast.StructType:
		ft.setFieldRefType(STRUCT)
		ft.parseStructType(t)
		return

	// 一个管道
	case *ast.ChanType:
		ft.setFieldRefType(Chan)
		ft.parseChanType(t)
		return

	// 一个接口类型
	case *ast.InterfaceType:
		ft.setFieldRefType(Interface)
		ft.parseInterfaceType(t)
		return

	// 一个map
	case *ast.MapType:
		ft.setFieldRefType(Map)
		ft.parseMapType(t)
		return

	// 一个指针
	case *ast.StarExpr:
		ft.setFieldRefType(Star)
		ft.parseStarType(t)
		return

	// 这里一般是调用别的包的类型
	case *ast.SelectorExpr:
		ft.parseSelectorExpr(t)
		return

	// 这里可以扩展, 目前忽略fuc，因为func字符串无法通过ast.Expr 拼接，需要上层联动
	case *ast.FuncType:
		// do nothing
		return
	}

}

func (ft *FieldType) setFieldRefType(frt FieldRefType) {
	if ft.FieldRefType == "" {
		ft.FieldRefType = frt
	}
}

// 切片的原子类型
func (ft *FieldType) parseDeclTypeArrayType(t *ast.ArrayType) {
	// 解析不同的type类型
	// 这个对象保存数组原子元素的类型
	switch t.Elt.(type) {
	// 这里一般是调用别的包的类型
	case *ast.SelectorExpr:
		ft.parseSelectorExpr(t.Elt)
		return

	// 这里一般是调用本包的类型
	case *ast.Ident:
		ft.parseIdent(t.Elt)

	// 指针
	case *ast.StarExpr:
		ft.parseStarType(t.Elt)

	case *ast.BasicLit, *ast.FuncLit, *ast.CompositeLit,
		*ast.ParenExpr, *ast.IndexExpr, *ast.SliceExpr, *ast.TypeAssertExpr,
		*ast.CallExpr, *ast.UnaryExpr, *ast.BinaryExpr, *ast.KeyValueExpr:
		// do nothing
		fmt.Print("未记录类型")
		return
	}

	return

}

func (ft *FieldType) parseStarType(star ast.Expr) {
	if se, ok := star.(*ast.StarExpr); ok {
		// 如果在本包
		if pn, ok := se.X.(*ast.Ident); ok {
			ft.FieldTypeName = pn.Name
			if ft.HasBaseTypeName() {
				ft.PackageName = ""
				ft.InPkg = false
			}
			ft.InPkg = true
		}

		// 如果在外包
		if _, ok := se.X.(*ast.SelectorExpr); ok {
			ft.parseSelectorExpr(se.X)
		}

		if ft.FieldRefType != Star {
			ft.FieldTypeName = Star.ToString() + ft.FieldTypeName
		}
	}
}

func (ft *FieldType) parseSelectorExpr(t ast.Expr) {
	if se, ok := t.(*ast.SelectorExpr); ok {
		if pn, ok := se.X.(*ast.Ident); ok {
			ft.PackageName = pn.Name
			ft.FieldTypeName = se.Sel.Name
			ft.InPkg = false
		}
	}
}

// 预先置入继承内容
func getPreInsertExtend(ident *ast.Ident, pkgName string) *FieldType {
	// 第一步获取有数据的obj
	if ident.Obj == nil {
		return nil
	}

	// 第二步获取有数据的Decl
	if ident.Obj.Decl == nil {
		return nil
	}

	// 拿到能拿到的部分，剩下的需要解析
	eft := &FieldType{
		PackageName: pkgName,
	}
	eft.FieldName = ident.Name
	eft.FieldStrut = *ident
	eft.parseDecl(ident.Obj.Decl)

	return eft
}

// 预先置入继承内容 ， 别的包内容
func getPreInsertExtendOther(se *ast.SelectorExpr, pkgName string) *FieldType {

	// 拿到能拿到的部分，剩下的需要解析
	eft := &FieldType{
		PackageName: pkgName,
	}
	eft.parseSelectorExpr(se)

	return eft
}

/*** *****************************************************************/

func (c *ChildTask) Run() {
	top := "package " + c.PackageName + "\n"
	if !IsExists("sort", c.PkgPaths) {
		c.PkgPaths = append(c.PkgPaths, "sort")
	}
	if len(c.PkgPaths) > 0 {
		top += "import (\n"
	}
	for _, p := range c.PkgPaths {
		top += "    \"" + p + "\"" + "\n"
	}
	if len(c.PkgPaths) > 0 {
		top += ")\n"
	}
	top += `
/*
此文件为自动生成，所有修改都不会生效
*/
`
	collect := top
	collect += renderCollect(c.TypeName, c.SubName, c.Fields)
	Write(collect, filepath.Join(c.WriteDirPath, fmt.Sprintf("lib_auto_generate_%s_collect.go", SnakeString(c.TypeName))))
}

type TASK struct {
	Strus []Stru
}

type ChildTask struct {
	Fields       Fields
	FromDirPath  string   // 解析的结构体，所在的文件夹
	PackageName  string   // 包名
	WriteDirPath string   // 生成的代码，写入的路劲
	TypeName     string   // 类型名字
	SubName      string   // 子类型名字
	PkgPaths     []string // 使用到的包名
}

func NewMainTask(Strus []Stru) TASK {
	return TASK{
		Strus: Strus,
	}
}

func (mt TASK) GetChildTasks() []ChildTask {
	list := mt.Strus
	tasks := make([]ChildTask, 0)
	for _, item := range list {
		pkgPaths := make([]string, 0)

		ft := item.StruType.Type
		// 目前只支持切片
		if ft.FieldRefType != SLICE {
			continue
		}

		// 子类型必须是有子元素
		if len(ft.ChildList) == 0 {
			panic("无效类型")
		}

		// 如果在本包，就不需要导入
		subName := ft.FieldTypeName
		if !ft.InPkg {
			subName = ft.PackageName + "." + subName
			if !IsExists(ft.Im.path, pkgPaths) {
				pkgPaths = append(pkgPaths, ft.Im.path)
			}
		}

		// 遍历成员变量
		filed := make([]DefineField, 0)
		pluckFiled := make([]DefineField, 0)
		pluckUniFiled := make([]DefineField, 0)
		mapFiled := make([]DefineField, 0)
		for _, v := range ft.ChildList {
			// 直接忽略小写
			if v.FieldName == "" || ('a' <= v.FieldName[0] && v.FieldName[0] <= 'z') {
				continue
			}

			// 需要导入的包
			// 所有子类型的插入, 除非基础类型 或者本包类型
			if !v.InPkg && !IsExists(v.Im.path, pkgPaths) && !v.HasBaseTypeName() {
				pkgPaths = append(pkgPaths, v.Im.path)
			}

			// 原子解析类型
			defineField := DefineField{v.FieldName, v.TagName.ToString(), v.getTypeName()}
			if v.IsPluckType() {
				pluckFiled = append(pluckFiled, defineField)
			}

			if v.IsPluckUniType() {
				pluckUniFiled = append(pluckUniFiled, defineField)
			}

			if v.IsGroupByType() {
				mapFiled = append(mapFiled, defineField)
			}

			filed = append(filed, defineField)

		}
		if len(filed) > 0 {
			tasks = append(tasks, ChildTask{
				Fields: Fields{
					All:      filed,
					Pluck:    pluckFiled,
					PluckUni: pluckUniFiled,
					Map:      mapFiled,
				},
				PackageName:  item.PackageName,
				FromDirPath:  filepath.Dir(item.FilePath),
				WriteDirPath: filepath.Dir(item.FilePath),
				TypeName:     item.StruName,
				SubName:      subName,
				PkgPaths:     pkgPaths,
			})
		}
	}

	return tasks
}
func (f FieldType) IsPluckUniType() bool {
	return f.IsBaseType()
}

func (f FieldType) IsGroupByType() bool {
	return f.IsBaseType()
}

func (f FieldType) IsBaseType() bool {
	if IsExists(f.FieldRefType, getBaseTypeList()) {
		return true
	}

	// 加入特殊的MUID
	if f.FieldTypeName == "MUID" {
		return true
	}

	return false
}

func (f FieldType) HasBaseTypeName() bool {
	if IsExists(FieldRefType(f.FieldTypeName), getBaseTypeList()) {
		return true
	}

	return false
}

func getBaseTypeList() []FieldRefType {
	return []FieldRefType{
		INT,
		INT8,
		INT16,
		INT32,
		INT64,
		UINT,
		UINT8,
		UINT16,
		UINT32,
		UINT64,
		FLOAT32,
		FLOAT64,
		STRING,
		BOOL,
		Interface,
	}
}

// 如果是基础类型
func (f FieldType) IsPluckType() bool {
	return true
}

func Write(text, path string) {
	ioutil.WriteFile(path, []byte(text), 0644)
}

type Render struct {
	TypeName string
	SubName  string
	Fields   Fields
}

type Fields struct {
	All      []DefineField
	Pluck    []DefineField
	PluckUni []DefineField
	Map      []DefineField
}

type DefineField struct {
	StructKey string
	Key       string
	Type      string
}

// 渲染
func (t Render) Render(tmp string) string {
	var doc bytes.Buffer
	tm, err := template.New("code").Parse(tmp)
	if err != nil {
		panic(err)
	}
	tm.Execute(&doc, t)
	html := doc.String()
	return html
}

func renderCollect(typeName string, subName string, filed Fields) string {
	t := Render{typeName, subName, filed}
	return t.Render(collectTemplate())
}

func collectTemplate() string {
	return `
{{ $name := .TypeName }}{{ $subName := .SubName }}
{{range .Fields.Pluck}}
func(s  {{$name}}) Pluck{{.StructKey}}() []{{.Type}}{
	list := make([]{{.Type}}, len(s))
	for i, v := range s{
		list[i] = v.{{.StructKey}}
	}
	return list
}
{{end}}

{{range .Fields.PluckUni}}
func(s  {{$name}}) PluckUni{{.StructKey}}() []{{.Type}}{
	uniMap := make(map[{{.Type}}]bool)
	list := make([]{{.Type}} ,0)
	for _, v := range s{
		_, ok := uniMap[v.{{.StructKey}}]
		if !ok {
			uniMap[v.{{.StructKey}}] = true
		    list = append(list, v.{{.StructKey}})
		}
	}
	return list
}
{{end}}

{{range .Fields.Map}}
func(s  {{$name}}) GroupBy{{.StructKey}}() map[{{.Type}}]{{$name}}{
	m := make(map[{{.Type}}]{{$name}})
	for _, v := range s{
		if _, ok := m[v.{{.StructKey}}]; !ok{
			m[v.{{.StructKey}}] = make({{$name}}, 0)
		}
		m[v.{{.StructKey}}] = append(m[v.{{.StructKey}}], v)
	}
	return m
}
{{end}}

func (s {{$name}}) SortByFunc (f func(i, j int) bool) {{$name}}{
	sort.SliceStable(s, f)
	return s
}

func(s  {{$name}}) Filter( f func(item {{$subName}}) bool) {{$name}}{
	m := make({{$name}}, 0)
	for _, v := range s{
		if f(v){
			m = append(m, v)
		}
	}
	return m
}

{{range .Fields.Map}}
func(s  {{$name}}) KeyBy{{.StructKey}}() map[{{.Type}}]{{$subName}}{
	m := make(map[{{.Type}}]{{$subName}})
	for _, v := range s{
		m[v.{{.StructKey}}] = v
	}
	return m
}
{{end}}

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

// 驼峰转蛇形
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
