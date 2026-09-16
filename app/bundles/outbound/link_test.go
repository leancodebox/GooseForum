package outbound

import "testing"

func TestDestinations(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "data:text/html,hello", "/login", "https://user@evil.test", "https://good.test\\@evil.test", "https://good.test\r\nLocation: evil"} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, raw := range []string{"https://example.com/a?q=x#foo", "//example.com/a", "http://example.com"} {
		if _, err := Parse(raw); err != nil {
			t.Errorf("rejected %q: %v", raw, err)
		}
	}
}

func TestWhitelistUsesExactHost(t *testing.T) {
	for _, host := range []string{"example.com.evil.test", "notexample.com", "sub.example.com"} {
		u, _ := Parse("https://" + host)
		if Allowed(u, []string{"example.com"}) {
			t.Errorf("allowed %s", host)
		}
	}
	u, _ := Parse("https://EXAMPLE.com/path")
	if !Allowed(u, []string{"example.com"}) {
		t.Fatal("exact host rejected")
	}
}

func TestNormalizeDomains(t *testing.T) {
	got, err := NormalizeDomains([]string{" Example.COM ", "", "example.com", "sub.example.com"})
	if err != nil || len(got) != 2 || got[0] != "example.com" {
		t.Fatalf("%v %v", got, err)
	}
	for _, value := range []string{"https://example.com", "*.example.com", "example.com:443", "example.com/path", "good.test@evil.test"} {
		if _, err := NormalizeDomains([]string{value}); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
}
