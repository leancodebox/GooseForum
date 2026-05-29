package appcache

import "testing"

func TestHardMaxCacheSizeMBFromMemory(t *testing.T) {
	tests := []struct {
		name       string
		totalBytes uint64
		ok         bool
		want       int
	}{
		{name: "fallback", ok: false, want: fallbackHardMaxCacheMB},
		{name: "minimum", totalBytes: 512 * bytesPerMiB, ok: true, want: minHardMaxCacheMB},
		{name: "calculated", totalBytes: 4 * 1024 * bytesPerMiB, ok: true, want: 40},
		{name: "maximum", totalBytes: 64 * 1024 * bytesPerMiB, ok: true, want: maxHardMaxCacheMB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hardMaxCacheSizeMBFromMemory(tt.totalBytes, tt.ok); got != tt.want {
				t.Fatalf("hardMaxCacheSizeMBFromMemory() = %d, want %d", got, tt.want)
			}
		})
	}
}
