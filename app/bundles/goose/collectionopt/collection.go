package collectionopt

//type Collection[T any] struct {
//	value []T // 数组
//}
//
//func NewCollection[T any](slice []T) *Collection[T] {
//	return &Collection[T]{value: []T{}}
//}
//
//func (c *Collection[T]) Map[R any](action func(T) R) *Collection[R] {
//	var res []R
//	for _, item := range c.value {
//		res = append(res, Map[T, R](item, action))
//	}
//	return NewCollection(res)
//}
//
//func Map[T, R any](item T, action func(T) R) R {
//	return action(item)
//}

func IsEmpty[T []any | map[any]any](col T) bool {
	return col == nil || len(col) == 0
}

func IsNotEmpty[T []any | map[any]any](col T) bool {
	return !IsEmpty(col)
}

func Map[T, R any](list []T, f func(T) R) (result []R) {
	result = make([]R, len(list))
	for index, item := range list {
		result[index] = f(item)
	}
	return
}

func ArrayMap[T, R any](f func(T) R, list []T) []R {
	return Map[T, R](list, f)
}

func Filter[T any](list []T, f func(T) bool) (result []T) {
	for _, item := range list {
		if f(item) {
			result = append(result, item)
		}
	}
	return
}

func ArrayFill[T any](startIndex int, num uint, value T) map[int]T {
	result := make(map[int]T, num)
	var i uint
	for i = 0; i < num; i++ {
		result[startIndex] = value
		startIndex++
	}
	return result
}

func Map2Slice[K comparable, V any](m map[K]V) []V {
	s := make([]V, len(m))
	i := 0
	for _, v := range m {
		s[i] = v
		i += 1
	}
	return s
}

func Slice2Map[V any, K comparable](s []V, f func(V) K) map[K]V {
	m := make(map[K]V, len(s))
	for _, v := range s {
		m[f(v)] = v
	}
	return m
}

func Map2Map[mK comparable, mV any, newK comparable, newV any](oldMap map[mK]mV, f func(mK, mV) (newK, newV)) map[newK]newV {
	newMap := make(map[newK]newV, len(oldMap))
	for k, v := range oldMap {
		nk, nv := f(k, v)
		newMap[nk] = nv
	}
	return newMap
}
