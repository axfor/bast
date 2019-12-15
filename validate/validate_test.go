package validate

import (
	"net/url"
	"testing"
)

func TestStruct(t *testing.T) {

	type VV struct {
		A string  `json:"a" v:"required|min:1"`
		B *string `json:"b" v:"required|min:1"`
	}

	g := "a"
	v := VV{A: "a", B: &g}
	v2 := []VV{{A: "b", B: &g}}
	v3 := map[string]VV{"c": VV{A: "c", B: &g}}

	vr := Validator{}

	err := vr.Struct(v)
	if err != nil {
		t.Error(err)
	}

	err = vr.Struct(v2)
	if err != nil {
		t.Error(err)
	}

	err = vr.Struct(v3)
	if err != nil {
		t.Error(err)
	}

	err = vr.Struct(map[string]string{"ccc": "ccc"})
	if err == nil {
		t.Error("FAIL")
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
			"eeeee",
		},
		"f": {
			"ff",
		},
	}
	vr := Validator{}
	err := vr.Request(v4, "d=required|int|min:12|max:16", "e=required|min:5")
	if err != nil {
		t.Error(err)
	}
}

type A struct {
	A string `json:"a" v:"min:1"`
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
