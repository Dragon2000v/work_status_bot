package telegram

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"work-status-bot/internal/datetime"
	"work-status-bot/internal/groups"
	"work-status-bot/internal/i18n"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"
)

func HelpMessage() string {
	return HelpMessageLang(i18n.Ukrainian)
}

func HelpMessageLang(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeyHelpText)
}

func MenuMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeyMenuTitle)
}

func SettingsMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeySettingsTitle)
}

func LanguageMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeyLanguageTitle)
}

func SetupMessage(lang i18n.Language, outcome string) string {
	switch outcome {
	case groups.OutcomeUpdated:
		return i18n.T(lang, i18n.KeySetupAlready)
	case groups.OutcomeReEnabled:
		return i18n.T(lang, i18n.KeySetupReEnabled)
	default:
		return i18n.T(lang, i18n.KeySetupComplete)
	}
}

func SetupGroupOnlyMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeySetupGroupOnly)
}

func SetupRequiredMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeySetupRequired)
}

func GroupsMessage(lang i18n.Language, configured []groups.ConfiguredGroup, fallbackChat int64) string {
	var lines []string
	lines = append(lines, i18n.T(lang, i18n.KeyGroupsHeader))
	if len(configured) == 0 {
		lines = append(lines, i18n.T(lang, i18n.KeyGroupsNone))
		if fallbackChat != 0 {
			lines = append(lines, fmt.Sprintf(i18n.T(lang, i18n.KeyGroupsFallback), fallbackChat))
		}
		return strings.Join(lines, "\n")
	}
	for _, group := range configured {
		status := "disabled"
		if group.Enabled {
			status = "enabled"
		}
		lines = append(lines, fmt.Sprintf("- %s (%d): %s", group.Title, group.TelegramChatID, status))
	}
	return strings.Join(lines, "\n")
}

func DisableGroupMessage(lang i18n.Language, outcome string) string {
	if outcome == groups.OutcomeAlreadyDisabled {
		return i18n.T(lang, i18n.KeyGroupAlreadyDisabled)
	}
	return i18n.T(lang, i18n.KeyGroupDisabled)
}

func NoStoredGroupMessage(lang i18n.Language) string {
	return i18n.T(lang, i18n.KeyGroupNoStored)
}

func ActionPromptMessage(lang i18n.Language, action string) string {
	switch action {
	case CallbackMenuAddPerson:
		return i18n.T(lang, i18n.KeyPromptAddPerson)
	case CallbackMenuStartWork:
		return i18n.T(lang, i18n.KeyPromptStartWork)
	case CallbackMenuStopWork:
		return i18n.T(lang, i18n.KeyPromptStopWork)
	case CallbackMenuReportMonth:
		return i18n.T(lang, i18n.KeyPromptReportMonth)
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction)
	}
}

func PersonAddedMessage(p people.Person) string {
	return PersonAddedMessageLang(i18n.Ukrainian, p)
}

func PersonAddedMessageLang(lang i18n.Language, p people.Person) string {
	return fmt.Sprintf(i18n.T(lang, i18n.KeyPersonAdded), p.FirstName, p.LastName)
}

func WorkStartedMessage(p people.Person, r works.WorkRecord) string {
	return WorkStartedMessageLang(i18n.Ukrainian, p, r)
}

func WorkStartedMessageLang(lang i18n.Language, p people.Person, r works.WorkRecord) string {
	return fmt.Sprintf(i18n.T(lang, i18n.KeyWorkStarted), p.FirstName, p.LastName, r.Title, works.FormatKyiv(r.StartedAt))
}

func StatusMessage(persons []people.Person, records []works.WorkRecord, now time.Time) string {
	return StatusMessageLang(i18n.Ukrainian, persons, records, now)
}

func StatusMessageLang(lang i18n.Language, persons []people.Person, records []works.WorkRecord, now time.Time) string {
	active := map[string]works.WorkRecord{}
	for _, r := range records {
		if r.Status == works.StatusActive {
			active[r.PersonID.Hex()] = r
		}
	}
	var lines []string
	for _, p := range persons {
		r, ok := active[p.ID.Hex()]
		if !ok {
			lines = append(lines, fmt.Sprintf(i18n.T(lang, i18n.KeyStatusInactive), p.FirstName, p.LastName))
			continue
		}
		lines = append(lines, fmt.Sprintf(i18n.T(lang, i18n.KeyStatusActive), p.FirstName, p.LastName, r.Title, works.FormatKyiv(r.StartedAt), works.FormatDuration(works.Elapsed(r, now))))
	}
	if len(lines) == 0 {
		return i18n.T(lang, i18n.KeyStatusNoPeople)
	}
	return strings.Join(lines, "\n")
}

func WorkStoppedMessage(p people.Person, r works.WorkRecord, warning string) string {
	return WorkStoppedMessageLang(i18n.Ukrainian, p, r, warning)
}

func WorkStoppedMessageLang(lang i18n.Language, p people.Person, r works.WorkRecord, warning string) string {
	stopped := ""
	if r.StoppedAt != nil {
		stopped = works.FormatKyiv(*r.StoppedAt)
	}
	msg := fmt.Sprintf(i18n.T(lang, i18n.KeyWorkStopped), p.FirstName, p.LastName, r.Title, stopped)
	if warning != "" {
		msg += "\n" + i18n.T(lang, i18n.KeyWarning) + ": " + warning
	}
	return msg
}

func StopAlertMessage(p people.Person, r works.WorkRecord) string {
	reason := ""
	if r.StopReason != "" {
		reason = " " + i18n.T(i18n.Ukrainian, i18n.KeyReason) + ": " + r.StopReason
	}
	return fmt.Sprintf(i18n.T(i18n.Ukrainian, i18n.KeyAlertStopped), p.FirstName, p.LastName, r.Title) + reason
}

func ReportMessage(result reports.GenerateResult) string {
	if result.Duplicate {
		return i18n.T(i18n.Ukrainian, i18n.KeyReportDuplicate) + "\n" + result.Report.Content
	}
	return result.Report.Content
}

func ErrorMessage(err error) string {
	return ErrorMessageLang(i18n.Ukrainian, err)
}

func ErrorMessageLang(lang i18n.Language, err error) string {
	if err == nil {
		return ""
	}
	if err == ErrMalformedCommand {
		return i18n.T(lang, i18n.KeyGenericError) + ": " + i18n.T(lang, i18n.KeyMalformedCommand)
	}
	return i18n.T(lang, i18n.KeyGenericError) + ": " + err.Error()
}

func DateTimeErrorMessage(lang i18n.Language, err error) string {
	if errors.Is(err, datetime.ErrFutureTime) {
		return i18n.T(lang, i18n.KeyDateTimeFuture)
	}
	return i18n.T(lang, i18n.KeyDateTimeInvalid)
}
