//Copyright 2018 The axx Authors. All rights reserved.
package ids

import "testing"

func Benchmark_ID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ID()
		}
	})
}
