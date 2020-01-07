//Copyright 2018 The axx Authors. All rights reserved.

package validate

import (
	"errors"
	"fmt"
	"regexp"
)

var ip = &ipValidate{}

var ipRegexp = regexp.MustCompile(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`)

type ipValidate struct {
}

func (c *ipValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("ip", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New(v.Trans("ip", val.TranKey))
	}
	if vv, ok := val.Value.(string); ok {
		if vv == "" || !ipRegexp.MatchString(vv) {
			return false, false, errors.New(v.Trans("ip", val.TranKey))
		}
	} else {
		if !ipRegexp.MatchString(fmt.Sprintf("%v", val.Value)) {
			return false, false, errors.New(v.Trans("ip", val.TranKey))
		}
	}
	return true, true, nil
}

func init() {
	Register("ip", ip)
}
