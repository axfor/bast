package validate

import (
	"errors"
	"reflect"
	"strconv"
)

var min = &minValidate{}

type minValidate struct {
}

func (c *minValidate) Verify(v *Validator, key string, val interface{}, kind reflect.Kind, param ...string) (bool, bool) {
	//fmt.Println("min", key, "=", val)
	if param == nil {
		return true, true
	}
	mi := "0"
	mv := 0
	if len(param) > 0 {
		mi = param[0]
		mv, _ = strconv.Atoi(mi)
	}
	msg := ""
	if kind == Int {
		if isint(val, mi) {
			return true, true
		}
		va, ok := val.(string)
		if ok {
			vv, err := strconv.ParseInt(va, 0, 64)
			if err == nil {
				mv, err := strconv.ParseInt(mi, 0, 64)
				ok = err == nil && (vv > mv)
				if ok {
					return ok, ok
				}
			}
		}
	} else if kind == String {
		msg = " length "
		va, ok := val.(string)
		if ok && len(va) > mv {
			return true, true
		}
	}
	v.SetError(errors.New(key + " min " + msg + " is " + mi))
	return false, false
}

func isint(val interface{}, minVal string) bool {
	if v, ok := val.(int); ok {
		mv, err := strconv.Atoi(minVal)
		return err == nil && v > mv
	} else if v, ok := val.(int64); ok {
		mv, err := strconv.ParseInt(minVal, 0, 64)
		return err == nil && (v > mv)
	} else if v, ok := val.(int32); ok {
		mv, err := strconv.ParseInt(minVal, 0, 32)
		return err == nil && (v > int32(mv))
	} else if v, ok := val.(int8); ok {
		mv, err := strconv.ParseInt(minVal, 0, 8)
		return err == nil && (v > int8(mv))
	} else if v, ok := val.(uint); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v > uint(mv))
	} else if v, ok := val.(uint64); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v > uint64(mv))
	} else if v, ok := val.(uint32); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v > uint32(mv))
	} else if v, ok := val.(uint8); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v > uint8(mv))
	}
	return false
}

func init() {
	Register("min", min)
}
