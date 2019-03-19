//Copyright 2018 The axx Authors. All rights reserved.
package ids

import (
	"testing"
)

func Benchmark_ID(t *testing.B) {
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := ID()
			if id <= 0 {
				t.Errorf("error=%d", id)
				break
			}
		}
	})
}
