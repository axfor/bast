//Copyright 2018 The axx Authors. All rights reserved.
package ids

import (
	"errors"
	"testing"
)

func Benchmark_ID(t *testing.B) {
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := ID()
			if id <= 0 {
				t.Error(errors.New("error"))
				break
			}
		}
	})
}
