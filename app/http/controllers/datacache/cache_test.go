package datacache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_GetOrLoad(t *testing.T) {
	c := Cache[string, string]{}
	a, _ := c.GetOrLoad("", func() (string, error) {
		return "a", nil
	}, time.Minute)
	c.GetOrLoad("", func() (string, error) {
		return "a", nil
	}, time.Minute)

	fmt.Println(a)
}
