package unicode

import "testing"

func TestDecode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "unicode escape", input: `hello \u4e16\u754c`, want: "hello 世界"},
		{name: "plain text", input: "hello", want: "hello"},
		{name: "invalid escape", input: `\u12zz`, want: `\u12zz`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Decode(tt.input); got != tt.want {
				t.Fatalf("Decode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
