package collectionopt

import (
	"fmt"
	"testing"
)

func MapOrSlice[K comparable, V, R any](mOrS any, f func(string, V) R, k K) []R {
	switch v := mOrS.(type) {
	case map[K]V:
		res := make([]R, 0, len(v))
		for k, vItem := range v {
			res = append(res, f(fmt.Sprintf("%v", k), vItem))
		}
		return res
	case []V:
		res := make([]R, 0, len(v))
		for k, vItem := range v {
			res = append(res, f(fmt.Sprintf("%v", k), vItem))
		}
		return res
	default:
		panic("unsupported collection type")
	}
}

func TestMOS(t *testing.T) {
	m := map[int]string{1: "2", 2: "2"}
	d := MapOrSlice(m, func(s string, v string) int {
		return 1
	}, 1)
	fmt.Println(d)
}

func TestArrayMap(t *testing.T) {
	a := []int{1, 2, 3, 4}
	r := ArrayMap(func(item int) string {
		return fmt.Sprintf("%v", item)
	}, a)
	fmt.Println(r)
}

func TestArrayFilter(t *testing.T) {
	a := []int{1, 2, 3, 4}
	a = Filter(a, func(item int) bool {
		return item > 2
	})
	fmt.Println(a)
}
