package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vampire-locker/codex-morning/internal/i18n"
)

func OpenCodex(workdir, codexBin, prompt string) error {
	if workdir == "" {
		return fmt.Errorf(i18n.Text("工作目录不能为空", "workdir is required"))
	}
	if codexBin == "" {
		codexBin = "codex"
	}
	scriptPath, err := WriteCommandFile(workdir, codexBin, prompt)
	if err != nil {
		return err
	}

	cmd := exec.Command("/usr/bin/open", "-a", "Terminal", scriptPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.Remove(scriptPath)
		return fmt.Errorf(i18n.Text("打开 Terminal 失败：%w\n%s", "open Terminal: %w\n%s"), err, string(out))
	}
	return nil
}

func WriteCommandFile(workdir, codexBin, prompt string) (string, error) {
	file, err := os.CreateTemp("", "codex-morning-*.command")
	if err != nil {
		return "", fmt.Errorf(i18n.Text("创建命令文件失败：%w", "create command file: %w"), err)
	}
	path := file.Name()

	content := "#!/bin/zsh\nrm -f \"$0\"\n" + BuildShellCommand(workdir, codexBin, prompt) + "\n"
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf(i18n.Text("写入命令文件失败：%w", "write command file: %w"), err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf(i18n.Text("关闭命令文件失败：%w", "close command file: %w"), err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf(i18n.Text("设置命令文件权限失败：%w", "chmod command file: %w"), err)
	}

	return path, nil
}

func BuildShellCommand(workdir, codexBin, prompt string) string {
	return "cd " + ShellQuote(workdir) + " && " + ShellQuote(codexBin) + " " + ShellQuote(prompt)
}

func ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
