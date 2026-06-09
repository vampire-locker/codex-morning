# codex-morning

Language: English | [简体中文](README.zh-CN.md)

Open a new macOS Terminal window at the right time of day and start Codex with a preset prompt.

## What It Does

`codex-morning` is a small macOS CLI tool for scheduling Codex startup around your work routine.

It installs a user-level `launchd` agent. At the configured time, macOS opens a new Terminal window, changes into your chosen project directory, and runs Codex with your preset prompt.

This is useful when you want Codex to start at a deliberate time, such as before you arrive at work, so your daily coding session can line up better with Codex's usage window. The exact available usage and refresh time are still determined by Codex and your account plan.

The default schedule is 09:00, and the default prompt is:

```text
codex，早上好
```

## Requirements

- macOS
- Codex CLI installed and signed in
- Go 1.22+ only if building from source

## Install

### Option 1: Download a Release

This is the easiest path for most users and does not require Go.

1. Open the [Releases](https://github.com/vampire-locker/codex-morning/releases) page.
2. Download the macOS build for your Mac:
   - Apple Silicon: `codex-morning-darwin-arm64`
   - Intel Mac: `codex-morning-darwin-amd64`
3. Rename the downloaded file to `codex-morning`.
4. Make it executable and move it into your PATH:

```bash
chmod +x codex-morning
sudo mv codex-morning /usr/local/bin/codex-morning
```

Confirm the install:

```bash
codex-morning help
```

### Option 2: Build From Source

Clone the repository, then run:

```bash
git clone git@github.com:vampire-locker/codex-morning.git
cd codex-morning
make install
```

This installs the `codex-morning` binary to `/usr/local/bin` by default. You can override the destination:

```bash
make install PREFIX="$HOME/.local"
```

## Schedule Codex

Install the daily LaunchAgent:

```bash
codex-morning install
```

Customize the time or prompt:

```bash
codex-morning install --time 09:00 --prompt "codex，早上好"
```

The installer stores the current working directory in the LaunchAgent, so run it from the project directory where you want Codex to start.

## Useful Commands

```bash
codex-morning run-once
codex-morning status
codex-morning uninstall
```

`run-once` opens a new Terminal window immediately and runs:

```bash
codex "codex，早上好"
```

`uninstall` unloads and removes the LaunchAgent.

## Dry Run

Print the plist without installing it:

```bash
codex-morning install --dry-run
```

## Test

```bash
make test
```

## Build

```bash
make build
```

The binary is written to `bin/codex-morning`.

## Go Install

If you already have Go installed, you can install directly from GitHub:

```bash
go install github.com/vampire-locker/codex-morning/cmd/codex-morning@latest
```

Make sure your Go bin directory is in `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Publishing Notes

For a beginner-friendly release, publish prebuilt binaries for:

- `darwin/arm64` for Apple Silicon Macs
- `darwin/amd64` for Intel Macs

This repository includes a release workflow. To publish a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will run tests, build both macOS binaries, create a GitHub Release, and upload the files.

Homebrew can be added later after a tap is published. Avoid documenting a Homebrew command before it works.

## License

MIT
