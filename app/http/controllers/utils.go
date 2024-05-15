package controllers

import "cmp"

var (
	minPageSize int = 10
	maxPageSize int = 30
)

func boundPageSize(pageSize int) int {
	return boundPageSizeWithRange(pageSize, minPageSize, maxPageSize)
}

func boundPageSizeWithRange[T cmp.Ordered](pageSize T, minN, maxN T) T {
	return min(max(pageSize, minN), maxN)
}
