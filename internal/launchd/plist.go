package launchd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
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

func Parse(data []byte) (Agent, error) {
	value, err := parsePlist(data)
	if err != nil {
		return Agent{}, err
	}
	root, ok := value.(map[string]any)
	if !ok {
		return Agent{}, fmt.Errorf(i18n.Text("plist 根节点不是 dict", "plist root is not a dict"))
	}

	agent := Agent{
		Label:            stringValue(root["Label"]),
		ProgramArguments: stringSlice(root["ProgramArguments"]),
		Stdout:           stringValue(root["StandardOutPath"]),
		Stderr:           stringValue(root["StandardErrorPath"]),
	}
	readSchedule(root["StartCalendarInterval"], &agent)
	return agent, nil
}

func parsePlist(data []byte) (any, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf(i18n.Text("plist 内容为空", "plist is empty"))
			}
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if ok && start.Name.Local == "plist" {
			return parsePlistValue(decoder)
		}
	}
}

func parsePlistValue(decoder *xml.Decoder) (any, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "dict":
			return parseDict(decoder)
		case "array":
			return parsePlistArray(decoder)
		case "string":
			return readElementText(decoder)
		case "integer":
			text, err := readElementText(decoder)
			if err != nil {
				return nil, err
			}
			return strconv.Atoi(strings.TrimSpace(text))
		default:
			return nil, fmt.Errorf(i18n.Text("不支持的 plist 元素：%s", "unsupported plist element: %s"), start.Name.Local)
		}
	}
}

func parseDict(decoder *xml.Decoder) (map[string]any, error) {
	values := map[string]any{}
	var key string
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local == "key" {
				key, err = readElementText(decoder)
				if err != nil {
					return nil, err
				}
				continue
			}
			if key == "" {
				return nil, fmt.Errorf(i18n.Text("plist dict 缺少 key", "plist dict value missing key"))
			}
			parsed, err := parseStartedValue(decoder, value)
			if err != nil {
				return nil, err
			}
			values[key] = parsed
			key = ""
		case xml.EndElement:
			if value.Name.Local == "dict" {
				return values, nil
			}
		}
	}
}

func parsePlistArray(decoder *xml.Decoder) ([]any, error) {
	var values []any
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch value := token.(type) {
		case xml.StartElement:
			parsed, err := parseStartedValue(decoder, value)
			if err != nil {
				return nil, err
			}
			values = append(values, parsed)
		case xml.EndElement:
			if value.Name.Local == "array" {
				return values, nil
			}
		}
	}
}

func parseStartedValue(decoder *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "dict":
		return parseDict(decoder)
	case "array":
		return parsePlistArray(decoder)
	case "string":
		return readElementText(decoder)
	case "integer":
		text, err := readElementText(decoder)
		if err != nil {
			return nil, err
		}
		return strconv.Atoi(strings.TrimSpace(text))
	default:
		return nil, fmt.Errorf(i18n.Text("不支持的 plist 元素：%s", "unsupported plist element: %s"), start.Name.Local)
	}
}

func readElementText(decoder *xml.Decoder) (string, error) {
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch value := token.(type) {
		case xml.CharData:
			builder.Write([]byte(value))
		case xml.EndElement:
			return builder.String(), nil
		}
	}
}

func readSchedule(value any, agent *Agent) {
	if dict, ok := value.(map[string]any); ok {
		agent.Hour = intValue(dict["Hour"])
		agent.Minute = intValue(dict["Minute"])
		return
	}

	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return
	}
	weekdays := map[int]bool{}
	for index, item := range items {
		dict, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if index == 0 {
			agent.Hour = intValue(dict["Hour"])
			agent.Minute = intValue(dict["Minute"])
		}
		weekday := intValue(dict["Weekday"])
		if weekday != 0 {
			weekdays[weekday] = true
		}
	}
	agent.WeekdaysOnly = len(weekdays) == 5 && weekdays[1] && weekdays[2] && weekdays[3] && weekdays[4] && weekdays[5]
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func intValue(value any) int {
	number, _ := value.(int)
	return number
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			values = append(values, text)
		}
	}
	return values
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
