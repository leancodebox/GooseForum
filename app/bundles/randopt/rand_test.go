package randopt

import (
	"fmt"
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
	fmt.Println(str)
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
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for start := 1; start <= 10000000; start++ {
				counter.Add()
			}
		}()
	}
	wg.Wait()
	fmt.Println(counter.Get())
}
