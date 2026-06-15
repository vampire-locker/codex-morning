package launchd

import (
	"strings"
	"testing"
)

func TestParseHHMM(t *testing.T) {
	hour, minute, err := ParseHHMM("09:30")
	if err != nil {
		t.Fatal(err)
	}
	if hour != 9 || minute != 30 {
		t.Fatalf("got %02d:%02d", hour, minute)
	}
}

func TestParseHHMMRejectsInvalidValues(t *testing.T) {
	values := []string{"9", "24:00", "09:60", "aa:00", "09:bb"}
	for _, value := range values {
		if _, _, err := ParseHHMM(value); err == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
}

func TestRenderEscapesXML(t *testing.T) {
	plist := Render(Agent{
		Label:            "com.example.codex",
		ProgramArguments: []string{"/tmp/codex-morning", "run-once", "--prompt", `codex "hi" & go`},
		Hour:             9,
		Minute:           0,
		Stdout:           "/tmp/out.log",
		Stderr:           "/tmp/err.log",
	})

	required := []string{
		"<string>com.example.codex</string>",
		"<integer>9</integer>",
		"<integer>0</integer>",
		`<string>codex &#34;hi&#34; &amp; go</string>`,
	}
	for _, needle := range required {
		if !strings.Contains(plist, needle) {
			t.Fatalf("plist missing %q:\n%s", needle, plist)
		}
	}
}

func TestRenderWeekdaysOnlyCalendarInterval(t *testing.T) {
	plist := Render(Agent{
		Label:            "com.example.codex",
		ProgramArguments: []string{"/tmp/codex-morning", "run-once"},
		Hour:             9,
		Minute:           30,
		WeekdaysOnly:     true,
		Stdout:           "/tmp/out.log",
		Stderr:           "/tmp/err.log",
	})

	if !strings.Contains(plist, "<key>StartCalendarInterval</key>\n  <array>") {
		t.Fatalf("plist missing calendar interval array:\n%s", plist)
	}
	for weekday := 1; weekday <= 5; weekday++ {
		needle := "<key>Weekday</key>\n      <integer>" + string(rune('0'+weekday)) + "</integer>"
		if !strings.Contains(plist, needle) {
			t.Fatalf("plist missing weekday %d:\n%s", weekday, plist)
		}
	}
	if strings.Contains(plist, "<integer>0</integer>") || strings.Contains(plist, "<integer>6</integer>") || strings.Contains(plist, "<integer>7</integer>") {
		t.Fatalf("plist includes weekend weekday:\n%s", plist)
	}
}
