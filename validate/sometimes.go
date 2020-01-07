//Copyright 2018 The axx Authors. All rights reserved.

package validate

import (
	"time"
)

var sometimes = &sometimesValidate{}

type sometimesValidate struct {
}

func (c *sometimesValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("sometimes", val.Key, "=", val.Value)
	if val.Value == nil {
		return true, false, nil
	}
	if v, ok := val.Value.(*string); ok && v == nil {
		return true, false, nil
	} else if v, ok := val.Value.(*time.Time); ok && v == nil {
		return true, false, nil
	}
	return true, true, nil
}

func init() {
	Register("sometimes", sometimes)
}
