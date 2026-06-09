package terminal

import "testing"

func TestShellQuote(t *testing.T) {
	got := ShellQuote("codex，早上好 'quote'")
	want := "'codex，早上好 '\\''quote'\\'''"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildShellCommand(t *testing.T) {
	got := BuildShellCommand("/Users/me/project", "/opt/homebrew/bin/codex", "codex，早上好")
	want := "cd '/Users/me/project' && '/opt/homebrew/bin/codex' 'codex，早上好'"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
