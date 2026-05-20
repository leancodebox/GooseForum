package mailservice

import "testing"

func TestBuildEmailActionURL(t *testing.T) {
	got := buildEmailActionURL("https://example.com/", "/activate", "a/b+c=")
	want := "https://example.com/activate?token=a%2Fb%2Bc%3D"
	if got != want {
		t.Fatalf("buildEmailActionURL() = %q, want %q", got, want)
	}
}
