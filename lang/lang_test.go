//Copyright 2018 The axx Authors. All rights reserved.

package lang

import (
	"fmt"
	"testing"
)

func Test_en(t *testing.T) {
	lang := "en"
	s := Trans(lang, "v.required", "name")
	if s != "The name field is required" {
		t.Fail()
	}
}

func Test_zh_cn(t *testing.T) {
	lang := "zh_cn"
	s := Trans(lang, "v.required", "name")
	if s != "name不能空" {
		t.Fail()
	}
	s = Trans("en", "v.max.string", "1111111", "222222")
	if s != "The 1111111 must be less than 222222 characters" {
		t.Fail()
	}
}

func Test_fmt(t *testing.T) {
	ts := map[string]string{
		"en.v.required":   "The %s field is required",
		"en.v.date":       "The %s is not a valid date",
		"en.v.int":        "The %s must be an integer",
		"en.v.max.string": "The %s must be less than %s characters",
		"en.v.max.int":    "The %s must be less than %s",
		"en.v.min.string": "The %s must must be greater than %s characters",
		"en.v.min.int":    "The %s must must be greater than %s",
		"en.v.email":      "The %s must be a valid email address",
		"en.v.ip":         "The %s must be a valid ip address",
		"en.v.match":      "The %s is a invalid format",
	}
	s := fmt.Sprintf(ts["en.v.max.string"], "1111111", "222222")
	if s != "The 1111111 must be less than 222222 characters" {
		t.Fail()
	}
}

func Benchmark_Optimization(t *testing.B) {
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := Trans("en", "v.max.string", "1111111", "222222")
			if s != "The 1111111 must be less than 222222 characters" {
				t.Fail()
			}
		}
	})
}

func Benchmark_fmt(t *testing.B) {
	ts := map[string]string{
		"en.v.required":   "The %s field is required",
		"en.v.date":       "The %s is not a valid date",
		"en.v.int":        "The %s must be an integer",
		"en.v.max.string": "The %s must be less than %s characters",
		"en.v.max.int":    "The %s must be less than %s",
		"en.v.min.string": "The %s must must be greater than %s characters",
		"en.v.min.int":    "The %s must must be greater than %s",
		"en.v.email":      "The %s must be a valid email address",
		"en.v.ip":         "The %s must be a valid ip address",
		"en.v.match":      "The %s is a invalid format",
	}
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := fmt.Sprintf(ts["en.v.max.string"], "1111111", "222222")
			if s != "The 1111111 must be less than 222222 characters" {
				t.Fail()
			}
		}
	})
}

func Test_file(t *testing.T) {
	err := Dir("./lang")
	if err != nil {
		t.Error(err)
	}
	s := Trans("en", "hi")
	if s != "hi" {
		t.Fail()
	}

	s = Trans("zh_cn", "hi")
	if s != "你好" {
		t.Fail()
	}
}
