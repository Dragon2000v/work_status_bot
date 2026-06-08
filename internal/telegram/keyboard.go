package telegram

import "work-status-bot/internal/i18n"

func MainMenuKeyboard(lang i18n.Language) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{
			button(lang, i18n.KeyMenuAddPerson, CallbackMenuAddPerson),
			button(lang, i18n.KeyMenuStartWork, CallbackMenuStartWork),
		},
		{
			button(lang, i18n.KeyMenuStatus, CallbackMenuStatus),
			button(lang, i18n.KeyMenuStopWork, CallbackMenuStopWork),
		},
		{
			button(lang, i18n.KeyMenuReportMonth, CallbackMenuReportMonth),
		},
		{
			button(lang, i18n.KeyMenuSettings, CallbackMenuSettings),
			button(lang, i18n.KeyMenuHelp, CallbackMenuHelp),
		},
	}}
}

func SettingsKeyboard(lang i18n.Language) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{
		button(lang, i18n.KeySettingsLanguage, CallbackSettingsLanguage),
	}}}
}

func LanguageKeyboard(lang i18n.Language) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{button(lang, i18n.KeyLanguageUkrainian, CallbackLanguageUK)},
		{button(lang, i18n.KeyLanguageEnglish, CallbackLanguageEN)},
		{button(lang, i18n.KeyLanguageRussian, CallbackLanguageRU)},
	}}
}

func button(lang i18n.Language, key, data string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: i18n.T(lang, key), CallbackData: data}
}
