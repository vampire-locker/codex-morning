package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunPrintsChineseUsage(t *testing.T) {
	output := captureStdout(t, func() {
		if err := run([]string{"help"}); err != nil {
			t.Fatal(err)
		}
	})

	required := []string{
		"codex-morning 用于在 macOS 上定时启动 Codex。",
		"用法：",
		"命令：",
		"创建并加载用户级 LaunchAgent。",
	}
	for _, value := range required {
		if !strings.Contains(output, value) {
			t.Fatalf("usage missing %q:\n%s", value, output)
		}
	}
}

func TestRunRejectsUnknownCommandInChinese(t *testing.T) {
	err := run([]string{"missing"})
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.Contains(err.Error(), "未知命令") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = write
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	output, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if err := read.Close(); err != nil {
		t.Fatal(err)
	}
	return string(output)
}
