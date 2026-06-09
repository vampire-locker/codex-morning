package terminal

import (
	"fmt"
	"os/exec"
	"strings"
)

func OpenCodex(workdir, codexBin, prompt string) error {
	if workdir == "" {
		return fmt.Errorf("workdir is required")
	}
	if codexBin == "" {
		codexBin = "codex"
	}
	command := BuildShellCommand(workdir, codexBin, prompt)
	script := `
function run(argv) {
  const terminal = Application("Terminal");
  terminal.activate();
  terminal.doScript(argv[0]);
}`
	cmd := exec.Command("/usr/bin/osascript", "-l", "JavaScript", "-e", script, command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("open Terminal: %w\n%s", err, string(out))
	}
	return nil
}

func BuildShellCommand(workdir, codexBin, prompt string) string {
	return "cd " + ShellQuote(workdir) + " && " + ShellQuote(codexBin) + " " + ShellQuote(prompt)
}

func ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
