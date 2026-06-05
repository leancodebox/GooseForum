package fileopt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileContentsLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.txt")

	if IsExist(path) {
		t.Fatalf("new temp file should not exist")
	}

	if err := PutContents(path, "hello"); err != nil {
		t.Fatalf("PutContents failed: %v", err)
	}
	if !IsExist(path) {
		t.Fatalf("file should exist after PutContents")
	}

	content, err := GetContents(path)
	if err != nil {
		t.Fatalf("GetContents failed: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("content = %q, want hello", string(content))
	}

	if err := FilePutContents(path, []byte(" world"), true); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	content, err = FileGetContents(path)
	if err != nil {
		t.Fatalf("FileGetContents failed: %v", err)
	}
	if string(content) != "hello world" {
		t.Fatalf("appended content = %q, want hello world", string(content))
	}
}

func TestIsExistOrCreate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "init.txt")

	if err := IsExistOrCreate(path, "init"); err != nil {
		t.Fatalf("IsExistOrCreate failed: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created file failed: %v", err)
	}
	if string(content) != "init" {
		t.Fatalf("created content = %q, want init", string(content))
	}

	if err := IsExistOrCreate(path, "ignored"); err != nil {
		t.Fatalf("IsExistOrCreate existing file failed: %v", err)
	}
	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing file failed: %v", err)
	}
	if string(content) != "init" {
		t.Fatalf("existing content = %q, want init", string(content))
	}
}

func TestDirExistOrCreate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")

	if err := DirExistOrCreate(dir); err != nil {
		t.Fatalf("DirExistOrCreate failed: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("created dir stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("created path should be directory")
	}
}
