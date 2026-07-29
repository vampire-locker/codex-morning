package codexcfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureProjectTrustedCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	workdir := "/Users/me/project"

	changed, err := EnsureProjectTrustedInFile(configPath, workdir)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected config to change")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	wantHeader := `[projects."/Users/me/project"]`
	if !strings.Contains(got, wantHeader) {
		t.Fatalf("missing header in %q", got)
	}
	if !strings.Contains(got, `trust_level = "trusted"`) {
		t.Fatalf("missing trust_level in %q", got)
	}
}

func TestEnsureProjectTrustedIdempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	workdir := "/Users/me/project"

	if _, err := EnsureProjectTrustedInFile(configPath, workdir); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureProjectTrustedInFile(configPath, workdir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected second call to be a no-op")
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("config changed unexpectedly:\n before=%q\n after=%q", before, after)
	}
}

func TestEnsureProjectTrustedUpdatesUntrusted(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	initial := "model = \"gpt\"\n\n[projects.\"/Users/me/project\"]\ntrust_level = \"untrusted\"\n\n[other]\nx = 1\n"
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureProjectTrustedInFile(configPath, "/Users/me/project")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected untrusted entry to be updated")
	}

	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, `trust_level = "trusted"`) {
		t.Fatalf("expected trusted, got %q", text)
	}
	if strings.Contains(text, `trust_level = "untrusted"`) {
		t.Fatalf("untrusted still present: %q", text)
	}
	if !strings.Contains(text, "model = \"gpt\"") || !strings.Contains(text, "[other]") {
		t.Fatalf("unrelated config was lost: %q", text)
	}
}

func TestEnsureProjectTrustedAppendsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	initial := "model = \"gpt\"\n"
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureProjectTrustedInFile(configPath, `/Users/me/proj "quoted"`)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected append")
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.HasPrefix(text, "model = \"gpt\"\n") {
		t.Fatalf("prefix lost: %q", text)
	}
	if !strings.Contains(text, `[projects."/Users/me/proj \"quoted\""]`) {
		t.Fatalf("escaped header missing: %q", text)
	}
}

func TestEnsureProjectTrustedInsertsTrustLevelInExistingSection(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	initial := "[projects.\"/tmp/app\"]\n# comment only\n\n[features]\nhooks = true\n"
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureProjectTrustedInFile(configPath, "/tmp/app")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected insert")
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, "[projects.\"/tmp/app\"]\ntrust_level = \"trusted\"\n") {
		t.Fatalf("trust_level not inserted correctly: %q", text)
	}
	if !strings.Contains(text, "[features]") {
		t.Fatalf("later section lost: %q", text)
	}
}

func TestEnsureProjectTrustedUsesCODEXHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEX_HOME", home)
	t.Setenv("HOME", t.TempDir())

	changed, err := EnsureProjectTrusted("/Users/me/from-env")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected change")
	}
	data, err := os.ReadFile(filepath.Join(home, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `[projects."/Users/me/from-env"]`) {
		t.Fatalf("unexpected content: %s", data)
	}
}

func TestTOMLQuotedKey(t *testing.T) {
	got := TOMLQuotedKey(`/Users/me/project "quoted"\path`)
	want := `"/Users/me/project \"quoted\"\\path"`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
