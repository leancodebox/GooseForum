package permission

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/datastruct"
)

func TestEnum(t *testing.T) {
	var l []datastruct.Option[string, Enum]
	for i := Admin; i.Name() != ""; i++ {
		l = append(l, datastruct.Option[string, Enum]{Name: i.Name(), Value: i})
	}
	if len(l) == 0 {
		t.Fatal("expected permission options")
	}
}
