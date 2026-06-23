package eventhandlers

import (
	"testing"
)

func TestCommentURL(t *testing.T) {
	if got := commentURL(123, 456); got != "/p/post/123#reply-456" {
		t.Fatalf("commentURL() = %q", got)
	}
	if got := commentURL(123, 0); got != "/p/post/123" {
		t.Fatalf("commentURL without comment = %q", got)
	}
	if got := commentURL(0, 456); got != "" {
		t.Fatalf("commentURL without article = %q", got)
	}
}
