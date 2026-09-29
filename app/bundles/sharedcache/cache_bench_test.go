package sharedcache

import (
	"io"
	"log/slog"
	"math/bits"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
)

func silenceBenchmarkLogs(b *testing.B) {
	b.Helper()
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	b.Cleanup(func() { slog.SetDefault(original) })
}

func BenchmarkCacheParallelHits(b *testing.B) {
	silenceBenchmarkLogs(b)
	for _, backend := range []string{"local", "shared"} {
		b.Run(backend, func(b *testing.B) {
			keys := make([]string, 1024)
			local := localcache.Cache[int]{MaxEntries: 1024}
			shared := Cache[int]{Name: "benchmark-hits", MaxEntries: 1024}
			var loads atomic.Int64
			load := func() (int, error) { loads.Add(1); return 1, nil }
			var get func(string) (int, error)
			if backend == "local" {
				get = func(key string) (int, error) { return local.GetOrLoadE(key, load, time.Hour) }
			} else {
				get = func(key string) (int, error) { return shared.GetOrLoadE(key, load, time.Hour) }
			}
			for i := range keys {
				keys[i] = strconv.Itoa(i)
				if _, err := get(keys[i]); err != nil {
					b.Fatal(err)
				}
			}
			loads.Store(0)
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					value, err := get(keys[i%len(keys)])
					if err != nil || value != 1 {
						b.Errorf("hit = %d, %v", value, err)
					}
					i++
				}
			})
			b.StopTimer()
			if loads.Load() != 0 {
				b.Fatalf("warm hits reloaded %d times", loads.Load())
			}
		})
	}
}

func BenchmarkCacheParallelInvalidation(b *testing.B) {
	silenceBenchmarkLogs(b)
	c := Cache[int]{Name: "benchmark-invalidation", MaxEntries: 1024}
	keys := make([]string, 1024)
	var loads atomic.Int64
	load := func() (int, error) { loads.Add(1); return 1, nil }
	for i := range keys {
		keys[i] = "user:" + strconv.Itoa(i) + ":audience:all"
		c.Set(keys[i], 1, time.Hour)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := keys[i%len(keys)]
			if i%100 == 0 {
				c.Delete(key)
			}
			if value, err := c.GetOrLoadE(key, load, time.Hour); value != 1 || err != nil {
				b.Errorf("load = %d, %v", value, err)
			}
			i++
		}
	})
	b.StopTimer()
	b.ReportMetric(float64(loads.Load())/float64(b.N), "loads/op")
}

func BenchmarkCachePrefixInvalidation(b *testing.B) {
	silenceBenchmarkLogs(b)
	c := Cache[int]{Name: "benchmark-prefix", MaxEntries: 2048}
	for i := range 2048 {
		c.Set("user:"+strconv.Itoa(i)+":audience:all", 1, time.Hour)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		// A nonmatching prefix measures scanning a full cache on every operation.
		c.DeletePrefix("user:missing:")
	}
}

func BenchmarkCacheParallelPrefixWorkload(b *testing.B) {
	silenceBenchmarkLogs(b)
	for _, invalidate := range []bool{false, true} {
		name := "reads-only"
		if invalidate {
			name = "prefix-1-percent"
		}
		b.Run(name, func(b *testing.B) {
			c := Cache[int]{Name: "benchmark-prefix-workload", MaxEntries: 2048}
			keys, prefixes := make([]string, 2048), make([]string, 2048)
			for i := range keys {
				prefixes[i] = "user:" + strconv.Itoa(i) + ":"
				keys[i] = prefixes[i] + "audience:all"
				c.Set(keys[i], 1, time.Hour)
			}
			var loads atomic.Int64
			var latency [64]atomic.Uint64
			load := func() (int, error) { loads.Add(1); return 1, nil }
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					index := i % len(keys)
					if invalidate && i%100 == 0 {
						if err := c.DeletePrefix(prefixes[index]); err != nil {
							b.Error(err)
						}
					}
					started := time.Now()
					value, err := c.GetOrLoadE(keys[index], load, time.Hour)
					elapsed := uint64(time.Since(started).Nanoseconds())
					latency[bits.Len64(elapsed)].Add(1)
					if value != 1 || err != nil {
						b.Errorf("prefix workload load = %d, %v", value, err)
					}
					i++
				}
			})
			b.StopTimer()
			b.ReportMetric(float64(loads.Load())/float64(b.N), "loads/op")
			var cumulative uint64
			for bucket := range latency {
				cumulative += latency[bucket].Load()
				if cumulative >= (uint64(b.N)*95+99)/100 {
					// Report the logarithmic bucket's upper bound, not exact latency.
					b.ReportMetric(float64(uint64(1)<<bucket), "read-p95-upper-ns")
					break
				}
			}
		})
	}
}
