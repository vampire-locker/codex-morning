package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vampire-locker/codex-morning/internal/app"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "codex-morning:", err)
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
		fs.StringVar(&opts.Time, "time", app.DefaultTime, "daily run time in HH:MM")
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt, "prompt passed to codex")
		fs.StringVar(&opts.Workdir, "workdir", "", "working directory for codex")
		fs.StringVar(&opts.CodexBin, "codex-bin", "", "path to codex binary")
		fs.StringVar(&opts.Label, "label", app.DefaultLabel, "launchd label")
		fs.BoolVar(&opts.DryRun, "dry-run", false, "print plist without installing")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Install(opts)

	case "run-once":
		fs := flag.NewFlagSet("run-once", flag.ExitOnError)
		opts := app.RunOptions{}
		fs.StringVar(&opts.Prompt, "prompt", app.DefaultPrompt, "prompt passed to codex")
		fs.StringVar(&opts.Workdir, "workdir", "", "working directory for codex")
		fs.StringVar(&opts.CodexBin, "codex-bin", "", "path to codex binary")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.RunOnce(opts)

	case "status":
		fs := flag.NewFlagSet("status", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, "launchd label")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Status(*label)

	case "uninstall":
		fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
		label := fs.String("label", app.DefaultLabel, "launchd label")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return app.Uninstall(*label)

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`codex-morning schedules a daily Codex greeting on macOS.

Usage:
  codex-morning install [--time 09:00] [--prompt "codex，早上好"]
  codex-morning run-once [--prompt "codex，早上好"]
  codex-morning status
  codex-morning uninstall

Commands:
  install    Create and load a user LaunchAgent.
  run-once   Open a new Terminal window and run codex with the prompt.
  status     Show LaunchAgent status.
  uninstall  Unload and remove the LaunchAgent.`)
}
