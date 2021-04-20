package private_model

import (
	"sort"
)

/*
此文件为自动生成，所有修改都不会生效
*/

func (s DefineFields) PluckStructKey() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.StructKey
	}
	return list
}

func (s DefineFields) PluckKey() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Key
	}
	return list
}

func (s DefineFields) PluckType() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Type
	}
	return list
}

func (s DefineFields) PluckNumber() []bool {
	list := make([]bool, len(s))
	for i, v := range s {
		list[i] = v.Number
	}
	return list
}

func (s DefineFields) PluckUniStructKey() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.StructKey]
		if !ok {
			uniMap[v.StructKey] = true
			list = append(list, v.StructKey)
		}
	}
	return list
}

func (s DefineFields) PluckUniKey() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Key]
		if !ok {
			uniMap[v.Key] = true
			list = append(list, v.Key)
		}
	}
	return list
}

func (s DefineFields) PluckUniType() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Type]
		if !ok {
			uniMap[v.Type] = true
			list = append(list, v.Type)
		}
	}
	return list
}

func (s DefineFields) PluckUniNumber() []bool {
	uniMap := make(map[bool]bool)
	list := make([]bool, 0)
	for _, v := range s {
		_, ok := uniMap[v.Number]
		if !ok {
			uniMap[v.Number] = true
			list = append(list, v.Number)
		}
	}
	return list
}

func (s DefineFields) GroupByStructKey() map[string]DefineFields {
	m := make(map[string]DefineFields)
	for _, v := range s {
		if _, ok := m[v.StructKey]; !ok {
			m[v.StructKey] = make(DefineFields, 0)
		}
		m[v.StructKey] = append(m[v.StructKey], v)
	}
	return m
}

func (s DefineFields) GroupByKey() map[string]DefineFields {
	m := make(map[string]DefineFields)
	for _, v := range s {
		if _, ok := m[v.Key]; !ok {
			m[v.Key] = make(DefineFields, 0)
		}
		m[v.Key] = append(m[v.Key], v)
	}
	return m
}

func (s DefineFields) GroupByType() map[string]DefineFields {
	m := make(map[string]DefineFields)
	for _, v := range s {
		if _, ok := m[v.Type]; !ok {
			m[v.Type] = make(DefineFields, 0)
		}
		m[v.Type] = append(m[v.Type], v)
	}
	return m
}

func (s DefineFields) GroupByNumber() map[bool]DefineFields {
	m := make(map[bool]DefineFields)
	for _, v := range s {
		if _, ok := m[v.Number]; !ok {
			m[v.Number] = make(DefineFields, 0)
		}
		m[v.Number] = append(m[v.Number], v)
	}
	return m
}

func (s DefineFields) SortByFunc(f func(i, j int) bool) DefineFields {
	sort.SliceStable(s, f)
	return s
}

func (s DefineFields) Filter(f func(item DefineField) bool) DefineFields {
	m := make(DefineFields, 0)
	for _, v := range s {
		if f(v) {
			m = append(m, v)
		}
	}
	return m
}

func (s DefineFields) KeyByStructKey() map[string]DefineField {
	m := make(map[string]DefineField)
	for _, v := range s {
		m[v.StructKey] = v
	}
	return m
}

func (s DefineFields) KeyByKey() map[string]DefineField {
	m := make(map[string]DefineField)
	for _, v := range s {
		m[v.Key] = v
	}
	return m
}

func (s DefineFields) KeyByType() map[string]DefineField {
	m := make(map[string]DefineField)
	for _, v := range s {
		m[v.Type] = v
	}
	return m
}

func (s DefineFields) KeyByNumber() map[bool]DefineField {
	m := make(map[bool]DefineField)
	for _, v := range s {
		m[v.Number] = v
	}
	return m
}
