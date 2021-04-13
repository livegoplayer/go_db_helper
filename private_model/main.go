package private_model

import (
	"flag"
	"github.com/livegoplayer/go_db_helper/mysql"
	"github.com/livegoplayer/go_db_helper/private_model/template"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
)

var DalDir = ""
var TemplateFileDir = ""
var BuildFileDir = ""

type Path string

func init() {
	flag.StringVar(&DalDir, "dal_dir", "", "dal_dir")
}

func Parse(APPRoot string) {
	// 这里写死，如果有需要，可以手动传入
	if DalDir == "" {
		DalDir = "repository_template_model"
	}
	TemplateFileDir = ParseDirPathByImport(APPRoot, GetTemplatePath())
	BuildFileDir = ParseDirPathByImport(APPRoot, GetBuildPath())

	if DalDir == "" || TemplateFileDir == "" {
		panic("参数错误")
	}

	CurPath := APPRoot

	dalFiles, _ := ioutil.ReadDir(filepath.Join(CurPath, DalDir))
	for _, item := range dalFiles {
		if item.IsDir() {
			// 检查里面是否有文件，没有的话就不必创建
			fs, _ := filepath.Glob("*")
			if len(fs) == 0 {
				continue
			}
			// 二级文件目录，使用私有方式构建
			fullPath := filepath.Join(CurPath, DalDir, item.Name())
			CpTemplate(filepath.Join(TemplateFileDir, "mysql"), fullPath, item.Name())
			task := NewMysqlTask(fullPath, item.Name())
			task.IsPrivate = true
			task.ModelFilterFunc = func(name string) bool {
				return strings.Contains(name, "Model.go") || defaultFilter(name)
			}
			task.Run()
		}
	}
}

func GetTemplatePath() string {
	var b template.Path
	str := reflect.TypeOf(b).PkgPath()
	return str
}

func GetBuildPath() string {
	var b mysql.DbName
	str := reflect.TypeOf(b).PkgPath()
	return str
}
