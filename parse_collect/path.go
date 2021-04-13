package parse_collect

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var APPRootModule = ""
var GoRoot = ""
var GoPath = ""
var GOMODCACHE = ""

// 根据包的导入名称获取包的实际位置
func ParseDirPathByImport(APPRoot string, importP string) string {
	if importP == "" {
		return ""
	}

	if IsBasePro(APPRoot, importP) {
		rootPath, _ := filepath.Split(APPRoot)
		returnPath := filepath.Join(rootPath, getRelativeDirNameByRealPackageName(APPRoot, importP))
		return returnPath
	}

	// 如果不是叨叨自定义的包
	realPath := pathStrParse(importP)
	returnPath := filepath.Join(GetGoROOT(), "src", filepath.FromSlash(realPath))
	if !Exists(returnPath) {
		returnPath = filepath.Join(GetGoROOT(), "src", "vendor", filepath.FromSlash(realPath))
		// 检查一下是否存在, 不存在则是第三方库文件，第三方库文件的话，逐个匹配
		if !Exists(returnPath) {
			returnPath = getCacheDir(realPath)
		}
	}

	return returnPath
}

func ParseImportByDirPath(appROOT, fileDir string) string {
	// 解析出import的包名
	packagePath := ""
	if fileDir == "" {
		return ""
	}

	// 如果filedir 匹配 GOMODCACHE 成功
	if strings.HasPrefix(fileDir, GetGOMODCACHE()) {
		packagePath = strings.Trim(filepath.ToSlash(strings.TrimPrefix(fileDir, filepath.Join(GetGOMODCACHE(), "/"))), "/")
		// 去除所有版本号
		temp := strings.Split(packagePath, "/")
		for k, v := range temp {
			reg := regexp.MustCompile(`(.*)?@.*`)
			t := reg.FindAllStringSubmatch(v, -1)
			if len(t) == 0 {
				continue
			}

			if len(t[0]) > 1 {
				temp[k] = t[0][1]
			}
		}

		packagePath = strings.Trim(strings.Join(temp, "/"), "/")
		packagePath = pathStrParseBack(packagePath)
		return packagePath
	}

	// 如果filedir 前缀匹配GoRoot成功, 有可能存在GOMODCACHE 以GOROOT 开头的情况
	ok := strings.HasPrefix(fileDir, GetGoROOT())
	if ok {
		packagePath = strings.TrimPrefix(fileDir, filepath.Join(GetGoROOT(), "src", "vendor"))
		if packagePath == fileDir {
			packagePath = strings.TrimPrefix(fileDir, filepath.Join(GetGoROOT(), "src"))
		}
		return strings.Trim(filepath.ToSlash(packagePath), "/")
	}

	// 这里是本项目下的包
	packagePath = getRealPackageNameByRelativeDirName(appROOT, fileDir)
	return filepath.ToSlash(packagePath)
}

func getRelativeDirNameByRealPackageName(APPRoot, RealPackageName string) string {
	_, appRootDirName := path.Split(APPRoot)
	temp := strings.Split(RealPackageName, "/")

	pathName := temp[0]
	// 如果用的是github试命名法
	if pathName == "github.com" {
		pathName = temp[2]
		temp = temp[2:]
	}
	if pathName == getAPPRootModule(APPRoot) {
		pathName = appRootDirName
	}

	return path.Join(temp...)
}

func getRealPackageNameByRelativeDirName(APPRoot, RelativeDirName string) string {
	pathSuffix := ""
	if strings.HasPrefix(RelativeDirName, APPRoot) {
		pathSuffix = filepath.ToSlash(strings.TrimPrefix(RelativeDirName, APPRoot))
	} else {
		panic(RelativeDirName + "该目录不是项目目录")
	}

	packageName := filepath.Join(getAPPRootModule(APPRoot), pathSuffix)
	return packageName
}

// 获取当前目录下的项目位置，并且匹配项目目录的位置，查看是不是当前项目的包
func IsBasePro(APPRoot string, modulePath string) bool {

	// 解析根目录下的go.mod
	APPRootModule = getAPPRootModule(APPRoot)

	if modulePath == "" {
		panic("请先初始化")
	}

	tempPath := strings.Split(modulePath, "/")
	_, rootName := filepath.Split(APPRootModule)
	pathName := tempPath[0]
	// 如果用的是github试命名法
	if pathName == "github.com" {
		pathName = tempPath[2]
	}

	if pathName == rootName {
		return true
	}
	return false
}

func getAPPRootModule(APPRoot string) string {
	if APPRootModule != "" {
		return APPRootModule
	}

	// 解析根目录下的go.mod
	file, err := ioutil.ReadFile(path.Join(APPRoot, "go.mod"))
	if err != nil {
		panic("读取go mod 失败" + err.Error())
	}
	reg := regexp.MustCompile(`module[\W]?(\S*)`)
	text := reg.FindAllSubmatch(file, 1)
	// 如果成功匹配
	if len(text) > 0 {
		APPRootModule = string(text[0][1][:])
	} else {
		panic("获取项目模块名称失败")
	}

	return APPRootModule
}

