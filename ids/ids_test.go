//Copyright 2018 The axx Authors. All rights reserved.
package ids

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func Test_ID(t *testing.T) {
	numProcs := 1 * runtime.GOMAXPROCS(0)
	start := time.Now().UnixNano() / 1000000
	rn := 20000000 //try 2 km
	nn := rn
	rn = rn / numProcs
	cds := make(chan int64, nn)
	var wg sync.WaitGroup
	wg.Add(numProcs)
	for index := 0; index < numProcs; index++ {
		go func() {
			defer wg.Done()
			for j := 0; j < rn; j++ {
				id := ID()
				if id <= start {
					t.Errorf("error=%d ", id)
					break
				}
				cds <- id
			}
		}()
	}
	wg.Wait()
	lg := len(cds)
	ds := make(map[int64]struct{}, nn)
	for index := 0; index < lg; index++ {
		id := <-cds
		_, ok := ds[id]
		if ok {
			t.Errorf("exist=%d,%d", id, len(ds))
			break
		}
		ds[id] = struct{}{}
	}
	// t.Logf("finish-e=%d,%d", len(ds), lg)

}
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
