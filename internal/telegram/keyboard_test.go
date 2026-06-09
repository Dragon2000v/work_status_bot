package telegram

import (
	"testing"

	"work-status-bot/internal/i18n"
)

func TestMainMenuKeyboardUkrainianReplyKeyboard(t *testing.T) {
	kb := MainMenuKeyboard(i18n.Ukrainian)
	if !kb.ResizeKeyboard || kb.OneTimeKeyboard || !kb.IsPersistent {
		t.Fatalf("bad reply keyboard flags: %#v", kb)
	}
	want := map[string]bool{
		"Додати людину":   true,
		"Почати роботу":   true,
		"Статус":          true,
		"Зупинити роботу": true,
		"Місячний звіт":   true,
		"Налаштування":    true,
		"Допомога":        true,
	}
	for _, row := range kb.Keyboard {
		for _, btn := range row {
			if !want[btn.Text] {
				t.Fatalf("unexpected button %q", btn.Text)
			}
			delete(want, btn.Text)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing buttons: %#v", want)
	}
}

func TestSettingsAndLanguageKeyboardLabels(t *testing.T) {
	settings := SettingsKeyboard(i18n.English)
	if got := settings.Keyboard[0][0].Text; got != "🌐 Мова / Language / Язык" {
		t.Fatalf("settings label = %q", got)
	}
	langs := LanguageKeyboard(i18n.Russian)
	got := []string{langs.Keyboard[0][0].Text, langs.Keyboard[1][0].Text, langs.Keyboard[2][0].Text}
	want := []string{"Українська", "English", "Русский"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("lang %d = %q", i, got[i])
		}
	}
}

func TestInlineLanguageKeyboardStillSupportsCallbacks(t *testing.T) {
	langs := InlineLanguageKeyboard(i18n.Ukrainian)
	if got := langs.InlineKeyboard[0][0].CallbackData; got != CallbackLanguageUK {
		t.Fatalf("callback = %q", got)
	}
}
