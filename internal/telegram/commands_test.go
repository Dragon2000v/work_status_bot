package telegram

import "testing"

func TestKeyboardTextMapsToCanonicalActions(t *testing.T) {
	cases := map[string]string{
		"Додати людину":   CommandAddPerson,
		"Почати роботу":   CommandStartWork,
		"Статус":          CommandStatus,
		"Зупинити роботу": CommandStopWork,
		"Місячний звіт":   CommandReportMonth,
		"Налаштування":    "/settings",
		"Допомога":        CommandHelp,
	}
	for text, want := range cases {
		got, ok := CommandForKeyboardText(text)
		if !ok || got != want {
			t.Fatalf("%q -> %q %v, want %q", text, got, ok, want)
		}
	}
}

func TestParseCommandNormalizesBotUsernameSuffix(t *testing.T) {
	cmd, err := ParseCommand("/status@workbot")
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Name != CommandStatus {
		t.Fatalf("name = %q", cmd.Name)
	}
}
