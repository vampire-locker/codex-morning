package launchd

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Agent struct {
	Label            string
	ProgramArguments []string
	Hour             int
	Minute           int
	Stdout           string
	Stderr           string
}

func ParseHHMM(value string) (int, int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("time must use HH:MM")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("hour must be 0-23")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("minute must be 0-59")
	}
	return hour, minute, nil
}

func PlistPath(label string) (string, error) {
	if label == "" {
		return "", fmt.Errorf("label is required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
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
  <dict>
    <key>Hour</key>
    <integer>` + strconv.Itoa(agent.Hour) + `</integer>
    <key>Minute</key>
    <integer>` + strconv.Itoa(agent.Minute) + `</integer>
  </dict>

  <key>StandardOutPath</key>
  <string>` + xmlEscape(agent.Stdout) + `</string>
  <key>StandardErrorPath</key>
  <string>` + xmlEscape(agent.Stderr) + `</string>
</dict>
</plist>
`
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
