//Copyright 2018 The axx Authors. All rights reserved.

package lang

import "testing"

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
}
