package datetime

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidFormat = errors.New("invalid date/time format")
	ErrFutureTime    = errors.New("start date/time cannot be in the future")
)

type Mode string

const (
	ModeNow       Mode = "now"
	ModeToday     Mode = "today"
	ModeYesterday Mode = "yesterday"
	ModeManual    Mode = "manual"
)

func ParseStart(input string, mode Mode, now time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		return time.Time{}, err
	}
	nowLocal := now.In(loc)
	input = strings.TrimSpace(input)
	var local time.Time
	switch mode {
	case ModeNow:
		local = nowLocal
	case ModeToday, ModeYesterday:
		local, err = parseClock(input, nowLocal, loc)
		if err == nil && mode == ModeYesterday {
			local = local.AddDate(0, 0, -1)
		}
	case ModeManual:
		local, err = parseManual(input, nowLocal, loc)
	default:
		err = ErrInvalidFormat
	}
	if err != nil {
		return time.Time{}, err
	}
	if local.After(nowLocal) {
		return time.Time{}, ErrFutureTime
	}
	return local.UTC(), nil
}

func parseClock(input string, nowLocal time.Time, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation("15:04", input, loc)
	if err != nil {
		return time.Time{}, ErrInvalidFormat
	}
	return time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
}

func parseManual(input string, nowLocal time.Time, loc *time.Location) (time.Time, error) {
	layouts := []string{"02.01.2006 15:04", "2006-01-02 15:04"}
	if strings.Count(input, ".") == 1 {
		parts := strings.SplitN(input, " ", 2)
		if len(parts) != 2 {
			return time.Time{}, ErrInvalidFormat
		}
		input = parts[0] + "." + strconv.Itoa(nowLocal.Year()) + " " + parts[1]
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, input, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalidFormat
}
