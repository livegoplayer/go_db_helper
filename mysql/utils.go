package mysql

import (
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

func AsString(v interface{}) string {
	switch v.(type) {
	case uint32:
		return strconv.FormatInt(int64(v.(uint32)), 10)
	case uint64:
		return strconv.FormatInt(int64(v.(uint64)), 10)
	case int:
		return strconv.Itoa(v.(int))
	case int32:
		return strconv.Itoa(int(v.(int32)))
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case float64:
		return strconv.FormatFloat(v.(float64), 'E', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v.(float32)), 'E', -1, 64)
	case string:
		return v.(string)
	case bool:
		if v.(bool) {
			return "true"
		} else {
			return "false"
		}
	case map[string]interface{}:
		return JsonEncode(v)
	default:
		return ""
	}
}

// 如果是字符串，则转化为int64, 浮点数丢失精度，无法转化则为0
func AsInt64(v interface{}) int64 {
	switch v.(type) {
	case uint32:
		return int64(v.(uint32))
	case uint64:
		return int64(v.(uint64))
	case int:
		return int64(v.(int))
	case int32:
		return int64(v.(int32))
	case int64:
		return v.(int64)
	case float64:
		return int64(v.(float64))
	case float32:
		return int64(v.(float32))
	case json.Number:
		vv, err := strconv.ParseInt(string(v.(json.Number)), 10, 64)
		if err != nil {
			return 0
		}
		return vv
	case string:
		vv, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return 0
		}
		return vv
	case bool:
		if v.(bool) {
			return 1
		} else {
			return 0
		}
	default:
		return 0
	}
}

func JsonEncode(data interface{}) string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonByte, err := json.Marshal(&data)
	if err != nil {
		fmt.Printf("json加密出错:" + err.Error())
	}
	return string(jsonByte[:])
}

func StrInSlice(ss []string, str string) (exist bool) {
	if len(ss) == 0 {
		return
	}

	for _, s := range ss {
		if s == str {
			exist = true
			break
		}
	}

	return
}
