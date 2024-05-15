package controllers

import (
	"testing"
)

func TestBoundPageSize(t *testing.T) {
	if boundPageSize(100) != maxPageSize {
		t.Error("boundPageSize error")
	}
	if boundPageSize(-1) != minPageSize {
		t.Error("boundPageSize error")
	}
	if boundPageSize(11) != 11 {
		t.Error("boundPageSize error")
	}
}
