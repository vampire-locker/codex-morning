package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	Time         string
	Prompt       string
	Workdir      string
	CodexBin     string
	Label        string
	WeekdaysOnly bool
	DryRun       bool
}

type RunOptions struct {
	Prompt   string
	Workdir  string
	CodexBin string
}

type LogsOptions struct {
	Label  string
	Follow bool
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
		Hour:         hour,
		Minute:       minute,
		WeekdaysOnly: opts.WeekdaysOnly,
		Stdout:       stdoutLogPath(opts.Label),
		Stderr:       stderrLogPath(opts.Label),
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
	if opts.WeekdaysOnly {
		fmt.Printf(i18n.Text("执行时间：周一到周五 %02d:%02d\n", "Schedule: Monday through Friday at %02d:%02d\n"), hour, minute)
	} else {
		fmt.Printf(i18n.Text("执行时间：每天 %02d:%02d\n", "Schedule: every day at %02d:%02d\n"), hour, minute)
	}
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

func Status(label string, verbose bool) error {
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
	} else {
		fmt.Println(i18n.Text("已加载：是", "Loaded: yes"))
	}
	if verbose {
		agent, err := readInstalledAgent(path)
		if err != nil {
			return err
		}
		printAgentDetails(path, agent)
	}
	return nil
}

func Logs(opts LogsOptions) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if opts.Label == "" {
		opts.Label = DefaultLabel
	}
	stdout := stdoutLogPath(opts.Label)
	stderr := stderrLogPath(opts.Label)
	if opts.Follow {
		return followLogs(stdout, stderr)
	}

	printLogFile(i18n.Text("标准输出日志", "stdout log"), stdout)
	printLogFile(i18n.Text("标准错误日志", "stderr log"), stderr)
	return nil
}

func Doctor(label string) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if label == "" {
		label = DefaultLabel
	}

	issues := 0
	check := func(name string, ok bool, detail string) {
		status := "[OK]"
		if !ok {
			status = "[FAIL]"
			issues++
		}
		if detail == "" {
			fmt.Printf("%s %s\n", status, name)
			return
		}
		fmt.Printf("%s %s: %s\n", status, name, detail)
	}

	check(i18n.Text("系统", "System"), runtime.GOOS == "darwin", runtime.GOOS)

	path, err := launchd.PlistPath(label)
	if err != nil {
		return err
	}
	info, statErr := os.Stat(path)
	installed := statErr == nil && !info.IsDir()
	if statErr != nil && !os.IsNotExist(statErr) {
		check(i18n.Text("LaunchAgent plist", "LaunchAgent plist"), false, statErr.Error())
	} else {
		check(i18n.Text("LaunchAgent plist", "LaunchAgent plist"), installed, path)
	}

	var agent launchd.Agent
	if installed {
		agent, err = readInstalledAgent(path)
		if err != nil {
			check(i18n.Text("解析 plist", "Parse plist"), false, err.Error())
		} else {
			check(i18n.Text("解析 plist", "Parse plist"), true, scheduleText(agent))
		}

		err = launchctl("print", userDomain()+"/"+label)
		check(i18n.Text("launchd 已加载", "launchd loaded"), err == nil, label)
	}

	workdir := argValue(agent.ProgramArguments, "--workdir")
	if workdir != "" {
		info, err := os.Stat(workdir)
		check(i18n.Text("工作目录", "Workdir"), err == nil && info.IsDir(), workdir)
	}

	codexBin := argValue(agent.ProgramArguments, "--codex-bin")
	if codexBin == "" {
		codexBin = "codex"
	}
	check(i18n.Text("Codex 可执行文件", "Codex binary"), codexBinAvailable(codexBin), codexBin)
	check(i18n.Text("Terminal 应用", "Terminal app"), terminalAppExists(), "Terminal")

	if issues > 0 {
		return fmt.Errorf(i18n.Text("doctor 发现 %d 个问题", "doctor found %d issue(s)"), issues)
	}
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

