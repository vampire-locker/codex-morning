package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func OpenCodex(workdir, codexBin, prompt string) error {
	if workdir == "" {
		return fmt.Errorf("工作目录不能为空")
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
		return fmt.Errorf("打开 Terminal 失败：%w\n%s", err, string(out))
	}
	return nil
}

func WriteCommandFile(workdir, codexBin, prompt string) (string, error) {
	file, err := os.CreateTemp("", "codex-morning-*.command")
	if err != nil {
		return "", fmt.Errorf("创建命令文件失败：%w", err)
	}
	path := file.Name()

	content := "#!/bin/zsh\nrm -f \"$0\"\n" + BuildShellCommand(workdir, codexBin, prompt) + "\n"
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("写入命令文件失败：%w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("关闭命令文件失败：%w", err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("设置命令文件权限失败：%w", err)
	}

	return path, nil
}

func BuildShellCommand(workdir, codexBin, prompt string) string {
	return "cd " + ShellQuote(workdir) + " && " + ShellQuote(codexBin) + " " + ShellQuote(prompt)
}

func ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
