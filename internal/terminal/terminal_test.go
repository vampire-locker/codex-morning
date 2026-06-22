package terminal

import (
	"os"
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	got := ShellQuote("codex，早上好 'quote'")
	want := "'codex，早上好 '\\''quote'\\'''"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildShellCommand(t *testing.T) {
	got := BuildShellCommand("/Users/me/project", "/opt/homebrew/bin/codex", "codex，早上好")
	want := `cd '/Users/me/project' && '/opt/homebrew/bin/codex' -c 'projects."/Users/me/project".trust_level="trusted"' 'codex，早上好'`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTOMLQuotedKey(t *testing.T) {
	got := TOMLQuotedKey(`/Users/me/project "quoted"\path`)
	want := `"/Users/me/project \"quoted\"\\path"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWriteCommandFile(t *testing.T) {
	path, err := WriteCommandFile("/Users/me/project", "/opt/homebrew/bin/codex", "codex，早上好")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Fatalf("got mode %v, want 0700", info.Mode().Perm())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	required := []string{
		"#!/bin/zsh",
		`rm -f "$0"`,
		`cd '/Users/me/project' && '/opt/homebrew/bin/codex' -c 'projects."/Users/me/project".trust_level="trusted"' 'codex，早上好'`,
	}
	for _, value := range required {
		if !strings.Contains(string(content), value) {
			t.Fatalf("command file missing %q:\n%s", value, content)
		}
	}
}
