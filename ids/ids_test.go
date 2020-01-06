//Copyright 2018 The axx Authors. All rights reserved.
package ids

import (
	"testing"
)

func Benchmark_ID(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		ID()
	}

}

func Benchmark_Parallel_ID(t *testing.B) {
	t.ReportAllocs()
	// t.ResetTimer()
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ID()
		}
	})
}
