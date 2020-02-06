package kv

import "testing"

func Test_Any(t *testing.T) {
	v := New()
	v["name"] = "111"
	v["age"] = 18
	v["sex"] = "girl"
	v["node"] = "bast"
	s := v.URL()
	if s != "age=18&name=111&node=bast&sex=girl" {
		t.Fail()
	}
	s2 := v.Join("&")
	if s != s2 {
		t.Fail()
	}

	t.Log(s2)
}

func Test_String(t *testing.T) {
	v := NewString()
	v["name"] = "111"
	v["age"] = "18"
	v["sex"] = "girl"
	v["node"] = "bast"
	s := v.URL()
	if s != "age=18&name=111&node=bast&sex=girl" {
		t.Fail()
	}
	s2 := v.Join("&")
	if s != s2 {
		t.Fail()
	}
}
