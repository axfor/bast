package validate

import (
	"errors"
	"strings"
	"time"
)

var required = &requiredValidate{}

type requiredValidate struct {
}

func (c *requiredValidate) Verify(v *Validator, val Val, param string) (pass bool, next bool, err error) {
	//fmt.Println("required", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New("The " + val.Key + " field is required")
	}
	if v, ok := val.Value.(string); ok && (v == "" || strings.TrimSpace(v) == "") {
		return false, false, errors.New("The " + val.Key + " field is required")
	} else if v, ok := val.Value.(time.Time); ok && v.IsZero() {
		return false, false, errors.New("The " + val.Key + " field is required")
	}
	return true, true, nil
}

func init() {
	Register("required", required)
}
