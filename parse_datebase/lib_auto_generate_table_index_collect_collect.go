package parse_datebase

import (
	"sort"
)

/*
此文件为自动生成，所有修改都不会生效
*/

func (s TableIndexCollect) PluckTable() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Table
	}
	return list
}

func (s TableIndexCollect) PluckNonUnique() []int64 {
	list := make([]int64, len(s))
	for i, v := range s {
		list[i] = v.NonUnique
	}
	return list
}

func (s TableIndexCollect) PluckKeyName() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.KeyName
	}
	return list
}

func (s TableIndexCollect) PluckSeqInIndex() []int64 {
	list := make([]int64, len(s))
	for i, v := range s {
		list[i] = v.SeqInIndex
	}
	return list
}

func (s TableIndexCollect) PluckColumnName() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.ColumnName
	}
	return list
}

func (s TableIndexCollect) PluckCollation() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Collation
	}
	return list
}

func (s TableIndexCollect) PluckCardinality() []int64 {
	list := make([]int64, len(s))
	for i, v := range s {
		list[i] = v.Cardinality
	}
	return list
}

func (s TableIndexCollect) PluckSubPart() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.SubPart
	}
	return list
}

func (s TableIndexCollect) PluckPacked() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Packed
	}
	return list
}

func (s TableIndexCollect) PluckIndexType() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.IndexType
	}
	return list
}

func (s TableIndexCollect) PluckComment() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.Comment
	}
	return list
}

func (s TableIndexCollect) PluckIndexComment() []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.IndexComment
	}
	return list
}

func (s TableIndexCollect) PluckUniTable() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Table]
		if !ok {
			uniMap[v.Table] = true
			list = append(list, v.Table)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniNonUnique() []int64 {
	uniMap := make(map[int64]bool)
	list := make([]int64, 0)
	for _, v := range s {
		_, ok := uniMap[v.NonUnique]
		if !ok {
			uniMap[v.NonUnique] = true
			list = append(list, v.NonUnique)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniKeyName() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.KeyName]
		if !ok {
			uniMap[v.KeyName] = true
			list = append(list, v.KeyName)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniSeqInIndex() []int64 {
	uniMap := make(map[int64]bool)
	list := make([]int64, 0)
	for _, v := range s {
		_, ok := uniMap[v.SeqInIndex]
		if !ok {
			uniMap[v.SeqInIndex] = true
			list = append(list, v.SeqInIndex)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniColumnName() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.ColumnName]
		if !ok {
			uniMap[v.ColumnName] = true
			list = append(list, v.ColumnName)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniCollation() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Collation]
		if !ok {
			uniMap[v.Collation] = true
			list = append(list, v.Collation)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniCardinality() []int64 {
	uniMap := make(map[int64]bool)
	list := make([]int64, 0)
	for _, v := range s {
		_, ok := uniMap[v.Cardinality]
		if !ok {
			uniMap[v.Cardinality] = true
			list = append(list, v.Cardinality)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniSubPart() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.SubPart]
		if !ok {
			uniMap[v.SubPart] = true
			list = append(list, v.SubPart)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniPacked() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Packed]
		if !ok {
			uniMap[v.Packed] = true
			list = append(list, v.Packed)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniIndexType() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.IndexType]
		if !ok {
			uniMap[v.IndexType] = true
			list = append(list, v.IndexType)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniComment() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.Comment]
		if !ok {
			uniMap[v.Comment] = true
			list = append(list, v.Comment)
		}
	}
	return list
}

func (s TableIndexCollect) PluckUniIndexComment() []string {
	uniMap := make(map[string]bool)
	list := make([]string, 0)
	for _, v := range s {
		_, ok := uniMap[v.IndexComment]
		if !ok {
			uniMap[v.IndexComment] = true
			list = append(list, v.IndexComment)
		}
	}
	return list
}

func (s TableIndexCollect) GroupByTable() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.Table]; !ok {
			m[v.Table] = make(TableIndexCollect, 0)
		}
		m[v.Table] = append(m[v.Table], v)
	}
	return m
}

func (s TableIndexCollect) GroupByNonUnique() map[int64]TableIndexCollect {
	m := make(map[int64]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.NonUnique]; !ok {
			m[v.NonUnique] = make(TableIndexCollect, 0)
		}
		m[v.NonUnique] = append(m[v.NonUnique], v)
	}
	return m
}

func (s TableIndexCollect) GroupByKeyName() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.KeyName]; !ok {
			m[v.KeyName] = make(TableIndexCollect, 0)
		}
		m[v.KeyName] = append(m[v.KeyName], v)
	}
	return m
}

