package telegram

import "work-status-bot/internal/i18n"

const (
	CallbackMenuAddPerson   = "menu:add_person"
	CallbackMenuStartWork   = "menu:start_work"
	CallbackMenuStatus      = "menu:status"
	CallbackMenuStopWork    = "menu:stop_work"
	CallbackMenuReportMonth = "menu:report_month"
	CallbackMenuSettings    = "menu:settings"
	CallbackMenuHelp        = "menu:help"

	CallbackSettingsLanguage = "settings:language"
	CallbackLanguageUK       = "settings:language:uk"
	CallbackLanguageEN       = "settings:language:en"
	CallbackLanguageRU       = "settings:language:ru"
)

func languageFromCallback(data string) (i18n.Language, bool) {
	switch data {
	case CallbackLanguageUK:
		return i18n.Ukrainian, true
	case CallbackLanguageEN:
		return i18n.English, true
	case CallbackLanguageRU:
		return i18n.Russian, true
	default:
		return i18n.Ukrainian, false
	}
}
