package mysql

import (
	"bytes"
	"github.com/jinzhu/gorm"
	"runtime"
	"strconv"
)

type transaction struct {
	db  *gorm.DB //数据库句柄
	idx int      //事务嵌套的计数值
}

const (
	_ADD    = "add"
	_REMOVE = "remove"
	_FIND   = "find"
)

type task struct {
	backch backChan
	key    int64
	con    DbName
	opera  string
}

type backTransaction struct {
	trans *transaction
	key   int64
}

type backChan chan backTransaction
type sendChan chan *task

var send sendChan
var transactionHandleMap map[string]*transaction

func newHandle(key int64, con DbName) backTransaction {
	ch := make(backChan)
	send <- &task{backch: ch, opera: _ADD, key: key, con: con}
	return <-ch
}

func getHandle(key int64, con DbName) backTransaction {
	ch := make(backChan)
	send <- &task{key: key, backch: ch, opera: _FIND, con: con}
	return <-ch
}

func removeHandle(key int64, con DbName) {
	ch := make(backChan)
	send <- &task{key: key, backch: ch, opera: _REMOVE, con: con}
	<-ch
	return
}

func extractGID(s []byte) int64 {
	s = s[len("goroutine "):]
	s = s[:bytes.IndexByte(s, ' ')]
	gid, _ := strconv.ParseInt(string(s), 10, 64)
	return gid
}

//用于获取协程id, 效率较慢，不应该被非常频繁的调用(但其实调用一次也才几千纳秒)
func getSlow() int64 {
	var buf [64]byte
	return extractGID(buf[:runtime.Stack(buf[:], false)])
}

func init() {
	transactionHandleMap = make(map[string]*transaction)
	send = make(sendChan)
	go func(ch sendChan) {
		for {
			task := <-ch
			mapIdx := strconv.FormatInt(task.key, 10) + "_" + task.con.String()
			switch task.opera {
			case _ADD:
				oh, ok := transactionHandleMap[mapIdx]
				if ok {
					task.backch <- backTransaction{key: task.key, trans: oh}
					continue
				}
				//fmt.Println(_db_list[task.con])
				tx := _db_list[task.con][WRITE].Begin()
				h := &transaction{db: tx}
				transactionHandleMap[mapIdx] = h
				task.backch <- backTransaction{key: task.key, trans: h}
			case _FIND:
				h, ok := transactionHandleMap[mapIdx]
				if !ok {
					task.backch <- backTransaction{key: task.key, trans: nil}
				} else {
					task.backch <- backTransaction{key: task.key, trans: h}
				}
			case _REMOVE:
				delete(transactionHandleMap, mapIdx)
				task.backch <- backTransaction{}
			}
		}
	}(send)
}
