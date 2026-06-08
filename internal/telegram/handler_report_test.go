package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
)

func TestHandlerReportMonth(t *testing.T) {
	personRepo := newTelegramPeopleRepo()
	_, _ = personRepo.Create(context.Background(), people.Person{FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")})
	reportSvc := reports.NewService(newTelegramReportsRepo(), personRepo, telegramReportWorkLister{}, nil, 0)
	reportSvc.SetNow(func() time.Time { return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC) })
	handler := NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{&telegramWorkRepo{}, personRepo, time.Now().UTC()}, reportSvc, nil)
	got, err := handler.HandleText(context.Background(), "/report_month")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "2026-06") {
		t.Fatalf("default month missing: %q", got)
	}
	got, _ = handler.HandleText(context.Background(), "/report_month 2026-06")
	if !strings.Contains(got, "already exists") {
		t.Fatalf("duplicate missing: %q", got)
	}
	got, _ = handler.HandleText(context.Background(), "/report_month bad")
	if !strings.HasPrefix(got, "Error:") {
		t.Fatalf("want invalid month error, got %q", got)
	}
}
