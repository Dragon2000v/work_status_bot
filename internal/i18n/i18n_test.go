package i18n

import "testing"

func TestTranslationLookupAndFallback(t *testing.T) {
	if got := T("", KeyMenuStatus); got != "Статус" {
		t.Fatalf("default Ukrainian = %q", got)
	}
	if got := T(Language("de"), KeyMenuHelp); got != "Допомога" {
		t.Fatalf("unsupported fallback = %q", got)
	}
	if got := T(English, KeyMenuHelp); got != "Help" {
		t.Fatalf("english = %q", got)
	}
	if got := T(Russian, KeyMenuHelp); got != "Помощь" {
		t.Fatalf("russian = %q", got)
	}
	if got := T(English, "missing.key"); got != "missing.key" {
		t.Fatalf("missing key = %q", got)
	}
}

func TestSupportedLanguages(t *testing.T) {
	for _, lang := range []Language{Ukrainian, English, Russian} {
		if !Supported(lang) {
			t.Fatalf("%s should be supported", lang)
		}
	}
	if Supported("pl") {
		t.Fatalf("pl should not be supported")
	}
}
