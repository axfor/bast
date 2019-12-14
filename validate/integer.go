package validate

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var integer = &integerValidate{}

type integerValidate struct {
}

func (c *integerValidate) Verify(v *Validator, key string, val interface{}, kind reflect.Kind, param ...string) (bool, bool) {
	//fmt.Println("integer", key, "=", val)
	if val == nil {
		v.SetError(errors.New(key + " is not int"))
		return false, false
	}
	if s, ok := val.(string); ok {
		if strings.TrimSpace(s) == "" {
			v.SetError(errors.New(key + " is not int"))
			return false, false
		}
		_, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			_, err = strconv.ParseInt(s, 0, 64)
		}
		if err != nil {
			v.SetError(errors.New(key + " is not int"))
		}
		return err == nil, err == nil
	} else if isInteger(val) {
		return true, true
	}
	return false, false
}

func isInteger(val interface{}) bool {
	ok := false
	if _, ok = val.(int); ok {
	} else if _, ok = val.(int64); ok {
	} else if _, ok = val.(int32); ok {
	} else if _, ok = val.(int8); ok {
	} else if _, ok = val.(uint); ok {
	} else if _, ok = val.(uint64); ok {
	} else if _, ok = val.(uint32); ok {
	} else if _, ok = val.(uint8); ok {
	}
	return ok
}

func init() {
	Register("int", integer)
}
