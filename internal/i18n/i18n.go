package i18n

type Language string

const (
	Ukrainian Language = "uk"
	English   Language = "en"
	Russian   Language = "ru"
)

const (
	KeyMenuTitle         = "menu.title"
	KeyMenuAddPerson     = "menu.add_person"
	KeyMenuStartWork     = "menu.start_work"
	KeyMenuStatus        = "menu.status"
	KeyMenuStopWork      = "menu.stop_work"
	KeyMenuReportMonth   = "menu.report_month"
	KeyMenuSettings      = "menu.settings"
	KeyMenuHelp          = "menu.help"
	KeyHelpText          = "help.text"
	KeySettingsTitle     = "settings.title"
	KeySettingsLanguage  = "settings.language"
	KeyLanguageTitle     = "settings.language.title"
	KeyLanguageUkrainian = "settings.language.uk"
	KeyLanguageEnglish   = "settings.language.en"
	KeyLanguageRussian   = "settings.language.ru"
	KeyLanguageChanged   = "settings.language.changed"
	KeyUnsupportedAction = "error.unsupported_action"
	KeyMalformedCommand  = "error.malformed_command"
	KeyGenericError      = "error.generic"
	KeyPromptAddPerson   = "prompt.add_person"
	KeyPromptStartWork   = "prompt.start_work"
	KeyPromptStopWork    = "prompt.stop_work"
	KeyPromptReportMonth = "prompt.report_month"
	KeyPersonAdded       = "person.added"
	KeyWorkStarted       = "work.started"
	KeyStatusInactive    = "status.inactive"
	KeyStatusActive      = "status.active"
	KeyStatusNoPeople    = "status.no_people"
	KeyWorkStopped       = "work.stopped"
	KeyWarning           = "warning"
	KeyAlertStopped      = "alert.stopped"
	KeyReason            = "reason"
	KeyReportDuplicate   = "report.duplicate"
)

