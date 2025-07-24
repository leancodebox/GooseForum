package datacache

import (
	"fmt"
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

	fmt.Println(a)
}
