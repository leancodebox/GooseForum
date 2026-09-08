package redirectopt

import "testing"

func TestLocal(t *testing.T) {
	for _, unsafe := range []string{"", "https://evil.example", "//evil.example", "/\\evil.example", "javascript:alert(1)", "/path\r\nLocation: https://evil.example"} {
		if got := Local(unsafe); got != "" {
			t.Errorf("Local(%q) = %q", unsafe, got)
		}
	}
	for _, safe := range []string{"/", "/oauth2/authorize?client_id=test", "/settings?tab=binding#profile"} {
		if got := Local(safe); got != safe {
			t.Errorf("Local(%q) = %q", safe, got)
		}
	}
}
