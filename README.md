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
   - Apple Silicon: `codex-morning-macos-apple-silicon`
   - Intel Mac: `codex-morning-macos-intel`
3. Rename the downloaded file to `codex-morning`.
4. Remove the browser download quarantine flag, make it executable, and move it into your PATH:

```bash
xattr -d com.apple.quarantine codex-morning 2>/dev/null || true
chmod +x codex-morning
sudo mv codex-morning /usr/local/bin/codex-morning
```

Confirm the install:

```bash
codex-morning help
```

The `xattr` command prevents the common macOS Gatekeeper prompt that says the downloaded binary cannot be verified. Building from source with `make install` or installing with `go install` avoids the browser download quarantine path.

If macOS asks whether `codex-morning` can control Terminal, you are likely running an older release. Download the latest version. Current versions open Terminal through a temporary `.command` file instead of requesting Terminal automation permission.

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

Install the daily LaunchAgent from the project directory where you want Codex to start:

```bash
cd /path/to/your/project
codex-morning install
```

You can also pass the working directory explicitly:

```bash
codex-morning install --workdir /path/to/your/project
```

Customize the time or prompt:

```bash
codex-morning install --time 09:00 --prompt "codex，早上好" --workdir /path/to/your/project
```

The installer stores the working directory in the LaunchAgent. Avoid installing from broad download folders such as `~/Downloads`; Codex may ask whether to trust the current directory before loading project-local config, hooks, and rules.

For the first run in a new directory, run this once and choose `Yes` if you trust the directory:

```bash
codex-morning run-once --workdir /path/to/your/project
```

After Codex records the directory as trusted, scheduled runs should no longer stop at that trust prompt for the same directory.

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

- `codex-morning-macos-apple-silicon` for Apple Silicon Macs
- `codex-morning-macos-intel` for Intel Macs

This repository includes a release workflow. To publish a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will run tests, build both macOS binaries, create a GitHub Release, and upload the files.

Homebrew can be added later after a tap is published. Avoid documenting a Homebrew command before it works.

## License

MIT
