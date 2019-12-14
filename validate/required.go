package validate

import (
	"errors"
	"strings"
)

var required = &requiredValidate{}

type requiredValidate struct {
}

func (c *requiredValidate) Verify(v *Validator, key string, val interface{}, param ...string) (bool, bool) {
	//fmt.Println("required", key, "=", val)
	if val == nil {
		v.SetError(errors.New(key + " is required"))
		return false, false
	}
	if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
		v.SetError(errors.New(key + " is required"))
		return false, false
	}
	return true, true
}

func init() {
	Register("required", required)
}
