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

func TestEnumIdAndBuildOptions(t *testing.T) {
	if got := SiteManager.Id(); got != uint64(SiteManager) {
		t.Fatalf("Id() = %d, want %d", got, SiteManager)
	}

	options := BuildOptions("en")
	if len(options) != int(SiteManager-Admin+1) {
		t.Fatalf("BuildOptions() length = %d, want %d", len(options), SiteManager-Admin+1)
	}
	for _, option := range options {
		if option.Name == "" || option.Label == "" {
			t.Fatalf("BuildOptions() contains empty label: %#v", option)
		}
		if option.Name != option.Value.LocalizedName("en") || option.Label != option.Value.LocalizedName("en") {
			t.Fatalf("BuildOptions() mismatch: %#v", option)
		}
	}
	if options[0].Label != "Administrator" {
		t.Fatalf("BuildOptions(\"en\")[0].Label = %q, want Administrator", options[0].Label)
	}
}
