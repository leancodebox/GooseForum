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
