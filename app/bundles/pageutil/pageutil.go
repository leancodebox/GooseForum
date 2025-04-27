package pageutil

import "cmp"

var (
	minPageSize int = 10
	maxPageSize int = 30
)

func BoundPageSize(pageSize int) int {
	return BoundPageSizeWithRange(pageSize, minPageSize, maxPageSize)
}

func BoundPageSizeWithRange[T cmp.Ordered](pageSize T, minN, maxN T) T {
	return min(max(pageSize, minN), maxN)
}
