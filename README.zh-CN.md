# codex-morning

语言：[English](README.md) | 简体中文

在每天合适的时间自动打开一个新的 macOS Terminal 窗口，并使用预设提示词启动 Codex。

## 这个工具做什么

`codex-morning` 是一个 macOS 命令行小工具，用来按照你的工作节奏定时启动 Codex。

它会安装一个用户级 `launchd` 定时任务。到达设定时间后，macOS 会打开一个新的 Terminal 窗口，进入你指定的项目目录，并用预设提示词运行 Codex。

这个工具适合希望在合适时间启动 Codex 的用户，例如在到公司前提前唤起 Codex，让每天的编码时段更好地对齐 Codex 的使用窗口。具体可用额度和刷新时间仍以 Codex 及你的账号套餐显示为准。

默认执行时间是 09:00，默认提示词是：

```text
codex，早上好
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
4. 添加执行权限，并移动到 PATH 目录：

```bash
chmod +x codex-morning
sudo mv codex-morning /usr/local/bin/codex-morning
```

确认安装成功：

```bash
codex-morning help
```

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

安装每天执行的 LaunchAgent：

```bash
codex-morning install
```

自定义时间或提示词：

```bash
codex-morning install --time 09:00 --prompt "codex，早上好"
```

安装器会把当前工作目录写入 LaunchAgent，所以请在你希望 Codex 启动的项目目录里执行安装命令。

## 常用命令

```bash
codex-morning run-once
codex-morning status
codex-morning uninstall
```

`run-once` 会立即打开一个新的 Terminal 窗口并运行：

```bash
codex "codex，早上好"
```

`uninstall` 会卸载并删除 LaunchAgent。

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
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions 会自动运行测试，构建两个 macOS 二进制文件，创建 GitHub Release，并上传文件。

Homebrew 可以等 tap 真正发布后再补。不要在 README 里提前写还不能运行的 Homebrew 命令。

## 许可证

MIT
