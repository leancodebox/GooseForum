package preferences

import "testing"

func TestSpliceConfig(t *testing.T) {
	if got := GetIntSlice("missing.list"); len(got) != 0 {
		t.Fatalf("GetIntSlice() = %v, want empty slice for missing key", got)
	}
}
