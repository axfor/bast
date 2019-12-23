package validate

import (
	"errors"
	"fmt"
	"regexp"
)

var email = &emailValidate{}

var emailRegexp = regexp.MustCompile(`^[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[a-zA-Z0-9](?:[\w-]*[\w])?$`)

type emailValidate struct {
}

func (c *emailValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("mail", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New(v.Trans("email", val.TranKey))
	}
	if vv, ok := val.Value.(string); ok {
		if vv == "" || !emailRegexp.MatchString(vv) {
			return false, false, errors.New(v.Trans("email", val.TranKey))
		}
	} else {
		if !emailRegexp.MatchString(fmt.Sprintf("%v", val.Value)) {
			return false, false, errors.New(v.Trans("email", val.TranKey))
		}
	}
	return true, true, nil
}

func init() {
	Register("email", email)
}
