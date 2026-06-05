package validate

import (
	"strings"
	"testing"
)

func TestValid(t *testing.T) {
	type payload struct {
		Name string `validate:"required"`
	}

	if err := Valid(payload{Name: "goose"}); err != nil {
		t.Fatalf("Valid returned unexpected error: %v", err)
	}

	err := Valid(payload{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if got := FormatError(err); !strings.Contains(got, "Name") {
		t.Fatalf("formatted error = %q, want field name", got)
	}
}
