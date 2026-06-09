package datetime

import (
	"errors"
	"testing"
	"time"
)

func TestParseStartFormats(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		input string
		mode  Mode
	}{
		{"dd.mm hh:mm", "08.06 14:30", ModeManual},
		{"dd.mm.yyyy hh:mm", "08.06.2026 14:30", ModeManual},
		{"yyyy-mm-dd hh:mm", "2026-06-08 14:30", ModeManual},
		{"today hh:mm", "14:30", ModeToday},
		{"yesterday hh:mm", "14:30", ModeYesterday},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseStart(tc.input, tc.mode, now)
			if err != nil {
				t.Fatal(err)
			}
			if got.Location() != time.UTC {
				t.Fatalf("not UTC: %s", got.Location())
			}
		})
	}
}

func TestParseStartRejectsFutureAndInvalid(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	if _, err := ParseStart("2099-01-01 10:00", ModeManual, now); !errors.Is(err, ErrFutureTime) {
		t.Fatalf("want future error, got %v", err)
	}
	if _, err := ParseStart("bad", ModeManual, now); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("want invalid format, got %v", err)
	}
}
