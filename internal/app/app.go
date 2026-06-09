package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vampire-locker/codex-morning/internal/i18n"
	"github.com/vampire-locker/codex-morning/internal/launchd"
	"github.com/vampire-locker/codex-morning/internal/terminal"
)

const (
	DefaultLabel = "com.codex-morning.agent"
	DefaultTime  = "09:00"
)

func DefaultPrompt() string {
	return i18n.DefaultPrompt()
}

type InstallOptions struct {
	Time     string
	Prompt   string
	Workdir  string
	CodexBin string
	Label    string
	DryRun   bool
}

type RunOptions struct {
	Prompt   string
	Workdir  string
	CodexBin string
}

func Install(opts InstallOptions) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if opts.Label == "" {
		opts.Label = DefaultLabel
	}
	if opts.Prompt == "" {
		opts.Prompt = DefaultPrompt()
	}

	hour, minute, err := launchd.ParseHHMM(opts.Time)
	if err != nil {
		return err
	}

	workdir, err := resolveWorkdir(opts.Workdir)
	if err != nil {
		return err
	}
	codexBin := resolveCodexBin(opts.CodexBin)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf(i18n.Text("解析当前可执行文件失败：%w", "resolve executable: %w"), err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	agent := launchd.Agent{
		Label: opts.Label,
		ProgramArguments: []string{
			exe,
			"run-once",
			"--workdir", workdir,
			"--prompt", opts.Prompt,
			"--codex-bin", codexBin,
		},
		Hour:   hour,
		Minute: minute,
		Stdout: filepath.Join(os.Getenv("HOME"), "Library", "Logs", opts.Label+".out.log"),
		Stderr: filepath.Join(os.Getenv("HOME"), "Library", "Logs", opts.Label+".err.log"),
	}

	plist := launchd.Render(agent)
	if opts.DryRun {
		fmt.Print(plist)
		return nil
	}

	path, err := launchd.PlistPath(opts.Label)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf(i18n.Text("创建 LaunchAgents 目录失败：%w", "create LaunchAgents directory: %w"), err)
	}
	if err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), "Library", "Logs"), 0755); err != nil {
		return fmt.Errorf(i18n.Text("创建日志目录失败：%w", "create logs directory: %w"), err)
	}
	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		return fmt.Errorf(i18n.Text("写入 plist 失败：%w", "write plist: %w"), err)
	}

	_ = launchctl("bootout", userDomain(), path)
	if err := launchctl("bootstrap", userDomain(), path); err != nil {
		return err
	}
	if err := launchctl("enable", userDomain()+"/"+opts.Label); err != nil {
		return err
	}

	fmt.Printf(i18n.Text("已安装：%s\n", "Installed %s\n"), opts.Label)
	fmt.Printf(i18n.Text("执行时间：每天 %02d:%02d\n", "Schedule: every day at %02d:%02d\n"), hour, minute)
	fmt.Printf(i18n.Text("工作目录：%s\n", "Workdir: %s\n"), workdir)
	fmt.Printf(i18n.Text("plist 文件：%s\n", "Plist: %s\n"), path)
	fmt.Println(i18n.Text("如果 Codex 询问是否信任此目录，请仅对你信任的目录选择 Yes。", "If Codex asks whether to trust this directory, choose Yes once for a directory you trust."))
	return nil
}

func RunOnce(opts RunOptions) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if opts.Prompt == "" {
		opts.Prompt = DefaultPrompt()
	}
	workdir, err := resolveWorkdir(opts.Workdir)
	if err != nil {
		return err
	}
	return terminal.OpenCodex(workdir, resolveCodexBin(opts.CodexBin), opts.Prompt)
}

func Status(label string) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if label == "" {
		label = DefaultLabel
	}
	path, err := launchd.PlistPath(label)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf(i18n.Text("未安装：%s\n", "Not installed: %s\n"), path)
			return nil
		}
		return err
	}

	fmt.Printf(i18n.Text("已安装：%s\n", "Installed: %s\n"), path)
	if err := launchctl("print", userDomain()+"/"+label); err != nil {
		fmt.Println(i18n.Text("已加载：否", "Loaded: no"))
		return nil
	}
	fmt.Println(i18n.Text("已加载：是", "Loaded: yes"))
	return nil
}

func Uninstall(label string) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if label == "" {
		label = DefaultLabel
	}
	path, err := launchd.PlistPath(label)
	if err != nil {
		return err
	}

	_ = launchctl("bootout", userDomain(), path)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(i18n.Text("删除 plist 失败：%w", "remove plist: %w"), err)
	}
	fmt.Printf(i18n.Text("已卸载：%s\n", "Uninstalled %s\n"), label)
	return nil
}

func requireDarwin() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf(i18n.Text("codex-morning 目前仅支持 macOS", "codex-morning currently supports macOS only"))
	}
	return nil
}

func resolveWorkdir(value string) (string, error) {
	var path string
	var err error
	if value == "" {
		path, err = os.Getwd()
	} else {
		path, err = filepath.Abs(value)
	}
	if err != nil {
		return "", fmt.Errorf(i18n.Text("解析工作目录失败：%w", "resolve workdir: %w"), err)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(i18n.Text("工作目录不存在：%s", "workdir does not exist: %s"), path)
		}
		return "", fmt.Errorf(i18n.Text("检查工作目录失败 %s：%w", "inspect workdir %s: %w"), path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf(i18n.Text("工作目录不是目录：%s", "workdir is not a directory: %s"), path)
	}
	return path, nil
}

func resolveCodexBin(value string) string {
	if value != "" {
		return value
	}
	if found, err := exec.LookPath("codex"); err == nil {
		return found
	}
	return "codex"
}

func launchctl(args ...string) error {
	cmd := exec.Command("launchctl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(i18n.Text("执行 launchctl %v 失败：%w\n%s", "launchctl %v: %w\n%s"), args, err, string(out))
	}
	return nil
}

func userDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}
