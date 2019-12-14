package validate

import (
	"errors"
	"strconv"
	"strings"
)

var integer = &integerValidate{}

type integerValidate struct {
}

func (c *integerValidate) Verify(v *Validator, key string, val interface{}, param ...string) (bool, bool) {
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
	} else if _, ok := val.(int); ok {
		return true, true
	} else if _, ok := val.(int8); ok {
		return true, true
	} else if _, ok := val.(int32); ok {
		return true, true
	} else if _, ok := val.(int64); ok {
		return true, true
	} else if _, ok := val.(uint); ok {
		return true, true
	} else if _, ok := val.(uint8); ok {
		return true, true
	} else if _, ok := val.(uint32); ok {
		return true, true
	} else if _, ok := val.(uint64); ok {
		return true, true
	}
	return false, false
}

func init() {
	Register("int", integer)
}
