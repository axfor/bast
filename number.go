//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"strconv"
	"strings"
)

//Int is auto handle string to int
type Int int

//UnmarshalJSON JSON UnmarshalJSON
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

//UInt is auto handle string to uint
type UInt uint

//UnmarshalJSON JSON UnmarshalJSON
func (c *UInt) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t uint64
		if b[0] == '"' {
			t, err = strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""), 10, 0)
		} else {
			t, err = strconv.ParseUint(string(b), 10, 0)
		}
		if err == nil {
			*c = UInt(t)
		}
	}
	return err
}

//UInt64 is auto handle string to uint
type UInt64 uint64

//UnmarshalJSON JSON UnmarshalJSON
func (c *UInt64) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t uint64
		if b[0] == '"' {
			t, err = strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\""), 10, 0)
		} else {
			t, err = strconv.ParseUint(string(b), 10, 0)
		}
		if err == nil {
			*c = UInt64(t)
		}
	}
	return err
}

//Int64 is auto handle string to int64
type Int64 int64

//UnmarshalJSON JSON  UnmarshalJSON
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

//Float32 is auto handle string to float32
type Float32 float32

//UnmarshalJSON JSON UnmarshalJSON
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

//Float64 is auto handle string to float64
type Float64 float64

//UnmarshalJSON JSON UnmarshalJSON
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
