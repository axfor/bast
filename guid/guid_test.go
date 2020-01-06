//Copyright 2018 The axx Authors. All rights reserved.

package guid

import "testing"

func TestGuid(t *testing.T) {
	s := GUID()
	if s == "" {
		t.Fail()
	}
}
