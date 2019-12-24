package validate

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
)

var match = &matchValidate{}

var matchRegexp sync.Map

type matchValidate struct {
}

func (c *matchValidate) Verify(v *Validator, val Val) (pass bool, next bool, err error) {
	//fmt.Println("match", val.Key, "=", val.Value)
	if val.Value == nil || val.Param == "" {
		return false, false, errors.New(v.Trans("match", val.TranKey))
	}
	var rr *regexp.Regexp
	r, ok := matchRegexp.Load(val.Param)
	if !ok {
		rr = regexp.MustCompile(val.Param)
		matchRegexp.Store(val.Param, rr)
	} else {
		rr = r.(*regexp.Regexp)
	}
	if vv, ok := val.Value.(string); ok {
		if vv == "" || !rr.MatchString(vv) {
			return false, false, errors.New(v.Trans("match", val.TranKey))
		}
	} else {
		if !rr.MatchString(fmt.Sprintf("%v", val.Value)) {
			return false, false, errors.New(v.Trans("match", val.TranKey))
		}
	}
	return true, true, nil
}

func init() {
	Register("match", match)
}
