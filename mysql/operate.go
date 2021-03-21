package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/livegoplayer/go_helper/utils"
)

func newOperat() *operat {
	return &operat{
		inc: make(map[string]int),
		set: make(map[string]interface{}),
	}
}

type operat struct {
	inc map[string]int
	set map[string]interface{}
}

func (op operat) cumulateInc(col string, num int) {
	if v, ok := op.inc[col]; !ok {
		op.inc[col] = num
	} else {
		op.inc[col] = v + num
	}
}

func (op operat) cumulateSet(col string, val interface{}) {
	op.set[col] = val
}

func (op operat) raw() utils.H {
	raw := make(map[string]interface{})
	for k, v := range op.set {
		raw[k] = v
	}

	for k, v := range op.inc {
		raw[k] = gorm.Expr(fmt.Sprintf("`%s` + ?", k), v)
	}
	return raw
}
