package mysql

//开启一个事务执行回调，推荐使用，防止错误使用commit和rollback
func Transaction(argus ...interface{}) {
	var con string
	var cb func()
	switch argus[0].(type) {
	case string:
		con = argus[0].(string)
		cb = argus[1].(func())
	default:
		con = DefaultCon
		cb = argus[0].(func())
	}
	Begin(con)
	defer func() {
		if r := recover(); r != nil {
			Rollback(con)
			//回滚后，继续抛出错误，给外层逻辑处理
			panic(r)
		} else {
			Commit(con)
		}
	}()
	cb()
}

//推荐使用Transaction函数
func Begin(cons ...interface{}) {
	var con string
	if len(con) == 0 {
		con = DefaultCon
	} else {
		con = cons[0].(string)
	}
	gid := getSlow()
	h := newHandle(gid, con).trans
	h.idx++
}

//推荐使用Transaction函数
func Commit(cons ...interface{}) {
	var con string
	if len(con) == 0 {
		con = DefaultCon
	} else {
		con = cons[0].(string)
	}
	gid := getSlow()
	handle := getHandle(gid, con).trans
	if handle == nil {
		//不处于事务中，直接返回
		Logger.Error("事务错误", "尝试提交不存在的事务")
		return
	}

	handle.idx--
	if handle.idx > 0 {
		//嵌套事务未结束
		return
	}
	handle.db.Commit()
	removeTransactionHandle(gid, con)
}

//推荐使用Transaction函数
func Rollback(cons ...interface{}) {
	var con string
	if len(con) == 0 {
		con = DefaultCon
	} else {
		con = cons[0].(string)
	}
	gid := getSlow()
	handle := getHandle(gid, con).trans
	if handle == nil {
		//不处于事务中，直接返回
		Logger.Error("事务错误", "尝试提交不存在的事务")
		return
	}

	handle.idx--
	if handle.idx > 0 {
		//嵌套事务未结束
		return
	}
	handle.db.Rollback()
	removeTransactionHandle(gid, con)
}

//当事务结束时，移除事务的句柄
func removeTransactionHandle(gid int64, con string) {
	removeHandle(gid, con)
}
