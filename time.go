//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

//Time yyyy-MM-dd HH:mm:ss format time
//1: auto handle string to time
type Time struct {
	time.Time
}

//MarshalJSON JSON MarshalJSON
func (t Time) MarshalJSON() ([]byte, error) {
	stamp := "\"\""
	if !t.IsZero() {
		stamp = fmt.Sprintf("\"%s\"", t.Time.Format("2006-01-02 15:04:05"))
	}

	return []byte(stamp), nil
}

//UnmarshalJSON JSON UnmarshalJSON
func (t *Time) UnmarshalJSON(b []byte) error {
	tt, err := byteToTime(b, "")
	if err == nil {
		*t = tt
	}
	return err
}

//Value support  sql.Driver interface
func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

//Scan support scan
func (t *Time) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Time{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to Time", v)
}

//T get time.Time
//if zero time return nil
func (t *Time) T() *time.Time {
	if !t.IsZero() {
		return &t.Time
	}
	return nil
}

//Now returns the current local time.
func Now() Time {
	tt := Time{Time: time.Now()}
	return tt
}

//Nows returns the current local *time.
func Nows() *Time {
	tt := Time{Time: time.Now()}
	return &tt
}

//NowsNil returns the nil *bast.time.
func NowsNil() *Time {
	var t Time
	return &t
}

//NowTime returns the current local *time.
func NowTime() *Time {
	tt := Time{Time: time.Now()}
	return &tt
}

//TimeByTime time.Time to Time
func TimeByTime(t time.Time) Time {
	tt := Time{Time: t}
	return Time(tt)
}

//TimesByTime time.Time to *Time
func TimesByTime(t time.Time) *Time {
	tt := TimeByTime(t)
	return &tt
}

//TimeWithString  string to Time
func TimeWithString(t string, layout ...string) (Time, error) {
	l := ""
	if layout != nil {
		l = layout[0]
	}
	return strToTime(t, l)
}

//strToTime
func strToTime(t string, layout string) (Time, error) {
	return byteToTime([]byte(t), layout)
}

//byteToTime
func byteToTime(b []byte, layout string) (Time, error) {
	var t Time
	var err error
	if b != nil && len(b) > 0 {
		s := strings.Trim(string(b), "\"")
		l := len(s)
		var v time.Time
		loc, _ := time.LoadLocation("Local")
		if l > 0 {
			if l <= 10 {
				if layout == "" {
					layout = "2006-01-02"
				}
				v, err = time.ParseInLocation(layout, s, loc)
			} else {
				if layout == "" {
					layout = "2006-01-02 15:04:05"
				}
				v, err = time.ParseInLocation(layout, s, loc)
			}
			if err == nil {
				t = Time{Time: v}
			} else {
				err = errors.New("bytes to time error")
			}
		}
	}
	return t, err
}

//String
func (t *Time) String() string {
	return t.Time.Format("2006-01-02 15:04:05")
}

//Format return format string time
//layout support yyyy、MM、dd、hh、HH、mm、ss
func (t *Time) Format(layout string) string {
	if layout == "" {
		layout = "2006-01-02"
	} else {
		layout = strings.Replace(layout, "yyyy", "2006", 1)
		layout = strings.Replace(layout, "MM", "01", 1)
		layout = strings.Replace(layout, "dd", "01", 1)
		layout = strings.Replace(layout, "hh", "01", 1)
		layout = strings.Replace(layout, "mm", "01", 1)
		layout = strings.Replace(layout, "ss", "01", 1)
	}
	return t.Time.Format(layout)
}

//Date yyyy-MM-dd format date
//1: auto handle string to time
type Date Time

//NowDate returns the current local date.
func NowDate() Date {
	return Date(Now())
}

//NowDates returns the current local *date.
func NowDates() *Date {
	d := NowDate()
	return &d
}

//DateByTime time.Time to Date
func DateByTime(t time.Time) Date {
	tt := Time{Time: t}
	return Date(tt)
}

//DatesByTime time.Time to *Date
func DatesByTime(t time.Time) *Date {
	tt := DateByTime(t)
	return &tt
}

//MarshalJSON JSON MarshalJSON
func (t Date) MarshalJSON() ([]byte, error) {
	stamp := "\"\""
	if !t.IsZero() {
		stamp = fmt.Sprintf("\"%s\"", t.Time.Format("2006-01-02"))
	}
	return []byte(stamp), nil
}

//UnmarshalJSON JSON UnmarshalJSON
func (t *Date) UnmarshalJSON(b []byte) error {
	if t != nil {
		tt := Time(*t)
		if err := tt.UnmarshalJSON(b); err == nil {
			*t = Date(tt)
		}
	}
	return nil
}

//Value support  sql.Driver interface
func (t Date) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

//Scan support scan
func (t *Date) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Date{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to Date", v)
}

//String
func (t *Date) String() string {
	return t.Time.Format("2006-01-02")
}
