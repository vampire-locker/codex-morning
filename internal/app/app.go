package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vampire-locker/codex-morning/internal/launchd"
	"github.com/vampire-locker/codex-morning/internal/terminal"
)

const (
	DefaultLabel  = "com.codex-morning.agent"
	DefaultPrompt = "codex，早上好"
	DefaultTime   = "09:00"
)

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
		opts.Prompt = DefaultPrompt
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
		return fmt.Errorf("resolve executable: %w", err)
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
		return fmt.Errorf("create launch agents directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), "Library", "Logs"), 0755); err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	_ = launchctl("bootout", userDomain(), path)
	if err := launchctl("bootstrap", userDomain(), path); err != nil {
		return err
	}
	if err := launchctl("enable", userDomain()+"/"+opts.Label); err != nil {
		return err
	}

	fmt.Printf("Installed %s\n", opts.Label)
	fmt.Printf("Schedule: every day at %02d:%02d\n", hour, minute)
	fmt.Printf("Workdir: %s\n", workdir)
	fmt.Printf("Plist: %s\n", path)
	fmt.Println("If Codex asks whether to trust this directory, choose Yes once for a directory you trust.")
	return nil
}

func RunOnce(opts RunOptions) error {
	if err := requireDarwin(); err != nil {
		return err
	}
	if opts.Prompt == "" {
		opts.Prompt = DefaultPrompt
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
			fmt.Printf("Not installed: %s\n", path)
			return nil
		}
		return err
	}

	fmt.Printf("Installed: %s\n", path)
	if err := launchctl("print", userDomain()+"/"+label); err != nil {
		fmt.Println("Loaded: no")
		return nil
	}
	fmt.Println("Loaded: yes")
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
		return fmt.Errorf("remove plist: %w", err)
	}
	fmt.Printf("Uninstalled %s\n", label)
	return nil
}

func requireDarwin() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("codex-morning currently supports macOS only")
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
		return "", fmt.Errorf("resolve workdir: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("workdir does not exist: %s", path)
		}
		return "", fmt.Errorf("inspect workdir %s: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workdir is not a directory: %s", path)
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
		return fmt.Errorf("launchctl %v: %w\n%s", args, err, string(out))
	}
	return nil
}

func userDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}
