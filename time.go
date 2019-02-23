//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

//Time yyyy-MM-dd HH:mm:ss 格式日期
type Time struct {
	time.Time
}

// type Time time.Time

//MarshalJSON 序列化方法
func (t *Time) MarshalJSON() ([]byte, error) {
	stamp := "\"\""
	if t != nil && !t.IsZero() {
		stamp = fmt.Sprintf("\"%s\"", t.Time.Format("2006-01-02 15:04:05"))
	}
	return []byte(stamp), nil
}

//UnmarshalJSON 反序列化方法
func (t *Time) UnmarshalJSON(b []byte) error {
	tt, err := byteToTime(b, "")
	if err == nil {
		*t = tt
	}
	return err
}

func strToTime(t string, layout string) (Time, error) {
	return byteToTime([]byte(t), layout)
}

func byteToTime(b []byte, layout string) (Time, error) {
	var t Time
	var err error

	if b != nil && len(b) > 0 {
		s := strings.Trim(string(b), "\"")
		l := len(s)
		var v time.Time
		loc, _ := time.LoadLocation("Local") //重要：获取时区
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
				err = errors.New("bytes to time")
			}
		}
	}
	return t, err
}

//Value 获取值
func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	// var ti = time.Time(t)
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

//T time
func (t *Time) T(v interface{}) *time.Time {
	if !t.IsZero() {
		return &t.Time
	}
	return nil
}

//Scan valueof time.Time
func (t *Time) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Time{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to Time", v)
}

//Now 当前时间
func Now() Time {
	tt := Time{Time: time.Now()}
	return tt
}

//Nows 当前时间-指针
func Nows() *Time {
	tt := Time{Time: time.Now()}
	return &tt
}

//NowPoint 当前时间-指针
func NowPoint() *Time {
	return NowTime()
}

//NowTime 当前时间-指针
func NowTime() *Time {
	tt := Time{Time: time.Now()}
	return &tt
}

//TimeByTime 根据time.Time构建 Time
func TimeByTime(t time.Time) Time {
	tt := Time{Time: t}
	return Time(tt)
}

//TimesByTime 根据time.Time构建 Time-指针
func TimesByTime(t time.Time) *Time {
	tt := TimeByTime(t)
	return &tt
}

//TimeWithString 字符串时间转化为时间
func TimeWithString(t string, layout ...string) (Time, error) {
	l := ""
	if layout != nil {
		l = layout[0]
	}
	return strToTime(t, l)
}

//String
func (t *Time) String() string {
	return t.Time.Format("2006-01-02 15:04:05")
}

//Format yyyy-MM-dd 格式的字符日期
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

//Date yyyy-MM-dd 格式日期
type Date Time

//NowDate 当前日期
func NowDate() Date {
	return Date(Now())
}

//NowDates 当前日期-指针
func NowDates() *Date {
	d := NowDate()
	return &d
}

//DateByTime 根据time.Time构建 Date
func DateByTime(t time.Time) Date {
	tt := Time{Time: t}
	return Date(tt)
}

//DatesByTime 根据time.Time构建 Date-指针
func DatesByTime(t time.Time) *Date {
	tt := DateByTime(t)
	return &tt
}

//MarshalJSON 序列化方法
func (t *Date) MarshalJSON() ([]byte, error) {
	stamp := "\"\""
	if t != nil && !t.IsZero() {
		stamp = fmt.Sprintf("\"%s\"", t.Time.Format("2006-01-02"))
	}
	return []byte(stamp), nil
}

//UnmarshalJSON 反序列化方法
//
func (t *Date) UnmarshalJSON(b []byte) error {
	if t != nil {
		tt := Time(*t)
		if err := tt.UnmarshalJSON(b); err == nil {
			*t = Date(tt)
		}
	}
	return nil
}

//Value 获取值
func (t Date) Value() (driver.Value, error) {
	var zeroTime time.Time
	//var ti = time.Time(t)
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

//Scan valueof time.Time
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
