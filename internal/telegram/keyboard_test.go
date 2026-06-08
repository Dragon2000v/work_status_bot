package telegram

import (
	"testing"

	"work-status-bot/internal/i18n"
)

func TestMainMenuKeyboardUkrainianLabelsAndCallbacks(t *testing.T) {
	kb := MainMenuKeyboard(i18n.Ukrainian)
	want := map[string]string{
		CallbackMenuAddPerson:   "Додати людину",
		CallbackMenuStartWork:   "Почати роботу",
		CallbackMenuStatus:      "Статус",
		CallbackMenuStopWork:    "Зупинити роботу",
		CallbackMenuReportMonth: "Місячний звіт",
		CallbackMenuSettings:    "Налаштування",
		CallbackMenuHelp:        "Допомога",
	}
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if want[btn.CallbackData] != btn.Text {
				t.Fatalf("button %s = %q", btn.CallbackData, btn.Text)
			}
			delete(want, btn.CallbackData)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing buttons: %#v", want)
	}
}

func TestSettingsAndLanguageKeyboardLabels(t *testing.T) {
	settings := SettingsKeyboard(i18n.English)
	if got := settings.InlineKeyboard[0][0].Text; got != "🌐 Мова / Language / Язык" {
		t.Fatalf("settings label = %q", got)
	}
	langs := LanguageKeyboard(i18n.Russian)
	got := []string{langs.InlineKeyboard[0][0].Text, langs.InlineKeyboard[1][0].Text, langs.InlineKeyboard[2][0].Text}
	want := []string{"Українська", "English", "Русский"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("lang %d = %q", i, got[i])
		}
	}
}
