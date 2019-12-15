package validate

import (
	"errors"
	"strings"
)

var required = &requiredValidate{}

type requiredValidate struct {
}

func (c *requiredValidate) Verify(v *Validator, val Val, param string) (pass bool, next bool, err error) {
	//fmt.Println("required", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New("The " + val.Key + " field is required")
	}
	if s, ok := val.Value.(string); ok && strings.TrimSpace(s) == "" {
		return false, false, errors.New("The " + val.Key + " field is required")
	}
	return true, true, nil
}

func init() {
	Register("required", required)
}
