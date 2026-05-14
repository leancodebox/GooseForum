package datacache

import (
	"testing"
	"time"
)

func TestCache_GetOrLoad(t *testing.T) {
	c := Cache[string]{}
	a, _ := c.GetOrLoadE("", func() (string, error) {
		return "a", nil
	}, time.Minute)
	c.GetOrLoadE("", func() (string, error) {
		return "a", nil
	}, time.Minute)

	if a != "a" {
		t.Fatalf("GetOrLoadE() = %q", a)
	}
}
