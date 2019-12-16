package validate

import (
	"net/url"
	"testing"
)

func TestStruct(t *testing.T) {

	type vv struct {
		A string  `json:"a" v:"required|min:1"`
		B *string `json:"b" v:"required|min:1"`
	}

	g := "a"
	v := vv{A: "a", B: &g}

	vr := Validator{}

	err := vr.Struct(v)
	if err != nil {
		t.Error(err)
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

//
// 	key1=required|int|min:1
// 	key2=required|string|min:1|max:12
//	key3=sometimes|date|required
func TestRules(t *testing.T) {
	v4 := url.Values{
		"d": {
			"15",
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
	}
	vr := Validator{}
	err := vr.Request(v4, "d=required|int|min:12|max:16", "e=required|min:5", "t=date", "z=sometimes|required|int")
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
	err := vr.Request(v4, "t=sometimes|required|date")
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
	err := vr.Request(v4, "t=required")
	if err != nil {
		t.Error(err)
	}
	vr2 := Validator{"zh-cn"}
	err = vr2.Request(v4, "t=required")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkFieldSuccess(b *testing.B) {
	vr := Validator{}
	type Foo struct {
		Valuer string `v:"min:1"`
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
