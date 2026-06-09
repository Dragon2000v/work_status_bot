package telegram

import (
	"errors"
	"strings"
)

const (
	CommandHelp         = "/help"
	CommandAddPerson    = "/add_person"
	CommandStartWork    = "/start_work"
	CommandStatus       = "/status"
	CommandStopWork     = "/stop_work"
	CommandReportMonth  = "/report_month"
	CommandSetup        = "/setup"
	CommandGroups       = "/groups"
	CommandDisableGroup = "/disable_group"
)

var ErrMalformedCommand = errors.New("malformed command")

type Command struct {
	Name      string
	FirstName string
	LastName  string
	Title     string
	Reason    string
	Month     string
}

func ParseCommand(text string) (Command, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Command{}, ErrMalformedCommand
	}
	fields := splitCommand(text)
	if len(fields) == 0 {
		return Command{}, ErrMalformedCommand
	}
	cmd := Command{Name: fields[0]}
	switch cmd.Name {
	case CommandHelp, CommandStatus, CommandSetup, CommandGroups, CommandDisableGroup:
		return cmd, nil
	case CommandAddPerson:
		if len(fields) != 3 {
			return Command{}, ErrMalformedCommand
		}
		cmd.FirstName, cmd.LastName = fields[1], fields[2]
	case CommandStartWork:
		if len(fields) < 4 {
			return Command{}, ErrMalformedCommand
		}
		cmd.FirstName, cmd.LastName, cmd.Title = fields[1], fields[2], strings.Join(fields[3:], " ")
	case CommandStopWork:
		if len(fields) < 3 {
			return Command{}, ErrMalformedCommand
		}
		cmd.FirstName, cmd.LastName = fields[1], fields[2]
		if len(fields) > 3 {
			cmd.Reason = strings.Join(fields[3:], " ")
		}
	case CommandReportMonth:
		if len(fields) > 2 {
			return Command{}, ErrMalformedCommand
		}
		if len(fields) == 2 {
			cmd.Month = fields[1]
		}
	default:
		return Command{}, ErrMalformedCommand
	}
	return cmd, nil
}

func splitCommand(text string) []string {
	var out []string
	var b strings.Builder
	inQuote := false
	for _, r := range text {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' || r == '\t' || r == '\n':
			if inQuote {
				b.WriteRune(r)
			} else if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}
