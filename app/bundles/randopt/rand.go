package randopt

import (
	"math/rand"
	"sync/atomic"

	"github.com/samber/lo"
)

func RandTrue(Molecular int, Denominator int) bool {
	return rand.Intn(Denominator) < Molecular
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var lettersLen = len(letters)

func RandomString(length int) string {
	return lo.RandomString(length, []rune(letters))
}

type Counter struct {
	number int64
}

func NewCounter(startId int64) *Counter {
	return &Counter{number: startId}
}

func (itself *Counter) Add() int64 {
	return atomic.AddInt64(&itself.number, 1)
}

func (itself *Counter) Get() int64 {
	return atomic.LoadInt64(&itself.number)
}

func (itself *Counter) Clean() int64 {
	atomic.StoreInt64(&itself.number, 0)
	return 0
}
