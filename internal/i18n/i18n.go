package i18n

type Language string

const (
	Ukrainian Language = "uk"
	English   Language = "en"
	Russian   Language = "ru"
)

const (
	KeyMenuTitle            = "menu.title"
	KeyMenuAddPerson        = "menu.add_person"
	KeyMenuStartWork        = "menu.start_work"
	KeyMenuStatus           = "menu.status"
	KeyMenuStopWork         = "menu.stop_work"
	KeyMenuReportMonth      = "menu.report_month"
	KeyMenuSettings         = "menu.settings"
	KeyMenuHelp             = "menu.help"
	KeyHelpText             = "help.text"
	KeySettingsTitle        = "settings.title"
	KeySettingsLanguage     = "settings.language"
	KeyLanguageTitle        = "settings.language.title"
	KeyLanguageUkrainian    = "settings.language.uk"
	KeyLanguageEnglish      = "settings.language.en"
	KeyLanguageRussian      = "settings.language.ru"
	KeyLanguageChanged      = "settings.language.changed"
	KeyUnsupportedAction    = "error.unsupported_action"
	KeyMalformedCommand     = "error.malformed_command"
	KeyGenericError         = "error.generic"
	KeyPromptAddPerson      = "prompt.add_person"
	KeyPromptStartWork      = "prompt.start_work"
	KeyPromptStopWork       = "prompt.stop_work"
	KeyPromptReportMonth    = "prompt.report_month"
	KeyPersonAdded          = "person.added"
	KeyWorkStarted          = "work.started"
	KeyStatusInactive       = "status.inactive"
	KeyStatusActive         = "status.active"
	KeyStatusNoPeople       = "status.no_people"
	KeyWorkStopped          = "work.stopped"
	KeyWarning              = "warning"
	KeyAlertStopped         = "alert.stopped"
	KeyReason               = "reason"
	KeyReportDuplicate      = "report.duplicate"
	KeySetupComplete        = "group.setup.complete"
	KeySetupAlready         = "group.setup.already"
	KeySetupReEnabled       = "group.setup.re_enabled"
	KeySetupGroupOnly       = "group.setup.group_only"
	KeySetupRequired        = "group.setup.required"
	KeyGroupsHeader         = "group.list.header"
	KeyGroupsNone           = "group.list.none"
	KeyGroupsFallback       = "group.list.fallback"
	KeyGroupDisabled        = "group.disabled"
	KeyGroupAlreadyDisabled = "group.already_disabled"
	KeyGroupNoStored        = "group.no_stored"
	KeyCancel               = "flow.cancel"
	KeyBack                 = "flow.back"
	KeyFlowCancelled        = "flow.cancelled"
	KeyFlowExpired          = "flow.expired"
	KeyAddPersonEnterName   = "add_person.enter_name"
	KeyAddPersonInvalidName = "add_person.invalid_name"
	KeyStopWorkEnterInput   = "stop_work.enter_input"
	KeyStartWorkEnterPerson = "start_work.enter_person"
	KeyStartWorkEnterTitle  = "start_work.enter_title"
	KeyStartWorkSelectTime  = "start_work.select_time"
	KeyStartWorkEnterTime   = "start_work.enter_time"
	KeyStartWorkEnterManual = "start_work.enter_manual"
	KeyDateTimeInvalid      = "datetime.invalid"
	KeyDateTimeFuture       = "datetime.future"
	KeyStartNow             = "start_work.now"
	KeyStartToday           = "start_work.today"
	KeyStartYesterday       = "start_work.yesterday"
	KeyStartManual          = "start_work.manual"
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
		Ukrainian: "Команди:\n/setup\n/groups\n/disable_group\n/add_person <ім'я> <прізвище>\n/start_work <ім'я> <прізвище> \"<назва>\"\n/status\n/stop_work <ім'я> <прізвище> [причина]\n/report_month [YYYY-MM]\n\nТакож можна користуватися кнопками меню.",
		English:   "Commands:\n/setup\n/groups\n/disable_group\n/add_person <first_name> <last_name>\n/start_work <first_name> <last_name> \"<title>\"\n/status\n/stop_work <first_name> <last_name> [reason]\n/report_month [YYYY-MM]\n\nYou can also use the menu buttons.",
		Russian:   "Команды:\n/setup\n/groups\n/disable_group\n/add_person <имя> <фамилия>\n/start_work <имя> <фамилия> \"<название>\"\n/status\n/stop_work <имя> <фамилия> [причина]\n/report_month [YYYY-MM]\n\nТакже можно использовать кнопки меню.",
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
	KeySetupComplete: {
		Ukrainian: "Налаштування групи завершено.",
		English:   "Group setup is complete.",
		Russian:   "Настройка группы завершена.",
	},
	KeySetupAlready: {
		Ukrainian: "Налаштування групи вже було завершено. Назву групи оновлено.",
		English:   "Group setup was already complete. Group title was updated.",
		Russian:   "Настройка группы уже была завершена. Название группы обновлено.",
	},
	KeySetupReEnabled: {
		Ukrainian: "Групу знову увімкнено.",
		English:   "Group has been re-enabled.",
		Russian:   "Группа снова включена.",
	},
	KeySetupGroupOnly: {
		Ukrainian: "Команда /setup доступна лише в групових чатах.",
		English:   "/setup is available only in group chats.",
		Russian:   "Команда /setup доступна только в групповых чатах.",
	},
	KeySetupRequired: {
		Ukrainian: "Цю групу не налаштовано для бота. Попросіть відповідального виконати /setup.",
		English:   "This group is not set up for the bot. Ask an operator to run /setup.",
		Russian:   "Эта группа не настроена для бота. Попросите ответственного выполнить /setup.",
	},
	KeyGroupsHeader: {
		Ukrainian: "Налаштовані групи:",
		English:   "Configured groups:",
		Russian:   "Настроенные группы:",
	},
	KeyGroupsNone: {
		Ukrainian: "- немає збережених груп",
		English:   "- no stored groups",
		Russian:   "- сохраненных групп нет",
	},
	KeyGroupsFallback: {
		Ukrainian: "- Fallback group (%d): enabled",
		English:   "- Fallback group (%d): enabled",
		Russian:   "- Fallback group (%d): enabled",
	},
	KeyGroupDisabled: {
		Ukrainian: "Поточну групу вимкнено.",
		English:   "Current group has been disabled.",
		Russian:   "Текущая группа отключена.",
	},
	KeyGroupAlreadyDisabled: {
		Ukrainian: "Поточна група вже вимкнена.",
		English:   "Current group is already disabled.",
		Russian:   "Текущая группа уже отключена.",
	},
	KeyGroupNoStored: {
		Ukrainian: "Для поточної групи немає збереженого запису, який можна вимкнути.",
		English:   "There is no stored record for the current group to disable.",
		Russian:   "Для текущей группы нет сохраненной записи, которую можно отключить.",
	},
	KeyCancel: {
		Ukrainian: "Скасувати",
		English:   "Cancel",
		Russian:   "Отменить",
	},
	KeyBack: {
		Ukrainian: "Назад",
		English:   "Back",
		Russian:   "Назад",
	},
	KeyFlowCancelled: {
		Ukrainian: "Дію скасовано.",
		English:   "Action cancelled.",
		Russian:   "Действие отменено.",
	},
	KeyFlowExpired: {
		Ukrainian: "Час очікування минув. Почніть дію знову.",
		English:   "This flow expired. Start again.",
		Russian:   "Время ожидания истекло. Начните действие заново.",
	},
	KeyAddPersonEnterName: {
		Ukrainian: "Введіть ім'я та прізвище. Наприклад: Іван Петренко",
		English:   "Enter first name and last name. Example: Ivan Petrenko",
		Russian:   "Введите имя и фамилию. Например: Иван Петренко",
	},
	KeyAddPersonInvalidName: {
		Ukrainian: "Потрібні ім'я та прізвище. Наприклад: Іван Петренко",
		English:   "First name and last name are required. Example: Ivan Petrenko",
		Russian:   "Нужны имя и фамилия. Например: Иван Петренко",
	},
	KeyStopWorkEnterInput: {
		Ukrainian: "Введіть ім'я та прізвище, за потреби причину. Наприклад: Іван Петренко завершено",
		English:   "Enter first name and last name, with optional reason. Example: Ivan Petrenko done",
		Russian:   "Введите имя и фамилию, при необходимости причину. Например: Иван Петренко завершено",
	},
	KeyStartWorkEnterPerson: {
		Ukrainian: "Введіть ім'я та прізвище. Наприклад: Іван Петренко",
		English:   "Enter first name and last name. Example: Ivan Petrenko",
		Russian:   "Введите имя и фамилию. Например: Иван Петренко",
	},
	KeyStartWorkEnterTitle: {
		Ukrainian: "Введіть назву роботи.",
		English:   "Enter work title.",
		Russian:   "Введите название работы.",
	},
	KeyStartWorkSelectTime: {
		Ukrainian: "Як встановити дату та час початку?",
		English:   "How should the start date and time be set?",
		Russian:   "Как установить дату и время начала?",
	},
	KeyStartWorkEnterTime: {
		Ukrainian: "Введіть час у форматі HH:mm. Наприклад: 09:30",
		English:   "Enter time as HH:mm. Example: 09:30",
		Russian:   "Введите время в формате HH:mm. Например: 09:30",
	},
	KeyStartWorkEnterManual: {
		Ukrainian: "Введіть дату й час: DD.MM HH:mm, DD.MM.YYYY HH:mm або YYYY-MM-DD HH:mm.",
		English:   "Enter date and time: DD.MM HH:mm, DD.MM.YYYY HH:mm, or YYYY-MM-DD HH:mm.",
		Russian:   "Введите дату и время: DD.MM HH:mm, DD.MM.YYYY HH:mm или YYYY-MM-DD HH:mm.",
	},
	KeyDateTimeInvalid: {
		Ukrainian: "Неправильний формат дати/часу. Приклади: 08.06 09:30, 08.06.2026 09:30, 2026-06-08 09:30.",
		English:   "Invalid date/time format. Examples: 08.06 09:30, 08.06.2026 09:30, 2026-06-08 09:30.",
		Russian:   "Неверный формат даты/времени. Примеры: 08.06 09:30, 08.06.2026 09:30, 2026-06-08 09:30.",
	},
	KeyDateTimeFuture: {
		Ukrainian: "Дата й час початку не можуть бути в майбутньому.",
		English:   "Start date/time cannot be in the future.",
		Russian:   "Дата и время начала не могут быть в будущем.",
	},
	KeyStartNow: {
		Ukrainian: "Почати зараз",
		English:   "Start now",
		Russian:   "Начать сейчас",
	},
	KeyStartToday: {
		Ukrainian: "Сьогодні",
		English:   "Today",
		Russian:   "Сегодня",
	},
	KeyStartYesterday: {
		Ukrainian: "Вчора",
		English:   "Yesterday",
		Russian:   "Вчера",
	},
	KeyStartManual: {
		Ukrainian: "Ввести дату вручну",
		English:   "Enter date manually",
		Russian:   "Ввести дату вручную",
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
