package ids

import "testing"

func QBenchmarkIDX(b *testing.B) {

	//for i := 0; i < b.N; i++ {
	//	IDX()
	b.RunParallel(func(pb *testing.PB) {
		IDX()
		for pb.Next() {
			IDX()
		}
	})
	//}
}

func BenchmarkID(b *testing.B) {

	//for i := 0; i < b.N; i++ {
	//	ID()
	b.RunParallel(func(pb *testing.PB) {
		ID()
		for pb.Next() {
			ID()
		}
	})

	//}
}
