//Copyright 2018 The axx Authors. All rights reserved.

package validate

import (
	"errors"
	"strings"
	"time"
)

var required = &requiredValidate{}

type requiredValidate struct {
}

func (c *requiredValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("required", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New(v.Trans("required", val.TranKey))
	}
	if vv, ok := val.Value.(string); ok && (vv == "" || strings.TrimSpace(vv) == "") {
		return false, false, errors.New(v.Trans("required", val.TranKey))
	} else if vv, ok := val.Value.(time.Time); ok && vv.IsZero() {
		return false, false, errors.New(v.Trans("required", val.TranKey))
	}
	return true, true, nil
}

func init() {
	Register("required", required)
}
