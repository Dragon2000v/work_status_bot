package telegram

import (
	"fmt"
	"strings"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"
)

func HelpMessage() string {
	return strings.Join([]string{
		"/add_person <first_name> <last_name>",
		"/start_work <first_name> <last_name> \"<title>\"",
		"/status",
		"/stop_work <first_name> <last_name> [reason]",
		"/report_month [YYYY-MM]",
	}, "\n")
}

func PersonAddedMessage(p people.Person) string {
	return fmt.Sprintf("Added %s %s.", p.FirstName, p.LastName)
}

func WorkStartedMessage(p people.Person, r works.WorkRecord) string {
	return fmt.Sprintf("Started work for %s %s: %s at %s.", p.FirstName, p.LastName, r.Title, works.FormatKyiv(r.StartedAt))
}

func StatusMessage(persons []people.Person, records []works.WorkRecord, now time.Time) string {
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
			lines = append(lines, fmt.Sprintf("%s %s: inactive", p.FirstName, p.LastName))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s: %s, active since %s, elapsed %s", p.FirstName, p.LastName, r.Title, works.FormatKyiv(r.StartedAt), works.FormatDuration(works.Elapsed(r, now))))
	}
	if len(lines) == 0 {
		return "No people tracked."
	}
	return strings.Join(lines, "\n")
}

func WorkStoppedMessage(p people.Person, r works.WorkRecord, warning string) string {
	stopped := ""
	if r.StoppedAt != nil {
		stopped = works.FormatKyiv(*r.StoppedAt)
	}
	msg := fmt.Sprintf("Stopped work for %s %s: %s at %s.", p.FirstName, p.LastName, r.Title, stopped)
	if warning != "" {
		msg += "\nWarning: " + warning
	}
	return msg
}

func StopAlertMessage(p people.Person, r works.WorkRecord) string {
	reason := ""
	if r.StopReason != "" {
		reason = " Reason: " + r.StopReason
	}
	return fmt.Sprintf("Alert: %s %s stopped work \"%s\".%s", p.FirstName, p.LastName, r.Title, reason)
}

func ReportMessage(result reports.GenerateResult) string {
	if result.Duplicate {
		return "Report already exists.\n" + result.Report.Content
	}
	return result.Report.Content
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return "Error: " + err.Error()
}
