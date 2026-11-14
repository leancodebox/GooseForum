package timeopt

import (
	"testing"
	"time"
)

func TestStr2Time(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	got := Str2Time(now.Format(time.DateTime))
	if got.Format(time.DateTime) != now.Format(time.DateTime) {
		t.Fatalf("Str2Time() = %v, want date-time %s", got, now.Format(time.DateTime))
	}
}

func TestStr2TimeCompatibilityLayouts(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "date only", input: "2026-10-03", want: "2026-10-03"},
		{name: "rfc3339", input: "2026-10-03T04:05:06Z", want: "2026-10-03T04:05:06Z"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Str2TimeE(tt.input)
			if err != nil {
				t.Fatalf("Str2TimeE(%q) error = %v", tt.input, err)
			}
			if got.Format(time.RFC3339) != tt.want && got.Format(time.DateOnly) != tt.want {
				t.Fatalf("Str2TimeE(%q) = %s, want %s", tt.input, got.Format(time.RFC3339), tt.want)
			}
		})
	}
	if got, err := Str2TimeE("not a time"); err == nil || !got.IsZero() {
		t.Fatalf("Str2TimeE invalid input = (%v, %v), want zero time and error", got, err)
	}
}
