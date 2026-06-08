package reports

import (
	"context"
	"errors"
	"testing"
)

func TestMonthlyReportDuplicateBehavior(t *testing.T) {
	repo := newMemReportsRepo()
	report := MonthlyReport{Month: "2026-06", Content: "x"}
	if _, err := repo.CreateMonthlyReport(context.Background(), report); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateMonthlyReport(context.Background(), report); !errors.Is(err, ErrReportExists) {
		t.Fatalf("want duplicate, got %v", err)
	}
}