func (s TableIndexCollect) GroupBySeqInIndex() map[int64]TableIndexCollect {
	m := make(map[int64]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.SeqInIndex]; !ok {
			m[v.SeqInIndex] = make(TableIndexCollect, 0)
		}
		m[v.SeqInIndex] = append(m[v.SeqInIndex], v)
	}
	return m
}

func (s TableIndexCollect) GroupByColumnName() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.ColumnName]; !ok {
			m[v.ColumnName] = make(TableIndexCollect, 0)
		}
		m[v.ColumnName] = append(m[v.ColumnName], v)
	}
	return m
}

func (s TableIndexCollect) GroupByCollation() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.Collation]; !ok {
			m[v.Collation] = make(TableIndexCollect, 0)
		}
		m[v.Collation] = append(m[v.Collation], v)
	}
	return m
}

func (s TableIndexCollect) GroupByCardinality() map[int64]TableIndexCollect {
	m := make(map[int64]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.Cardinality]; !ok {
			m[v.Cardinality] = make(TableIndexCollect, 0)
		}
		m[v.Cardinality] = append(m[v.Cardinality], v)
	}
	return m
}

func (s TableIndexCollect) GroupBySubPart() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.SubPart]; !ok {
			m[v.SubPart] = make(TableIndexCollect, 0)
		}
		m[v.SubPart] = append(m[v.SubPart], v)
	}
	return m
}

func (s TableIndexCollect) GroupByPacked() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.Packed]; !ok {
			m[v.Packed] = make(TableIndexCollect, 0)
		}
		m[v.Packed] = append(m[v.Packed], v)
	}
	return m
}

func (s TableIndexCollect) GroupByIndexType() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.IndexType]; !ok {
			m[v.IndexType] = make(TableIndexCollect, 0)
		}
		m[v.IndexType] = append(m[v.IndexType], v)
	}
	return m
}

func (s TableIndexCollect) GroupByComment() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.Comment]; !ok {
			m[v.Comment] = make(TableIndexCollect, 0)
		}
		m[v.Comment] = append(m[v.Comment], v)
	}
	return m
}

func (s TableIndexCollect) GroupByIndexComment() map[string]TableIndexCollect {
	m := make(map[string]TableIndexCollect)
	for _, v := range s {
		if _, ok := m[v.IndexComment]; !ok {
			m[v.IndexComment] = make(TableIndexCollect, 0)
		}
		m[v.IndexComment] = append(m[v.IndexComment], v)
	}
	return m
}

func (s TableIndexCollect) SortByFunc(f func(i, j int) bool) TableIndexCollect {
	sort.SliceStable(s, f)
	return s
}

func (s TableIndexCollect) Filter(f func(item TableIndex) bool) TableIndexCollect {
	m := make(TableIndexCollect, 0)
	for _, v := range s {
		if f(v) {
			m = append(m, v)
		}
	}
	return m
}

func (s TableIndexCollect) KeyByTable() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.Table] = v
	}
	return m
}

func (s TableIndexCollect) KeyByNonUnique() map[int64]TableIndex {
	m := make(map[int64]TableIndex)
	for _, v := range s {
		m[v.NonUnique] = v
	}
	return m
}

func (s TableIndexCollect) KeyByKeyName() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.KeyName] = v
	}
	return m
}

func (s TableIndexCollect) KeyBySeqInIndex() map[int64]TableIndex {
	m := make(map[int64]TableIndex)
	for _, v := range s {
		m[v.SeqInIndex] = v
	}
	return m
}

func (s TableIndexCollect) KeyByColumnName() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.ColumnName] = v
	}
	return m
}

func (s TableIndexCollect) KeyByCollation() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.Collation] = v
	}
	return m
}

func (s TableIndexCollect) KeyByCardinality() map[int64]TableIndex {
	m := make(map[int64]TableIndex)
	for _, v := range s {
		m[v.Cardinality] = v
	}
	return m
}

func (s TableIndexCollect) KeyBySubPart() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.SubPart] = v
	}
	return m
}

func (s TableIndexCollect) KeyByPacked() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.Packed] = v
	}
	return m
}

func (s TableIndexCollect) KeyByIndexType() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.IndexType] = v
	}
	return m
}

func (s TableIndexCollect) KeyByComment() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.Comment] = v
	}
	return m
}

func (s TableIndexCollect) KeyByIndexComment() map[string]TableIndex {
	m := make(map[string]TableIndex)
	for _, v := range s {
		m[v.IndexComment] = v
	}
	return m
}
