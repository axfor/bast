//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"strconv"
)

//Int is auto handle string to int
type Int int

//UnmarshalJSON JSON UnmarshalJSON
func (c *Int) UnmarshalJSON(b []byte) error {
	var err error
	if b != nil && len(b) > 0 {
		var t int
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err = strconv.Atoi(s)
			if err == nil {
				*c = Int(t)
			}
		}
	}
	return err
}

//Int64 is auto handle string to int64
type Int64 int64

//UnmarshalJSON JSON  UnmarshalJSON
func (c *Int64) UnmarshalJSON(b []byte) error {
	if b != nil && len(b) > 0 {
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			*c = Int64(t)
		}
	}
	return nil
}

//UInt is auto handle string to uint
type UInt uint

//UnmarshalJSON JSON UnmarshalJSON
func (c *UInt) UnmarshalJSON(b []byte) error {
	if b != nil && len(b) > 0 {
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err := strconv.ParseUint(s, 10, 0)
			if err != nil {
				return err
			}
			*c = UInt(t)
		}
	}
	return nil
}

//UInt64 is auto handle string to uint
type UInt64 uint64

//UnmarshalJSON JSON UnmarshalJSON
func (c *UInt64) UnmarshalJSON(b []byte) error {
	if b != nil && len(b) > 0 {
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err := strconv.ParseUint(s, 10, 0)
			if err != nil {
				return err
			}
			*c = UInt64(t)
		}
	}
	return nil
}

//Float32 is auto handle string to float32
type Float32 float32

//UnmarshalJSON JSON UnmarshalJSON
func (c *Float32) UnmarshalJSON(b []byte) error {
	if b != nil && len(b) > 0 {
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return err
			}
			*c = Float32(t)
		}
	}
	return nil
}

//Float64 is auto handle string to float64
type Float64 float64

//UnmarshalJSON JSON UnmarshalJSON
func (c *Float64) UnmarshalJSON(b []byte) error {
	if b != nil && len(b) > 0 {
		s := string(b)
		if s != "" && s != "\"\"" {
			if b[0] == '"' {
				s = s[1 : len(s)-1]
			}
			t, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			*c = Float64(t)
		}
	}
	return nil
}
