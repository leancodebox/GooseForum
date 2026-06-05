package lineopt

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/fileopt"
)

func TestReadLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	if err := fileopt.PutContents(path, "first\nsecond\nthird\n"); err != nil {
		t.Fatalf("write test file failed: %v", err)
	}

	var lines []string
	if err := ReadLine(path, func(item string) {
		lines = append(lines, item)
	}); err != nil {
		t.Fatalf("ReadLine failed: %v", err)
	}

	want := []string{"first", "second", "third"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("lines = %#v, want %#v", lines, want)
	}
}

func TestReadLineMissingFile(t *testing.T) {
	if err := ReadLine(filepath.Join(t.TempDir(), "missing.txt"), func(string) {}); err == nil {
		t.Fatalf("expected missing file error")
	}
}
