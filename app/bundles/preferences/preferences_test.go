package preferences

import (
	"fmt"
	"testing"
)

func TestSpliceConfig(t *testing.T) {
	d := GetIntSlice("db.list")
	fmt.Println(d)
}