func stdoutLogPath(label string) string {
	return filepath.Join(os.Getenv("HOME"), "Library", "Logs", label+".out.log")
}

func stderrLogPath(label string) string {
	return filepath.Join(os.Getenv("HOME"), "Library", "Logs", label+".err.log")
}

func readInstalledAgent(path string) (launchd.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return launchd.Agent{}, fmt.Errorf(i18n.Text("读取 plist 失败：%w", "read plist: %w"), err)
	}
	agent, err := launchd.Parse(data)
	if err != nil {
		return launchd.Agent{}, fmt.Errorf(i18n.Text("解析 plist 失败：%w", "parse plist: %w"), err)
	}
	return agent, nil
}

func printAgentDetails(path string, agent launchd.Agent) {
	fmt.Printf(i18n.Text("标签：%s\n", "Label: %s\n"), agent.Label)
	fmt.Printf(i18n.Text("计划：%s\n", "Schedule: %s\n"), scheduleText(agent))
	fmt.Printf(i18n.Text("工作目录：%s\n", "Workdir: %s\n"), argValue(agent.ProgramArguments, "--workdir"))
	fmt.Printf(i18n.Text("Codex 路径：%s\n", "Codex binary: %s\n"), argValue(agent.ProgramArguments, "--codex-bin"))
	fmt.Printf(i18n.Text("提示词：%s\n", "Prompt: %s\n"), argValue(agent.ProgramArguments, "--prompt"))
	fmt.Printf(i18n.Text("标准输出日志：%s\n", "Stdout log: %s\n"), agent.Stdout)
	fmt.Printf(i18n.Text("标准错误日志：%s\n", "Stderr log: %s\n"), agent.Stderr)
	fmt.Printf(i18n.Text("plist 文件：%s\n", "Plist: %s\n"), path)
}

func scheduleText(agent launchd.Agent) string {
	if agent.WeekdaysOnly {
		return fmt.Sprintf(i18n.Text("周一到周五 %02d:%02d", "Monday through Friday at %02d:%02d"), agent.Hour, agent.Minute)
	}
	return fmt.Sprintf(i18n.Text("每天 %02d:%02d", "every day at %02d:%02d"), agent.Hour, agent.Minute)
}

func argValue(args []string, name string) string {
	for index := 0; index < len(args)-1; index++ {
		if args[index] == name {
			return args[index+1]
		}
	}
	return ""
}

func printLogFile(title, path string) {
	fmt.Printf("==> %s: %s <==\n", title, path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(i18n.Text("日志文件还不存在。", "Log file does not exist yet."))
			return
		}
		fmt.Println(err)
		return
	}
	text := lastLines(string(data), 80)
	if strings.TrimSpace(text) == "" {
		fmt.Println(i18n.Text("日志为空。", "Log is empty."))
		return
	}
	fmt.Print(text)
	if !strings.HasSuffix(text, "\n") {
		fmt.Println()
	}
}

func followLogs(paths ...string) error {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	if len(existing) == 0 {
		return fmt.Errorf(i18n.Text("日志文件还不存在", "log files do not exist yet"))
	}

	args := append([]string{"-n", "80", "-f"}, existing...)
	cmd := exec.Command("tail", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func lastLines(text string, limit int) string {
	if limit <= 0 || text == "" {
		return ""
	}
	lines := strings.SplitAfter(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) <= limit {
		return strings.Join(lines, "")
	}
	return strings.Join(lines[len(lines)-limit:], "")
}

func codexBinAvailable(value string) bool {
	if value == "" {
		return false
	}
	if strings.ContainsRune(value, filepath.Separator) {
		info, err := os.Stat(value)
		return err == nil && !info.IsDir()
	}
	_, err := exec.LookPath(value)
	return err == nil
}

func terminalAppExists() bool {
	for _, path := range []string{
		"/System/Applications/Utilities/Terminal.app",
		"/Applications/Utilities/Terminal.app",
	} {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}