// 如果是github的包
// 如果包的写法是大写，拉下来的时候会被转换成小写并且加！前缀
func pathStrParse(path string) string {
	if path == "" {
		return ""
	}

	if !strings.HasPrefix(path, "github.com") {
		return path
	}
	var r []int32
	for _, v := range path {
		// 如果是大写字符
		if 'A' <= v && v <= 'Z' {
			v += 'a' - 'A'
			r = append(r, '!')
		}
		r = append(r, v)
	}

	return string(r[:])
}

// 如果是github的包
// 如果包的写法是大写，拉下来的时候会被转换成小写并且加！前缀
func pathStrParseBack(path string) string {
	if path == "" {
		return ""
	}

	if !strings.HasPrefix(path, "github.com") {
		return path
	}
	var r []int32
	for k, v := range path {
		// 如果是感叹号
		if '!' == path[k] {
			continue
		}
		if k > 0 && '!' == path[k-1] {
			v += 'A' - 'a'
		}

		r = append(r, v)
	}

	return string(r[:])
}

func GetGoROOT() string {
	if GoRoot != "" {
		return GoRoot
	}
	initFunc()

	return GoRoot
}

func GetGoPath() string {
	if GoPath != "" {
		return GoPath
	}
	initFunc()

	return GoPath
}

func GetGOMODCACHE() string {
	if GOMODCACHE != "" {
		return GOMODCACHE
	}
	initFunc()

	return GOMODCACHE
}

func initFunc() {
	GoRoot = filepath.FromSlash(os.Getenv("GOROOT"))
	GoPath = filepath.FromSlash(os.Getenv("GOPATH"))
	GOMODCACHE = filepath.Join(GoPath, "pkg", "mod")
	if GoRoot == "" {
		GoRoot = GoPath
	}
	if GoRoot == "" {
		panic("请先设置GOROOT环境变量")
	}
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) // os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 获取第三方库文件的dir
func getCacheDir(path string) string {
	temp := strings.Split(path, "/")
	searchPath := filepath.Join(GOMODCACHE, filepath.FromSlash(strings.Join(temp[0:1], "/")))
	if !Exists(searchPath) {
		panic("找不到对应的库" + path)
	}

	// 搜索版本号出现的位置,
	i := 0
	for {
		if len(temp) < i+1 {
			break
		}
		searchPath = filepath.Join(GOMODCACHE, filepath.FromSlash(strings.Join(temp[0:i+1], "/")))
		// 如果不存在有两种可能性， 第一个目录夹杂了版本号，第二个 本来就不存在
		if !Exists(searchPath) {
			break
		}
		i++
	}

	// 重制极限的searchPath
	searchPath = filepath.Join(GOMODCACHE, filepath.FromSlash(strings.Join(temp[0:i], "/")))
	fileInfoList, err := ioutil.ReadDir(searchPath)
	if err != nil {
		panic(err)
	}

	// 取版本号最大的
	dirName := ""
	dirLevel := ""
	for _, v := range fileInfoList {
		if v.IsDir() {
			name := v.Name()
			// 正则匹配文件夹的名字
			reg := regexp.MustCompile(temp[i] + `@(.*)`)
			t := reg.FindAllStringSubmatch(name, -1)
			if len(t) > 0 {
				if len(t[0]) > 1 {
					// 去除各种版本号缓存字符串
					t[0][1] = strings.Split(t[0][1], "-")[0]
					if dirLevel == "" || VersionThan(dirLevel, t[0][1]) < 0 {
						dirName = t[0][0]
						dirLevel = t[0][1]
					}
				}
			}
		}
	}

	if dirName == "" {
		panic("找不到对应的库" + path)
	}

	// 拿到了dirname之后开始拼接
	p := filepath.Join(searchPath, dirName)
	if i < len(temp) {
		p = filepath.Join(p, filepath.FromSlash(strings.Join(temp[i+1:], "/")))
	}

	return p
}

func VersionThan(va string, vb string) int {
	vas := strings.Split(va, ".")
	vbs := strings.Split(vb, ".")

	count := len(vas)
	if len(vas) < len(vbs) {
		count = len(vbs)
	}
	for i := 0; i < count; i++ {
		a := 0
		if len(vas) < i+1 {
			a = 0
		} else {
			a, _ = strconv.Atoi(vas[i])
		}

		b := 0
		if len(vbs) < i+1 {
			b = 0
		} else {
			b, _ = strconv.Atoi(vbs[i])
		}
		if a > b {
			return 1
		}
		if a < b {
			return -1
		}
	}
	return 0
}
