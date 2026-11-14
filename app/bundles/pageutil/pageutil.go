package pageutil

import (
	"cmp"
	"strconv"
)

const (
	minPageSize = 10
	maxPageSize = 30
)

type pageSizeNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func BoundPageSize(pageSize int) int {
	return BoundPageSizeWithRange(pageSize, minPageSize, maxPageSize)
}

func BoundPageSizeWithRange[T cmp.Ordered](pageSize T, minN, maxN T) T {
	return min(max(pageSize, minN), maxN)
}

// DefaultPageSize returns defaultSize when pageSize is not positive.
func DefaultPageSize[T pageSizeNumber](pageSize, defaultSize T) T {
	if pageSize <= 0 {
		return defaultSize
	}
	return pageSize
}

// TrimCursorPage trims an over-fetched cursor page and returns its next cursor.
// Callers should query pageSize+1 records before invoking it.
func TrimCursorPage[T any](items []T, pageSize int, cursor func(T) uint64) ([]T, uint64, bool) {
	items, hasNext := TrimPage(items, pageSize)
	nextCursor := uint64(0)
	if hasNext && len(items) > 0 {
		nextCursor = cursor(items[len(items)-1])
	}
	return items, nextCursor, hasNext
}

// TrimPage trims an over-fetched offset page and reports whether another page
// exists. Callers should query pageSize+1 records before invoking it.
func TrimPage[T any](items []T, pageSize int) ([]T, bool) {
	hasNext := len(items) > pageSize
	if hasNext {
		items = items[:pageSize]
	}
	return items, hasNext
}

// NormalizePageSize applies a service-specific default and maximum page size.
func NormalizePageSize(pageSize, defaultSize, maxSize int) int {
	pageSize = DefaultPageSize(pageSize, defaultSize)
	if pageSize > maxSize {
		return maxSize
	}
	return pageSize
}

// ParseUintCursor parses a positive decimal cursor and treats invalid values as empty.
func ParseUintCursor(cursor string) uint64 {
	id, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil || id == 0 {
		return 0
	}
	return id
}

// FormatUintCursor formats a positive cursor, keeping zero as an empty cursor.
func FormatUintCursor(id uint64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(id, 10)
}
