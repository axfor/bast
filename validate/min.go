package validate

import (
	"errors"
	"reflect"
	"strconv"
)

var min = &minValidate{}

type minValidate struct {
}

func (c *minValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("min", val.Key, "=", val.Value)
	if val.Param == "" {
		return true, true, nil
	}
	msg := ""
	if val.Expect == Int {
		if val.Real == reflect.String {
			va, ok := val.Value.(string)
			if ok {
				vv, err := strconv.ParseInt(va, 0, 64)
				if err == nil {
					mv, err := strconv.ParseInt(val.Param, 0, 64)
					ok = err == nil && (vv >= mv)
					if ok {
						return ok, ok, nil
					}
				}
			}
		} else {
			if isIntWithMin(val) {
				return true, true, nil
			}
		}
	} else if val.Expect == String || val.Expect == Email {
		msg = " characters"
		va, ok := val.Value.(string)
		mv, err := strconv.Atoi(val.Param)
		if err != nil {
			return false, false, errors.New(val.Key + " min param is invalid")
		}
		if ok && len(va) >= mv {
			return true, true, nil
		}
	}
	return false, false, errors.New("The " + val.Key + " must be greater than " + val.Param + msg)
}

func isIntWithMin(val Val) bool {
	if v, ok := val.Value.(int); ok {
		mv, err := strconv.Atoi(val.Param)
		return err == nil && v >= mv
	} else if v, ok := val.Value.(int64); ok {
		mv, err := strconv.ParseInt(val.Param, 0, 64)
		return err == nil && (v >= mv)
	} else if v, ok := val.Value.(int32); ok {
		mv, err := strconv.ParseInt(val.Param, 0, 32)
		return err == nil && (v >= int32(mv))
	} else if v, ok := val.Value.(int8); ok {
		mv, err := strconv.ParseInt(val.Param, 0, 8)
		return err == nil && (v >= int8(mv))
	} else if v, ok := val.Value.(uint); ok {
		mv, err := strconv.ParseUint(val.Param, 0, 64)
		return err == nil && (v >= uint(mv))
	} else if v, ok := val.Value.(uint64); ok {
		mv, err := strconv.ParseUint(val.Param, 0, 64)
		return err == nil && (v >= uint64(mv))
	} else if v, ok := val.Value.(uint32); ok {
		mv, err := strconv.ParseUint(val.Param, 0, 64)
		return err == nil && (v >= uint32(mv))
	} else if v, ok := val.Value.(uint8); ok {
		mv, err := strconv.ParseUint(val.Param, 0, 64)
		return err == nil && (v >= uint8(mv))
	}
	return false
}

func init() {
	Register("min", min)
}
