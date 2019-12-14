package validate

import (
	"net/url"
	"testing"
)

type VV struct {
	A string  `json:"a" v:"required|min:1"`
	B *string `json:"b" v:"required|min:1"`
}

func Test_V(t *testing.T) {
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
// 	key=required|int|min:1
// 	key=required|string|min:1
//	key=sometimes|date|required
func Test_Rules(t *testing.T) {
	v4 := url.Values{
		"d": {
			"22",
		},
		"e": {
			"eeeee",
		},
		"f": {
			"ff",
		},
	}
	vr := Validator{}
	err := vr.Request(v4, "d=required|int|min:12", "e=required|min:5")
	if err != nil {
		t.Error(err)
	}
}
