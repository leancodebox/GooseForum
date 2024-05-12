package permission

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"testing"
)

func TestEnum(t *testing.T) {
	var l []datastruct.Option[string, Enum]
	for i := Admin; i.Name() != ""; i++ {
		l = append(l, datastruct.Option[string, Enum]{Name: i.Name(), Value: i})
	}
	fmt.Println(l)
}
