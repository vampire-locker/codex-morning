package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWorkdirReturnsAbsoluteDirectory(t *testing.T) {
	dir := t.TempDir()

	got, err := resolveWorkdir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("got %q, want %q", got, dir)
	}
}

func TestResolveWorkdirRejectsMissingDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	_, err := resolveWorkdir(missing)
	if err == nil {
		t.Fatal("expected missing directory to fail")
	}
	if !strings.Contains(err.Error(), "工作目录不存在") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveWorkdirRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := resolveWorkdir(path)
	if err == nil {
		t.Fatal("expected file path to fail")
	}
	if !strings.Contains(err.Error(), "工作目录不是目录") {
		t.Fatalf("unexpected error: %v", err)
	}
}
