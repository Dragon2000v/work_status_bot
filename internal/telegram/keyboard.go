package telegram

import "work-status-bot/internal/i18n"

func MainMenuKeyboard(lang i18n.Language) *ReplyKeyboardMarkup {
	return replyKeyboard([][]KeyboardButton{
		{
			replyButton(lang, i18n.KeyMenuAddPerson),
			replyButton(lang, i18n.KeyMenuStartWork),
		},
		{
			replyButton(lang, i18n.KeyMenuStatus),
			replyButton(lang, i18n.KeyMenuStopWork),
		},
		{
			replyButton(lang, i18n.KeyMenuReportMonth),
		},
		{
			replyButton(lang, i18n.KeyMenuSettings),
			replyButton(lang, i18n.KeyMenuHelp),
		},
	})
}

func SettingsKeyboard(lang i18n.Language) *ReplyKeyboardMarkup {
	return replyKeyboard([][]KeyboardButton{
		{replyButton(lang, i18n.KeySettingsLanguage)},
		{replyButton(lang, i18n.KeyMenuHelp), replyButton(lang, i18n.KeyCancel)},
	})
}

func LanguageKeyboard(lang i18n.Language) *ReplyKeyboardMarkup {
	return replyKeyboard([][]KeyboardButton{
		{replyButton(lang, i18n.KeyLanguageUkrainian)},
		{replyButton(lang, i18n.KeyLanguageEnglish)},
		{replyButton(lang, i18n.KeyLanguageRussian)},
		{replyButton(lang, i18n.KeyBack), replyButton(lang, i18n.KeyCancel)},
	})
}

func StartTimeModeKeyboard(lang i18n.Language) *ReplyKeyboardMarkup {
	return replyKeyboard([][]KeyboardButton{
		{replyButton(lang, i18n.KeyStartNow), replyButton(lang, i18n.KeyStartToday)},
		{replyButton(lang, i18n.KeyStartYesterday), replyButton(lang, i18n.KeyStartManual)},
		{replyButton(lang, i18n.KeyCancel)},
	})
}

func replyKeyboard(rows [][]KeyboardButton) *ReplyKeyboardMarkup {
	return &ReplyKeyboardMarkup{
		Keyboard:        rows,
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
		IsPersistent:    true,
	}
}

func replyButton(lang i18n.Language, key string) KeyboardButton {
	return KeyboardButton{Text: i18n.T(lang, key)}
}

func button(lang i18n.Language, key, data string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: i18n.T(lang, key), CallbackData: data}
}

func InlineSettingsKeyboard(lang i18n.Language) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{
		button(lang, i18n.KeySettingsLanguage, CallbackSettingsLanguage),
	}}}
}

func InlineLanguageKeyboard(lang i18n.Language) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{button(lang, i18n.KeyLanguageUkrainian, CallbackLanguageUK)},
		{button(lang, i18n.KeyLanguageEnglish, CallbackLanguageEN)},
		{button(lang, i18n.KeyLanguageRussian, CallbackLanguageRU)},
	}}
}
