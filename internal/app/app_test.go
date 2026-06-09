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
	setChineseLocale(t)

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
	setChineseLocale(t)

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

func TestDefaultPromptFollowsLocale(t *testing.T) {
	t.Setenv("LC_ALL", "zh_CN.UTF-8")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	if got := DefaultPrompt(); got != "codex，早上好" {
		t.Fatalf("got %q", got)
	}

	t.Setenv("LC_ALL", "en_US.UTF-8")
	if got := DefaultPrompt(); got != "codex, good morning" {
		t.Fatalf("got %q", got)
	}
}

func setChineseLocale(t *testing.T) {
	t.Helper()

	t.Setenv("LC_ALL", "zh_CN.UTF-8")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
}
