package validate

import "testing"

type VV struct {
	A string  `json:"a" v:"required|min:1"`
	B *string `json:"b" v:"required|min:1"`
}

func Test_V(t *testing.T) {
	g := "h"
	v := VV{A: "111111name", B: &g}
	vr := Validator{}
	err := vr.Struct(v)
	if err != nil {
		t.Error(err)
	}
}
