package launchd

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vampire-locker/codex-morning/internal/i18n"
)

type Agent struct {
	Label            string
	ProgramArguments []string
	Hour             int
	Minute           int
	WeekdaysOnly     bool
	Stdout           string
	Stderr           string
}

func ParseHHMM(value string) (int, int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf(i18n.Text("时间必须使用 HH:MM 格式", "time must use HH:MM"))
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf(i18n.Text("小时必须在 0-23 之间", "hour must be 0-23"))
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf(i18n.Text("分钟必须在 0-59 之间", "minute must be 0-59"))
	}
	return hour, minute, nil
}

func PlistPath(label string) (string, error) {
	if label == "" {
		return "", fmt.Errorf(i18n.Text("label 不能为空", "label is required"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf(i18n.Text("解析用户主目录失败：%w", "resolve home directory: %w"), err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func Render(agent Agent) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>` + xmlEscape(agent.Label) + `</string>

  <key>ProgramArguments</key>
  <array>
` + renderArray(agent.ProgramArguments) + `  </array>

  <key>StartCalendarInterval</key>
` + renderCalendarInterval(agent) + `

  <key>StandardOutPath</key>
  <string>` + xmlEscape(agent.Stdout) + `</string>
  <key>StandardErrorPath</key>
  <string>` + xmlEscape(agent.Stderr) + `</string>
</dict>
</plist>
`
}

func renderCalendarInterval(agent Agent) string {
	if !agent.WeekdaysOnly {
		return `  <dict>
    <key>Hour</key>
    <integer>` + strconv.Itoa(agent.Hour) + `</integer>
    <key>Minute</key>
    <integer>` + strconv.Itoa(agent.Minute) + `</integer>
  </dict>`
	}

	var builder strings.Builder
	builder.WriteString("  <array>\n")
	for weekday := 1; weekday <= 5; weekday++ {
		builder.WriteString("    <dict>\n")
		builder.WriteString("      <key>Weekday</key>\n")
		builder.WriteString("      <integer>")
		builder.WriteString(strconv.Itoa(weekday))
		builder.WriteString("</integer>\n")
		builder.WriteString("      <key>Hour</key>\n")
		builder.WriteString("      <integer>")
		builder.WriteString(strconv.Itoa(agent.Hour))
		builder.WriteString("</integer>\n")
		builder.WriteString("      <key>Minute</key>\n")
		builder.WriteString("      <integer>")
		builder.WriteString(strconv.Itoa(agent.Minute))
		builder.WriteString("</integer>\n")
		builder.WriteString("    </dict>\n")
	}
	builder.WriteString("  </array>")
	return builder.String()
}

func renderArray(values []string) string {
	var builder strings.Builder
	for _, value := range values {
		builder.WriteString("    <string>")
		builder.WriteString(xmlEscape(value))
		builder.WriteString("</string>\n")
	}
	return builder.String()
}

func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}
