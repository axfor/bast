package validate

import (
	"errors"
	"reflect"
	"strconv"
)

var min = &minValidate{}

type minValidate struct {
}

func (c *minValidate) Verify(v *Validator, val Val, param string) (pass bool, next bool, err error) {
	//fmt.Println("min", val.Key, "=", val.Value)
	if param == "" {
		return true, true, nil
	}
	mv, err := strconv.Atoi(param)
	if err != nil {
		return false, false, errors.New(val.Key + " min param is invalid")
	}
	msg := ""
	if val.Expect == Int {
		if val.Real == reflect.String {
			va, ok := val.Value.(string)
			if ok {
				vv, err := strconv.ParseInt(va, 0, 64)
				if err == nil {
					mv, err := strconv.ParseInt(param, 0, 64)
					ok = err == nil && (vv >= mv)
					if ok {
						return ok, ok, nil
					}
				}
			}
		} else {
			if isIntWithMin(val, param) {
				return true, true, nil
			}
		}
	} else if val.Expect == String || val.Expect == Email {
		msg = "characters"
		va, ok := val.Value.(string)
		if ok && len(va) >= mv {
			return true, true, nil
		}
	}
	return false, false, errors.New("The " + val.Key + " must be greater than " + param + " " + msg)
}

func isIntWithMin(val interface{}, minVal string) bool {
	if v, ok := val.(int); ok {
		mv, err := strconv.Atoi(minVal)
		return err == nil && v >= mv
	} else if v, ok := val.(int64); ok {
		mv, err := strconv.ParseInt(minVal, 0, 64)
		return err == nil && (v >= mv)
	} else if v, ok := val.(int32); ok {
		mv, err := strconv.ParseInt(minVal, 0, 32)
		return err == nil && (v >= int32(mv))
	} else if v, ok := val.(int8); ok {
		mv, err := strconv.ParseInt(minVal, 0, 8)
		return err == nil && (v >= int8(mv))
	} else if v, ok := val.(uint); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v >= uint(mv))
	} else if v, ok := val.(uint64); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v >= uint64(mv))
	} else if v, ok := val.(uint32); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v >= uint32(mv))
	} else if v, ok := val.(uint8); ok {
		mv, err := strconv.ParseUint(minVal, 0, 64)
		return err == nil && (v >= uint8(mv))
	}
	return false
}

func init() {
	Register("min", min)
}
