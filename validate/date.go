package validate

import (
	"errors"
	"time"
)

var loc *time.Location

func init() {
	loc, _ = time.LoadLocation("Local")
}

var date = &dateValidate{}

type dateValidate struct {
}

func (c *dateValidate) Verify(v *Validator, val Val, param string) (pass bool, next bool, err error) {
	//fmt.Println("date", val.Key, "=", val.Value)
	if val.Value == nil {
		return false, false, errors.New("The " + val.Key + " is not a valid date")
	}
	if _, ok := val.Value.(time.Time); ok {
		return true, true, nil
	}
	if v, ok := val.Value.(string); ok {
		if v == "" || len(v) < 8 {
			return false, false, errors.New("The " + val.Key + " is not a valid date")
		}
		ok = s2t(v)
		if ok {
			return true, true, nil
		}
	}
	return false, false, errors.New("The " + val.Key + " is not a valid date")
}

func s2t(s string) bool {
	layout := ""
	var err error
	l := len(s)
	var v time.Time
	if l <= 8 {
		layout = "2006-1-2"
	} else if l <= 10 {
		layout = "2006-01-02"
	} else if l <= 16 {
		layout = "2006-01-02 15:04"
	} else {
		layout = "2006-01-02 15:04:05"
	}
	v, err = time.ParseInLocation(layout, s, loc)
	vv := v.String()
	return err == nil && !v.IsZero() && vv != ""
}
func init() {
	Register("date", date)
}
