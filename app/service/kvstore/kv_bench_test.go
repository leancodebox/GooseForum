package kvstore

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func useBenchmarkStore(b *testing.B) {
	b.Helper()
	resetForTest()
	preferences.Set("badger.path", b.TempDir())
	b.Cleanup(Close)
}

func BenchmarkUpdateBytes(b *testing.B) {
	useBenchmarkStore(b)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		err := UpdateBytes("mutate", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
			return UpdateSet, []byte("benchmark-value"), nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBytesParallelSameKey(b *testing.B) {
	useBenchmarkStore(b)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := UpdateBytes("parallel", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
				return UpdateSet, []byte("benchmark-value"), nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
