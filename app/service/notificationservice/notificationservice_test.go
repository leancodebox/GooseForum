package notificationservice

import "testing"

func TestNormalizePageSize(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "default for zero", in: 0, want: DefaultNotificationPageSize},
		{name: "default for negative", in: -1, want: DefaultNotificationPageSize},
		{name: "keeps valid size", in: 12, want: 12},
		{name: "caps max size", in: MaxNotificationPageSize + 1, want: MaxNotificationPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePageSize(tt.in); got != tt.want {
				t.Fatalf("normalizePageSize(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
