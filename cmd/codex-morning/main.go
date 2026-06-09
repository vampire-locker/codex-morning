package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vampire-locker/codex-morning/internal/app"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "codex-morning：", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "install":
		fs := flag.NewFlagSet("install", flag.ExitOnError)
		opts := app.InstallOptions{}
		fs.StringVar(&opts.Time, "time", app.DefaultTime, "每天执行时间，格式为 HH:MM")
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt, "传给 Codex 的提示词")
		fs.StringVar(&opts.Workdir, "workdir", "", "Codex 启动时进入的工作目录")
		fs.StringVar(&opts.CodexBin, "codex-bin", "", "Codex 可执行文件路径")
		fs.StringVar(&opts.Label, "label", app.DefaultLabel, "launchd 任务标签")
		fs.BoolVar(&opts.DryRun, "dry-run", false, "只打印 plist，不安装")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Install(opts)

	case "run-once":
		fs := flag.NewFlagSet("run-once", flag.ExitOnError)
		opts := app.RunOptions{}
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt, "传给 Codex 的提示词")
		fs.StringVar(&opts.Workdir, "workdir", "", "Codex 启动时进入的工作目录")
		fs.StringVar(&opts.CodexBin, "codex-bin", "", "Codex 可执行文件路径")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.RunOnce(opts)

	case "status":
		fs := flag.NewFlagSet("status", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, "launchd 任务标签")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Status(*label)

	case "uninstall":
		fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, "launchd 任务标签")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Uninstall(*label)

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		return fmt.Errorf("未知命令 %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`codex-morning 用于在 macOS 上定时启动 Codex。

用法：
  codex-morning install [--time 09:00] [--prompt "codex，早上好"]
  codex-morning run-once [--prompt "codex，早上好"]
  codex-morning status
  codex-morning uninstall

命令：
  install    创建并加载用户级 LaunchAgent。
  run-once   立即打开新的 Terminal 窗口，并用提示词运行 Codex。
  status     查看 LaunchAgent 状态。
  uninstall  卸载并删除 LaunchAgent。`)
}
