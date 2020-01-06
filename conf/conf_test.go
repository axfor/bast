//Copyright 2018 The axx Authors. All rights reserved.

package conf

import (
	"testing"
)

func TestConf(t *testing.T) {
	SetPath("../config.conf")
	c := Confs()
	if c == nil || len(c) <= 0 {
		t.Fail()
	}

	c1 := Conf()
	if c1 == nil || c1.Key == "" {
		t.Fail()
	}
}
