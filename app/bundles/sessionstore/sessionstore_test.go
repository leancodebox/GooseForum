package sessionstore

import "testing"

func TestSessionCookieSecure(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		appEnv    string
		want      bool
	}{
		{name: "local http", serverURL: "http://localhost:5234", appEnv: "local", want: false},
		{name: "production https", serverURL: "https://example.com", appEnv: "production", want: true},
		{name: "production http", serverURL: "http://example.com", appEnv: "production", want: false},
		{name: "production missing url", serverURL: "", appEnv: "production", want: true},
		{name: "local missing url", serverURL: "", appEnv: "local", want: false},
		{name: "trim and case", serverURL: " HTTPS://example.com ", appEnv: "production", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sessionCookieSecure(tt.serverURL, tt.appEnv); got != tt.want {
				t.Fatalf("sessionCookieSecure(%q, %q) = %v, want %v", tt.serverURL, tt.appEnv, got, tt.want)
			}
		})
	}
}
