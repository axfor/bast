//Copyright 2018 The axx Authors. All rights reserved.

package validate

import (
	"errors"
	"net/url"
	"testing"

	"github.com/aixiaoxiang/bast/lang"
)

type vv struct {
	A string  `json:"a" v:"required|min:1"`
	B *string `json:"b" v:"required|min:1"`
}

//customer verify
func (c vv) Verify(v *Validator) error {
	if c.A == "a" && *c.B != "b" {
		return errors.New("A equals a B must equals b")
	}
	return nil
}

func TestStruct(t *testing.T) {

	g := "a"
	v := vv{A: "a", B: &g}

	vr := Validator{}

	err := vr.Struct(v)
	if err != nil {
		//t.Error(err)
	}

	v2 := []vv{{A: "b", B: &g}}
	err = vr.Struct(v2)
	if err != nil {
		t.Error(err)
	}

	v3 := map[string]vv{"c": vv{A: "c", B: &g}}
	err = vr.Struct(v3)
	if err != nil {
		t.Error(err)
	}

}

//syntax:  key[/key translator][/split divide (default is |)]@verify1[:verify1 param]|verify2
//such as:
// 	key1@required|int|min:1
// 	key2/key2_translator@required|string|min:1|max:12
//	key3@sometimes|date|required
func TestRules(t *testing.T) {
	v4 := url.Values{
		"d": {
			"16",
		},
		"e": {
			"eeeeeee",
		},
		"f": {
			"ff",
		},
		"t": {
			"2019-01-01 12:12",
		},
		"email": {
			"a@a.com",
		},
		"host": {
			"127.0.0.1",
		},
		"bast": {
			"github.aixiaoxiang/bast",
		},
		"bast1": {
			"github.axx/bast",
		},
	}
	vr := Validator{"zh-cn"}
	err := vr.Request(v4,
		"d/d_name@required|int|min:12|max:16",
		"e@required|min:5",
		"t@date",
		"z@sometimes|required|int",
		"email@email",
		`bast//!@match:^([a-zA-Z]+\.[a-zA-Z]*\/[a-zA-Z]*)$`,  //regular expression match will cache the same //new a regular
		`bast1//!@match:^([a-zA-Z]+\.[a-zA-Z]*\/[a-zA-Z]*)$`, //form cache //not new regular
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSometimesRules(t *testing.T) {
	v4 := url.Values{
		"t": {
			"2019-01-01 12:12",
		},
	}

	vr := Validator{}
	err := vr.Request(v4, "t@sometimes|required|date")
	if err != nil {
		t.Error(err)
	}
}

func Test_Lang(t *testing.T) {
	v4 := url.Values{
		"t": {
			"",
		},
	}
	vr := Validator{"en"}
	err := vr.Request(v4, "t@required")
	if err == nil {
		t.Error("The t field is required")
	}

	vr2 := Validator{"zh-cn"}
	err = vr2.Request(v4, "t@required")
	if err == nil {
		t.Error("The t field is required")
	}
}

func BenchmarkFieldSuccess(b *testing.B) {
	vr := Validator{}
	type Foo struct {
		Valuer string `json:"v" v:"min:1"`
	}

	validFoo := &Foo{Valuer: "1"}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = vr.Struct(validFoo)
	}
}

func BenchmarkFieldSuccessParallel(t *testing.B) {
	vr := &Validator{}
	type Foo struct {
		Valuer string `v:"min:1"`
	}

	validFoo := &Foo{Valuer: "1"}

	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = vr.Struct(validFoo)
		}
	})
}

func init() {

	//register translator key
	lang.RegisterKey("zh-cn", "t", "标题")

	//register translator keys
	lang.RegisterKeys("zh-cn", map[string]string{
		"d_name": "地址信息",
		"e":      "名称",
		"d":      "地址",
	})
}
