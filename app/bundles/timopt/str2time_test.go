package timeopt

import (
	"fmt"
	"testing"
	"time"
)

func TestStr2Time(t *testing.T) {
	fmt.Println(Str2Time(time.Now().Format(time.DateTime)))
}
