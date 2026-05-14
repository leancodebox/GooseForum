package randopt

import (
	"sync"
	"testing"
)

func BenchmarkB(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = RandomString(10)
	}
}

func TestRand(t *testing.T) {
	str := RandomString(10)
	if len(str) != 10 {
		t.Fatalf("RandomString length = %d", len(str))
	}
}

type CounterOld struct {
	lock   sync.Mutex
	number int64
}

func (itself *CounterOld) Add() int64 {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	itself.number += 1
	return itself.number
}

func (itself *CounterOld) Get() int64 {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	return itself.number
}

func TestIdMakerP(t *testing.T) {
	counter := Counter{}
	wg := sync.WaitGroup{}
	const goroutines = 10
	const increments = 10000
	for i := 1; i <= goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for start := 1; start <= increments; start++ {
				counter.Add()
			}
		}()
	}
	wg.Wait()
	if got, want := counter.Get(), int64(goroutines*increments); got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}
}
