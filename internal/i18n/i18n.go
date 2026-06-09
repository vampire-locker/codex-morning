package i18n

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Language string

const (
	Chinese Language = "zh"
	English Language = "en"
)

func Current() Language {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := os.Getenv(name); value != "" {
			return languageFromCode(value)
		}
	}

	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("/usr/bin/defaults", "read", "-g", "AppleLanguages").Output(); err == nil {
			return languageFromAppleLanguages(string(out))
		}
	}

	return English
}

func Text(zh, en string) string {
	if Current() == Chinese {
		return zh
	}
	return en
}

func DefaultPrompt() string {
	return Text("codex，早上好", "codex, good morning")
}

func Usage(defaultPrompt string) string {
	return Text(`codex-morning 用于在 macOS 上定时启动 Codex。

用法：
  codex-morning install [--time 09:00] [--prompt "`+defaultPrompt+`"]
  codex-morning run-once [--prompt "`+defaultPrompt+`"]
  codex-morning status
  codex-morning uninstall

命令：
  install    创建并加载用户级 LaunchAgent。
  run-once   立即打开新的 Terminal 窗口，并用提示词运行 Codex。
  status     查看 LaunchAgent 状态。
  uninstall  卸载并删除 LaunchAgent。`, `codex-morning schedules Codex startup on macOS.

Usage:
  codex-morning install [--time 09:00] [--prompt "`+defaultPrompt+`"]
  codex-morning run-once [--prompt "`+defaultPrompt+`"]
  codex-morning status
  codex-morning uninstall

Commands:
  install    Create and load a user LaunchAgent.
  run-once   Open a new Terminal window and run Codex with the prompt.
  status     Show LaunchAgent status.
  uninstall  Unload and remove the LaunchAgent.`)
}

func languageFromAppleLanguages(value string) Language {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == '(' || r == ')' || r == ',' || r == '"' || r == '\'' || r == '\n' || r == '\t' || r == ' '
	})
	for _, field := range fields {
		if field != "" {
			return languageFromCode(field)
		}
	}
	return English
}

func languageFromCode(value string) Language {
	value = strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(value, "zh") {
		return Chinese
	}
	return English
}
