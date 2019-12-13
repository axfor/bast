package validate

import "errors"

import "strings"

var required = &requiredValidate{}

type requiredValidate struct {
}

func (c *requiredValidate) Verify(v *Validator, key string, val interface{}) (bool, bool) {

	if val == nil {
		v.SetError(errors.New("key is required"))
		return true, false
	}
	if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
		v.SetError(errors.New("key is required"))
		return true, false
	}
	return false, true
}

func init() {
	Register("required", required)
}
