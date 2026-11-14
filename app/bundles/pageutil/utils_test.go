package pageutil

import (
	"testing"
)

func TestBoundPageSize(t *testing.T) {
	if BoundPageSize(100) != maxPageSize {
		t.Error("BoundPageSize error")
	}
	if BoundPageSize(-1) != minPageSize {
		t.Error("BoundPageSize error")
	}
	if BoundPageSize(11) != 11 {
		t.Error("BoundPageSize error")
	}
}

func TestTrimCursorPage(t *testing.T) {
	items, nextCursor, hasNext := TrimCursorPage([]struct{ ID uint64 }{{1}, {2}, {3}}, 2, func(item struct{ ID uint64 }) uint64 { return item.ID })
	if len(items) != 2 || items[1].ID != 2 || nextCursor != 2 || !hasNext {
		t.Fatalf("unexpected trimmed page: items=%v nextCursor=%d hasNext=%v", items, nextCursor, hasNext)
	}

	items, nextCursor, hasNext = TrimCursorPage([]struct{ ID uint64 }{{1}}, 2, func(item struct{ ID uint64 }) uint64 { return item.ID })
	if len(items) != 1 || nextCursor != 0 || hasNext {
		t.Fatalf("unexpected final page: items=%v nextCursor=%d hasNext=%v", items, nextCursor, hasNext)
	}
}

func TestTrimPage(t *testing.T) {
	items, hasNext := TrimPage([]int{1, 2, 3}, 2)
	if len(items) != 2 || !hasNext {
		t.Fatalf("TrimPage() = %#v, %v; want two items and hasNext", items, hasNext)
	}
	items, hasNext = TrimPage([]int{1}, 2)
	if len(items) != 1 || hasNext {
		t.Fatalf("TrimPage() = %#v, %v; want one item and no next page", items, hasNext)
	}
}

func TestNormalizePageSize(t *testing.T) {
	if got := NormalizePageSize(0, 20, 50); got != 20 {
		t.Fatalf("default page size = %d, want 20", got)
	}
	if got := NormalizePageSize(100, 20, 50); got != 50 {
		t.Fatalf("capped page size = %d, want 50", got)
	}
	if got := NormalizePageSize(30, 20, 50); got != 30 {
		t.Fatalf("regular page size = %d, want 30", got)
	}
}

func TestDefaultPageSize(t *testing.T) {
	if got := DefaultPageSize(0, 20); got != 20 {
		t.Fatalf("DefaultPageSize(0, 20) = %d, want 20", got)
	}
	if got := DefaultPageSize(uint64(0), uint64(10)); got != 10 {
		t.Fatalf("DefaultPageSize(uint64(0), uint64(10)) = %d, want 10", got)
	}
	if got := DefaultPageSize(30, 20); got != 30 {
		t.Fatalf("DefaultPageSize(30, 20) = %d, want 30", got)
	}
}

func TestUintCursor(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		want uint64
	}{
		{name: "empty", in: "", want: 0},
		{name: "invalid", in: "abc", want: 0},
		{name: "zero", in: "0", want: 0},
		{name: "positive", in: "42", want: 42},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseUintCursor(tt.in); got != tt.want {
				t.Fatalf("ParseUintCursor(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
	if got := FormatUintCursor(0); got != "" {
		t.Fatalf("FormatUintCursor(0) = %q, want empty", got)
	}
	if got := FormatUintCursor(42); got != "42" {
		t.Fatalf("FormatUintCursor(42) = %q, want 42", got)
	}
}
