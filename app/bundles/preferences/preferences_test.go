package preferences

import "testing"

func TestSpliceConfig(t *testing.T) {
	if got := GetIntSlice("missing.list"); len(got) != 0 {
		t.Fatalf("GetIntSlice() = %v, want empty slice for missing key", got)
	}
}

func TestIsTestMode(t *testing.T) {
	if !IsTestMode() {
		t.Fatal("IsTestMode() = false, want true under go test")
	}
}

func TestTypedPreferences(t *testing.T) {
	Set("test.string", "goose")
	Set("test.int", 12)
	Set("test.float", 1.5)
	Set("test.int64", int64(64))
	Set("test.uint", uint(7))
	Set("test.bool", true)
	Set("test.strings", []string{"a", "b"})
	Set("test.ints", []int{1, 2})
	Set("test.map", map[string]string{"k": "v"})

	if got := Get("test.string"); got != "goose" {
		t.Fatalf("Get = %q", got)
	}
	if got := GetString("missing.string", "fallback"); got != "fallback" {
		t.Fatalf("GetString default = %q", got)
	}
	if got := GetInt("test.int"); got != 12 {
		t.Fatalf("GetInt = %d", got)
	}
	if got := GetFloat64("test.float"); got != 1.5 {
		t.Fatalf("GetFloat64 = %f", got)
	}
	if got := GetInt64("test.int64"); got != 64 {
		t.Fatalf("GetInt64 = %d", got)
	}
	if got := GetUint("test.uint"); got != 7 {
		t.Fatalf("GetUint = %d", got)
	}
	if got := GetBool("test.bool"); !got {
		t.Fatalf("GetBool = false")
	}
	if got := GetStringSlice("test.strings"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("GetStringSlice = %#v", got)
	}
	if got := GetIntSlice("test.ints"); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("GetIntSlice = %#v", got)
	}
	if got := GetStringMapString("test.map"); got["k"] != "v" {
		t.Fatalf("GetStringMapString = %#v", got)
	}
	if !IsSet("test.string") {
		t.Fatalf("IsSet should see configured value")
	}
	if all := All(); len(all) == 0 {
		t.Fatalf("All should include loaded settings")
	}
}

func TestExclusivePreferences(t *testing.T) {
	Set("scoped.name", "goose")
	Set("scoped.count", 3)
	Set("scoped.ratio", 2.5)
	Set("scoped.big", int64(99))
	Set("scoped.flag", true)
	Set("scoped.map", map[string]string{"a": "b"})

	prefs := GetExclusivePreferences("scoped")
	if got := prefs.Get("name"); got != "goose" {
		t.Fatalf("Get = %q", got)
	}
	if got := prefs.GetString("missing", "fallback"); got != "fallback" {
		t.Fatalf("GetString default = %q", got)
	}
	if got := prefs.GetInt("count"); got != 3 {
		t.Fatalf("GetInt = %d", got)
	}
	if got := prefs.GetFloat64("ratio"); got != 2.5 {
		t.Fatalf("GetFloat64 = %f", got)
	}
	if got := prefs.GetInt64("big"); got != 99 {
		t.Fatalf("GetInt64 = %d", got)
	}
	if got := prefs.GetUint("count"); got != 3 {
		t.Fatalf("GetUint = %d", got)
	}
	if got := prefs.GetBool("flag"); !got {
		t.Fatalf("GetBool = false")
	}
	if got := prefs.GetStringMapString("map"); got["a"] != "b" {
		t.Fatalf("GetStringMapString = %#v", got)
	}
}
