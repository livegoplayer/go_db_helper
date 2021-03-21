package mysql

import (
	"strings"
)

func FieldTypeToGolangType(tp string) string {
	if strings.Contains(tp, "int") {
		return "int64"
	} else if strings.Contains(tp, "decimal") {
		return "float64"
	} else if strings.Contains(tp, "double") {
		return "float64"
	} else if strings.Contains(tp, "float") {
		return "float64"
	} else if strings.Contains(tp, "char") {
		return "string"
	} else if strings.Contains(tp, "text") {
		return "string"
	} else if strings.Contains(tp, "timestamp") {
		return "string"
	} else if strings.Contains(tp, "datetime") {
		return "string"
	} else if strings.Contains(tp, "enum") {
		return "string"
	} else {
		return ""
	}
}

type Query struct {
	build *Build `gorm:"-"    json:"-"`
}

func (q *Query) GetBuild() *Build {
	return q.build
}

func (q *Query) SetBuild(b *Build) {
	q.build = b
}
