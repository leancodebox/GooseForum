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
