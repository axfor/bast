package validate

import (
	"errors"
	"strconv"
	"strings"
)

var integer = &integerValidate{}

type integerValidate struct {
}

func (c *integerValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("integer", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New("The " + val.Key + " field is int")
	}
	if s, ok := val.Value.(string); ok {
		if strings.TrimSpace(s) == "" {
			return false, false, errors.New("The " + val.Key + " field is int")
		}
		_, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			_, err = strconv.ParseInt(s, 0, 64)
		}
		if err != nil {
			err = errors.New("The " + val.Key + " field is int")
		} else {
			//val.Value = vv
		}
		return err == nil, err == nil, err
	} else if isInteger(val) {
		return true, true, nil
	}
	return false, false, nil
}

func isInteger(val Val) bool {
	ok := false
	if _, ok = val.Value.(int); ok {
	} else if _, ok = val.Value.(int64); ok {
	} else if _, ok = val.Value.(int32); ok {
	} else if _, ok = val.Value.(int8); ok {
	} else if _, ok = val.Value.(uint); ok {
	} else if _, ok = val.Value.(uint64); ok {
	} else if _, ok = val.Value.(uint32); ok {
	} else if _, ok = val.Value.(uint8); ok {
	}
	return ok
}

func init() {
	Register("int", integer)
}