var translations = map[string]map[Language]string{
	KeyMenuTitle: {
		Ukrainian: "Оберіть дію:",
		English:   "Choose an action:",
		Russian:   "Выберите действие:",
	},
	KeyMenuAddPerson: {
		Ukrainian: "Додати людину",
		English:   "Add person",
		Russian:   "Добавить человека",
	},
	KeyMenuStartWork: {
		Ukrainian: "Почати роботу",
		English:   "Start work",
		Russian:   "Начать работу",
	},
	KeyMenuStatus: {
		Ukrainian: "Статус",
		English:   "Status",
		Russian:   "Статус",
	},
	KeyMenuStopWork: {
		Ukrainian: "Зупинити роботу",
		English:   "Stop work",
		Russian:   "Остановить работу",
	},
	KeyMenuReportMonth: {
		Ukrainian: "Місячний звіт",
		English:   "Monthly report",
		Russian:   "Месячный отчет",
	},
	KeyMenuSettings: {
		Ukrainian: "Налаштування",
		English:   "Settings",
		Russian:   "Настройки",
	},
	KeyMenuHelp: {
		Ukrainian: "Допомога",
		English:   "Help",
		Russian:   "Помощь",
	},
	KeyHelpText: {
		Ukrainian: "Команди:\n/add_person <ім'я> <прізвище>\n/start_work <ім'я> <прізвище> \"<назва>\"\n/status\n/stop_work <ім'я> <прізвище> [причина]\n/report_month [YYYY-MM]\n\nТакож можна користуватися кнопками меню.",
		English:   "Commands:\n/add_person <first_name> <last_name>\n/start_work <first_name> <last_name> \"<title>\"\n/status\n/stop_work <first_name> <last_name> [reason]\n/report_month [YYYY-MM]\n\nYou can also use the menu buttons.",
		Russian:   "Команды:\n/add_person <имя> <фамилия>\n/start_work <имя> <фамилия> \"<название>\"\n/status\n/stop_work <имя> <фамилия> [причина]\n/report_month [YYYY-MM]\n\nТакже можно использовать кнопки меню.",
	},
	KeySettingsTitle: {
		Ukrainian: "Налаштування",
		English:   "Settings",
		Russian:   "Настройки",
	},
	KeySettingsLanguage: {
		Ukrainian: "🌐 Мова / Language / Язык",
		English:   "🌐 Мова / Language / Язык",
		Russian:   "🌐 Мова / Language / Язык",
	},
	KeyLanguageTitle: {
		Ukrainian: "Оберіть мову:",
		English:   "Choose language:",
		Russian:   "Выберите язык:",
	},
	KeyLanguageUkrainian: {
		Ukrainian: "Українська",
		English:   "Українська",
		Russian:   "Українська",
	},
	KeyLanguageEnglish: {
		Ukrainian: "English",
		English:   "English",
		Russian:   "English",
	},
	KeyLanguageRussian: {
		Ukrainian: "Русский",
		English:   "Русский",
		Russian:   "Русский",
	},
	KeyLanguageChanged: {
		Ukrainian: "Мову збережено.",
		English:   "Language saved.",
		Russian:   "Язык сохранен.",
	},
	KeyUnsupportedAction: {
		Ukrainian: "Невідома дія.",
		English:   "Unknown action.",
		Russian:   "Неизвестное действие.",
	},
	KeyMalformedCommand: {
		Ukrainian: "Неправильна команда.",
		English:   "Malformed command.",
		Russian:   "Неверная команда.",
	},
	KeyGenericError: {
		Ukrainian: "Помилка",
		English:   "Error",
		Russian:   "Ошибка",
	},
	KeyPromptAddPerson: {
		Ukrainian: "Надішліть команду: /add_person <ім'я> <прізвище>",
		English:   "Send command: /add_person <first_name> <last_name>",
		Russian:   "Отправьте команду: /add_person <имя> <фамилия>",
	},
	KeyPromptStartWork: {
		Ukrainian: "Надішліть команду: /start_work <ім'я> <прізвище> \"<назва>\"",
		English:   "Send command: /start_work <first_name> <last_name> \"<title>\"",
		Russian:   "Отправьте команду: /start_work <имя> <фамилия> \"<название>\"",
	},
	KeyPromptStopWork: {
		Ukrainian: "Надішліть команду: /stop_work <ім'я> <прізвище> [причина]",
		English:   "Send command: /stop_work <first_name> <last_name> [reason]",
		Russian:   "Отправьте команду: /stop_work <имя> <фамилия> [причина]",
	},
	KeyPromptReportMonth: {
		Ukrainian: "Надішліть команду: /report_month [YYYY-MM]",
		English:   "Send command: /report_month [YYYY-MM]",
		Russian:   "Отправьте команду: /report_month [YYYY-MM]",
	},
	KeyPersonAdded: {
		Ukrainian: "Додано %s %s.",
		English:   "Added %s %s.",
		Russian:   "Добавлен(а) %s %s.",
	},
	KeyWorkStarted: {
		Ukrainian: "Роботу для %s %s розпочато: %s о %s.",
		English:   "Started work for %s %s: %s at %s.",
		Russian:   "Работа для %s %s начата: %s в %s.",
	},
	KeyStatusInactive: {
		Ukrainian: "%s %s: неактивний",
		English:   "%s %s: inactive",
		Russian:   "%s %s: неактивен",
	},
	KeyStatusActive: {
		Ukrainian: "%s %s: %s, активний з %s, минуло %s",
		English:   "%s %s: %s, active since %s, elapsed %s",
		Russian:   "%s %s: %s, активен с %s, прошло %s",
	},
	KeyStatusNoPeople: {
		Ukrainian: "Немає відстежуваних людей.",
		English:   "No people tracked.",
		Russian:   "Нет отслеживаемых людей.",
	},
	KeyWorkStopped: {
		Ukrainian: "Роботу для %s %s зупинено: %s о %s.",
		English:   "Stopped work for %s %s: %s at %s.",
		Russian:   "Работа для %s %s остановлена: %s в %s.",
	},
	KeyWarning: {
		Ukrainian: "Попередження",
		English:   "Warning",
		Russian:   "Предупреждение",
	},
	KeyAlertStopped: {
		Ukrainian: "Увага: %s %s зупинив(ла) роботу \"%s\".",
		English:   "Alert: %s %s stopped work \"%s\".",
		Russian:   "Внимание: %s %s остановил(а) работу \"%s\".",
	},
	KeyReason: {
		Ukrainian: "Причина",
		English:   "Reason",
		Russian:   "Причина",
	},
	KeyReportDuplicate: {
		Ukrainian: "Звіт уже існує.",
		English:   "Report already exists.",
		Russian:   "Отчет уже существует.",
	},
}

func Normalize(lang Language) Language {
	switch lang {
	case Ukrainian, English, Russian:
		return lang
	default:
		return Ukrainian
	}
}

func T(lang Language, key string) string {
	values, ok := translations[key]
	if !ok {
		return key
	}
	lang = Normalize(lang)
	if text, ok := values[lang]; ok && text != "" {
		return text
	}
	if text, ok := values[Ukrainian]; ok && text != "" {
		return text
	}
	return key
}

func Supported(lang Language) bool {
	switch lang {
	case Ukrainian, English, Russian:
		return true
	default:
		return false
	}
}
