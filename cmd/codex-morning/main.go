package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vampire-locker/codex-morning/internal/app"
	"github.com/vampire-locker/codex-morning/internal/i18n"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, i18n.Text("codex-morning：", "codex-morning:"), err)
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
		fs.StringVar(&opts.Time, "time", app.DefaultTime, i18n.Text("每天执行时间，格式为 HH:MM", "daily run time in HH:MM"))
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt(), i18n.Text("传给 Codex 的提示词", "prompt passed to Codex"))
		fs.StringVar(&opts.Workdir, "workdir", "", i18n.Text("Codex 启动时进入的工作目录", "working directory for Codex"))
		fs.StringVar(&opts.CodexBin, "codex-bin", "", i18n.Text("Codex 可执行文件路径", "path to Codex binary"))
		fs.StringVar(&opts.Label, "label", app.DefaultLabel, i18n.Text("launchd 任务标签", "launchd label"))
		fs.BoolVar(&opts.DryRun, "dry-run", false, i18n.Text("只打印 plist，不安装", "print plist without installing"))
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Install(opts)

	case "run-once":
		fs := flag.NewFlagSet("run-once", flag.ExitOnError)
		opts := app.RunOptions{}
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt(), i18n.Text("传给 Codex 的提示词", "prompt passed to Codex"))
		fs.StringVar(&opts.Workdir, "workdir", "", i18n.Text("Codex 启动时进入的工作目录", "working directory for Codex"))
		fs.StringVar(&opts.CodexBin, "codex-bin", "", i18n.Text("Codex 可执行文件路径", "path to Codex binary"))
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.RunOnce(opts)

	case "status":
		fs := flag.NewFlagSet("status", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, i18n.Text("launchd 任务标签", "launchd label"))
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Status(*label)

	case "uninstall":
		fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, i18n.Text("launchd 任务标签", "launchd label"))
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Uninstall(*label)

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		return fmt.Errorf(i18n.Text("未知命令 %q", "unknown command %q"), args[0])
	}
}

func printUsage() {
	fmt.Println(i18n.Usage(app.DefaultPrompt()))
}
