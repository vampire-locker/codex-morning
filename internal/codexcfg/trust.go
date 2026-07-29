package codexcfg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vampire-locker/codex-morning/internal/i18n"
)

// EnsureProjectTrusted marks workdir as a trusted Codex project in the
// user config (~/.codex/config.toml, or $CODEX_HOME/config.toml).
//
// Current Codex interactive TUI still prompts for directory trust unless the
// project is persisted in config.toml; a one-off -c override is not enough.
// Returns true when the config file was modified.
func EnsureProjectTrusted(workdir string) (bool, error) {
	if workdir == "" {
		return false, fmt.Errorf(i18n.Text("工作目录不能为空", "workdir is required"))
	}
	return EnsureProjectTrustedInFile(ConfigPath(), workdir)
}

// ConfigPath returns the Codex user config.toml path.
func ConfigPath() string {
	return filepath.Join(HomeDir(), "config.toml")
}

// HomeDir returns the Codex home directory.
func HomeDir() string {
	if home := strings.TrimSpace(os.Getenv("CODEX_HOME")); home != "" {
		return home
	}
	return filepath.Join(os.Getenv("HOME"), ".codex")
}

// EnsureProjectTrustedInFile is the file-oriented form of EnsureProjectTrusted.
// It is exported for tests that inject a temporary config path.
func EnsureProjectTrustedInFile(configPath, workdir string) (bool, error) {
	if workdir == "" {
		return false, fmt.Errorf(i18n.Text("工作目录不能为空", "workdir is required"))
	}
	if configPath == "" {
		return false, fmt.Errorf(i18n.Text("Codex 配置路径不能为空", "codex config path is required"))
	}

	header := projectHeader(workdir)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf(i18n.Text("读取 Codex 配置失败：%w", "read Codex config: %w"), err)
		}
		content := header + "\ntrust_level = \"trusted\"\n"
		if err := writeFileAtomic(configPath, []byte(content), 0600); err != nil {
			return false, err
		}
		return true, nil
	}

	updated, changed := upsertTrustedProject(string(data), header)
	if !changed {
		return false, nil
	}
	if err := writeFileAtomic(configPath, []byte(updated), fileMode(configPath, 0600)); err != nil {
		return false, err
	}
	return true, nil
}

func projectHeader(workdir string) string {
	return "[projects." + TOMLQuotedKey(workdir) + "]"
}

// TOMLQuotedKey returns a double-quoted TOML key with escapes applied.
func TOMLQuotedKey(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func upsertTrustedProject(content, header string) (string, bool) {
	lines := splitKeepEnds(content)
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(stripLineEnding(line)) == header {
			start = i
			break
		}
	}

	if start == -1 {
		builder := strings.Builder{}
		builder.WriteString(content)
		if content != "" {
			if !strings.HasSuffix(content, "\n") {
				builder.WriteByte('\n')
			}
			// Separate a newly appended table from prior content with a blank line.
			if !strings.HasSuffix(content, "\n\n") {
				builder.WriteByte('\n')
			}
		}
		builder.WriteString(header)
		builder.WriteByte('\n')
		builder.WriteString("trust_level = \"trusted\"\n")
		return builder.String(), true
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(stripLineEnding(lines[i]))
		if strings.HasPrefix(trimmed, "[") {
			end = i
			break
		}
	}

	trustLine := -1
	alreadyTrusted := false
	for i := start + 1; i < end; i++ {
		key, value, ok := parseAssignment(stripLineEnding(lines[i]))
		if !ok {
			continue
		}
		if key != "trust_level" {
			continue
		}
		trustLine = i
		if value == "trusted" {
			alreadyTrusted = true
		}
		break
	}
	if alreadyTrusted {
		return content, false
	}

	replacement := "trust_level = \"trusted\"" + lineEndingOf(lines[start])
	if trustLine >= 0 {
		lines[trustLine] = replacement
	} else {
		insertAt := start + 1
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertAt]...)
		newLines = append(newLines, replacement)
		newLines = append(newLines, lines[insertAt:]...)
		lines = newLines
	}
	return strings.Join(lines, ""), true
}

func parseAssignment(line string) (key, value string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false
	}
	eq := strings.IndexByte(trimmed, '=')
	if eq <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(trimmed[:eq])
	raw := strings.TrimSpace(trimmed[eq+1:])
	if len(raw) >= 2 {
		if (raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'') {
			raw = raw[1 : len(raw)-1]
		}
	}
	return key, raw, true
}

func splitKeepEnds(content string) []string {
	if content == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] != '\n' {
			continue
		}
		lines = append(lines, content[start:i+1])
		start = i + 1
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

func stripLineEnding(line string) string {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line
}

func lineEndingOf(line string) string {
	if strings.HasSuffix(line, "\r\n") {
		return "\r\n"
	}
	if strings.HasSuffix(line, "\n") {
		return "\n"
	}
	return "\n"
}

func fileMode(path string, fallback os.FileMode) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return fallback
	}
	return info.Mode().Perm()
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf(i18n.Text("创建 Codex 配置目录失败：%w", "create Codex config directory: %w"), err)
	}

	tmp, err := os.CreateTemp(dir, ".config.toml.*")
	if err != nil {
		return fmt.Errorf(i18n.Text("创建临时 Codex 配置失败：%w", "create temporary Codex config: %w"), err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf(i18n.Text("写入临时 Codex 配置失败：%w", "write temporary Codex config: %w"), err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf(i18n.Text("设置 Codex 配置权限失败：%w", "chmod Codex config: %w"), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf(i18n.Text("关闭临时 Codex 配置失败：%w", "close temporary Codex config: %w"), err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf(i18n.Text("更新 Codex 配置失败：%w", "replace Codex config: %w"), err)
	}
	cleanup = false
	return nil
}
