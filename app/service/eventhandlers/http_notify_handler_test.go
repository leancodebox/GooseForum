package eventhandlers

import (
	"testing"
)

func TestPostURL(t *testing.T) {
	if got := postURL(123, 456); got != "/p/post/123#post-456" {
		t.Fatalf("postURL() = %q", got)
	}
	if got := postURL(123, 0); got != "/p/post/123" {
		t.Fatalf("postURL without post = %q", got)
	}
	if got := postURL(0, 456); got != "" {
		t.Fatalf("postURL without topic = %q", got)
	}
}
