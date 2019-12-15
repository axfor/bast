package validate

import (
	"errors"
	"reflect"
	"strconv"
)

var max = &maxValidate{}

type maxValidate struct {
}

func (c *maxValidate) Verify(v *Validator, val Val, param string) (pass bool, next bool, err error) {
	//fmt.Println("max", val.Key, "=", val.Value)
	if param == "" {
		return true, true, nil
	}
	mv, err := strconv.Atoi(param)
	if err != nil {
		return false, false, errors.New(val.Key + " max param is invalid")
	}
	msg := ""
	if val.Expect == Int {
		if val.Real == reflect.String {
			va, ok := val.Value.(string)
			if ok {
				vv, err := strconv.ParseInt(va, 0, 64)
				if err == nil {
					mv, err := strconv.ParseInt(param, 0, 64)
					ok = err == nil && (vv <= mv)
					if ok {
						return ok, ok, nil
					}
				}
			}
		} else {
			if isIntWithMax(val, param) {
				return true, true, nil
			}
		}
	} else if val.Expect == String || val.Expect == Email {
		msg = "characters"
		va, ok := val.Value.(string)
		if ok && len(va) <= mv {
			return true, true, nil
		}
	}
	return false, false, errors.New("The " + val.Key + " must be less than " + param + " " + msg)
}

func isIntWithMax(val interface{}, maxVal string) bool {
	if v, ok := val.(int); ok {
		mv, err := strconv.Atoi(maxVal)
		return err == nil && v <= mv
	} else if v, ok := val.(int64); ok {
		mv, err := strconv.ParseInt(maxVal, 0, 64)
		return err == nil && (v <= mv)
	} else if v, ok := val.(int32); ok {
		mv, err := strconv.ParseInt(maxVal, 0, 32)
		return err == nil && (v <= int32(mv))
	} else if v, ok := val.(int8); ok {
		mv, err := strconv.ParseInt(maxVal, 0, 8)
		return err == nil && (v <= int8(mv))
	} else if v, ok := val.(uint); ok {
		mv, err := strconv.ParseUint(maxVal, 0, 64)
		return err == nil && (v <= uint(mv))
	} else if v, ok := val.(uint64); ok {
		mv, err := strconv.ParseUint(maxVal, 0, 64)
		return err == nil && (v <= uint64(mv))
	} else if v, ok := val.(uint32); ok {
		mv, err := strconv.ParseUint(maxVal, 0, 64)
		return err == nil && (v <= uint32(mv))
	} else if v, ok := val.(uint8); ok {
		mv, err := strconv.ParseUint(maxVal, 0, 64)
		return err == nil && (v <= uint8(mv))
	}
	return false
}

func init() {
	Register("max", max)
}
