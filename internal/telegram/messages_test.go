package telegram

import (
	"strings"
	"testing"

	"work-status-bot/internal/i18n"
)

func TestLocalizedHelpAndErrors(t *testing.T) {
	if got := HelpMessageLang(i18n.Ukrainian); !strings.Contains(got, "/add_person") || !strings.Contains(got, "кнопками") {
		t.Fatalf("uk help = %q", got)
	}
	if got := HelpMessageLang(i18n.English); !strings.Contains(got, "menu buttons") {
		t.Fatalf("en help = %q", got)
	}
	if got := ErrorMessageLang(i18n.Russian, ErrMalformedCommand); !strings.HasPrefix(got, "Ошибка:") {
		t.Fatalf("ru error = %q", got)
	}
}
