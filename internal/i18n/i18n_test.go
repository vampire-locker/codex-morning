package i18n

import "testing"

func TestLanguageFromCode(t *testing.T) {
	tests := []struct {
		value string
		want  Language
	}{
		{"zh_CN.UTF-8", Chinese},
		{"zh-Hans", Chinese},
		{"en_US.UTF-8", English},
		{"C", English},
		{"", English},
	}

	for _, tt := range tests {
		if got := languageFromCode(tt.value); got != tt.want {
			t.Fatalf("languageFromCode(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestLanguageFromAppleLanguagesUsesFirstLanguage(t *testing.T) {
	if got := languageFromAppleLanguages("(zh-Hans-US, en-US)"); got != Chinese {
		t.Fatalf("got %q, want %q", got, Chinese)
	}
	if got := languageFromAppleLanguages("(en-US, zh-Hans-US)"); got != English {
		t.Fatalf("got %q, want %q", got, English)
	}
}
