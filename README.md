# codex-morning

语言：简体中文 | [English](README.en.md)

在每天合适的时间自动打开一个新的 macOS Terminal 窗口，并使用预设提示词启动 Codex。

## 这个工具做什么

`codex-morning` 是一个 macOS 命令行小工具，用来按照你的工作节奏定时启动 Codex。

它会安装一个用户级 `launchd` 定时任务。到达设定时间后，macOS 会打开一个新的 Terminal 窗口，进入你指定的项目目录，并用预设提示词运行 Codex。

这个工具适合希望在合适时间启动 Codex 的用户，例如在到公司前提前唤起 Codex，让每天的编码时段更好地对齐 Codex 的使用窗口。

实际例子：如果你上午 10 点上班，12 点到 14 点休息，可以在早上 9 点提前激活使用 Codex。这样 5 小时使用窗口刚好在下午 14 点刷新，回到工位后可以继续使用。具体可用额度和刷新时间仍以 Codex 及你的账号套餐显示为准。

默认执行时间是 09:00。终端提示和默认 Codex 提示词会根据系统语言自动选择：

```text
中文系统：codex，早上好
其他语言系统：codex, good morning
```

## 环境要求

- macOS
- 已安装并登录 Codex CLI
- 只有从源码构建时才需要 Go 1.22+

## 安装

### 方式一：下载 Release

这是最适合普通用户的方式，不需要安装 Go。

1. 打开 [Releases](https://github.com/vampire-locker/codex-morning/releases) 页面。
2. 根据你的 Mac 下载对应文件：
   - Apple Silicon：`codex-morning-macos-apple-silicon`
   - Intel Mac：`codex-morning-macos-intel`
3. 把下载的文件重命名为 `codex-morning`。
4. 移除浏览器下载附加的 quarantine 标记，添加执行权限，并移动到 PATH 目录：

```bash
xattr -d com.apple.quarantine codex-morning 2>/dev/null || true
chmod +x codex-morning
sudo mv codex-morning /usr/local/bin/codex-morning
```

确认安装成功：

```bash
codex-morning help
```

上面的 `xattr` 命令用于避免 macOS Gatekeeper 弹出“无法验证下载文件”的提示。如果使用 `make install` 从源码构建，或者使用 `go install` 安装，就不会走浏览器下载 quarantine 这条路径。

如果 macOS 提示 `codex-morning` 想要控制 Terminal，通常说明你运行的是旧版本。请下载最新版本。当前版本会通过临时 `.command` 文件打开 Terminal，不再请求控制 Terminal 的自动化权限。

### 方式二：从源码构建

克隆仓库后运行：

```bash
git clone git@github.com:vampire-locker/codex-morning.git
cd codex-morning
make install
```

默认会把 `codex-morning` 安装到 `/usr/local/bin`。也可以指定安装目录：

```bash
make install PREFIX="$HOME/.local"
```

## 设置定时启动

请在你希望 Codex 启动的项目目录里安装每天执行的 LaunchAgent：

```bash
cd /path/to/your/project
codex-morning install
```

也可以显式指定工作目录：

```bash
codex-morning install --workdir /path/to/your/project
```

自定义时间或提示词：

```bash
codex-morning install --time 09:00 --prompt "codex，早上好" --workdir /path/to/your/project
```

只在周一到周五启动，周末不启动：

```bash
codex-morning install --weekdays-only
```

安装器会把工作目录写入 LaunchAgent。不要在 `~/Downloads` 这类宽泛的下载目录里直接安装。

为了支持无人值守启动，`codex-morning` 会在 `install` / `run-once` 时把工作目录持久写入 `~/.codex/config.toml`（或 `$CODEX_HOME/config.toml`）的 trusted project，并在启动 Codex 时再附带一次 `-c` 覆盖。仅靠运行时 `-c` 在当前 Codex 交互模式下往往仍会停在目录信任提示；持久写入后一般就不会再询问。请只为你信任的项目目录安装任务，因为 trusted project 允许 Codex 加载项目级 `.codex/config.toml`、hooks 和 rules。

## 常用命令

```bash
codex-morning run-once
codex-morning status
codex-morning status --verbose
codex-morning logs
codex-morning doctor
codex-morning uninstall
```

`run-once` 会先把工作目录标记为 Codex trusted project（写入用户 `config.toml`），再打开新的 Terminal 窗口，并用当前系统语言对应的默认提示词启动 Codex。例如中文系统会运行类似：

```bash
codex -c 'projects."/path/to/your/project".trust_level="trusted"' "codex，早上好"
```

对应 `config.toml` 片段：

```toml
[projects."/path/to/your/project"]
trust_level = "trusted"
```

其他语言系统默认提示词为 `codex, good morning`。

`uninstall` 会卸载并删除 LaunchAgent。

`status --verbose` 会显示已安装的执行计划、工作目录、Codex 路径、提示词和日志路径。

`logs` 会打印 LaunchAgent 最近的标准输出和标准错误日志。需要持续跟踪日志时使用：

```bash
codex-morning logs --follow
```

`doctor` 会检查 macOS 环境、LaunchAgent plist、launchd 加载状态、工作目录、Codex 可执行文件和 Terminal 应用，适合排查定时任务没有启动的问题。

## 预览配置

只打印 plist，不实际安装：

```bash
codex-morning install --dry-run
```

## 测试

```bash
make test
```

## 构建

```bash
make build
```

生成的二进制文件位于 `bin/codex-morning`。

## Go Install

如果你已经安装了 Go，也可以直接从 GitHub 安装：

```bash
go install github.com/vampire-locker/codex-morning/cmd/codex-morning@latest
```

确保 Go 的 bin 目录在 `PATH` 中：

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## 发布说明

为了让普通用户更容易安装，建议发布预编译二进制：

- `codex-morning-macos-apple-silicon`：Apple Silicon Mac
- `codex-morning-macos-intel`：Intel Mac

仓库里已经包含 release workflow。发布新版本时运行：

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

把 `vX.Y.Z` 替换为下一个语义化版本号，例如 `v0.1.4`。

GitHub Actions 会自动运行测试，构建两个 macOS 二进制文件，创建 GitHub Release，并上传文件。

Homebrew 可以等 tap 真正发布后再补。不要在 README 里提前写还不能运行的 Homebrew 命令。

## 许可证

MIT
