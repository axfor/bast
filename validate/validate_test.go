package validate

import "testing"

type VV struct {
	A string  `json:"a" v:"required|min:1"`
	B *string `json:"b" v:"required|min:1"`
}

func Test_V(t *testing.T) {
	g := "h"
	v := VV{A: "", B: &g}
	v2 := []VV{{A: "111111name", B: &g}}
	v3 := map[string]VV{"aa": VV{A: "111111name", B: &g}}

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
	if err != nil {
		t.Error(err)
	}
}
