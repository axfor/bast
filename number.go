package bast

import (
	"strconv"
	"strings"
)

//Int json 序列化是 string 到 int
type Int int

//UnmarshalJSON 反序列化方法
func (c *Int) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t int
		if b[0] == '"' {
			t, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""))
		} else {
			t, err = strconv.Atoi(string(b))
		}
		if err == nil {
			*c = Int(t)
		}
	}
	return err
}

//Int64 json 序列化是 string 到 Int64
type Int64 int64

//UnmarshalJSON 反序列化方法
func (c *Int64) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t int64
		if b[0] == '"' {
			t, err = strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""), 10, 64)
		} else {
			t, err = strconv.ParseInt(string(b), 10, 64)
		}
		if err == nil {
			*c = Int64(t)
		}
	}
	return err
}

//Float32 json 序列化是 string 到 float32
type Float32 float32

//UnmarshalJSON 反序列化方法
func (c *Float32) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t float64
		if b[0] == '"' {
			t, err = strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""), 10)
		} else {
			t, err = strconv.ParseFloat(string(b), 10)
		}
		if err == nil {
			*c = Float32(t)
		}
	}
	return err
}

//Float64 json 序列化是 string 到 float64
type Float64 float64

//UnmarshalJSON 反序列化方法
func (c *Float64) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t float64
		if b[0] == '"' {
			t, err = strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""), 10)
		} else {
			t, err = strconv.ParseFloat(string(b), 10)
		}
		if err == nil {
			*c = Float64(t)
		}
	}
	return err
}
