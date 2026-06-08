package reports

import (
	"context"
	"testing"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestReportsServiceMonthOverlapCases(t *testing.T) {
	repo := newMemReportsRepo()
	personRepo := newReportPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	stopInMonth := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	stopAfterMonth := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	records := []works.WorkRecord{
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "active", Status: works.StatusActive, StartedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)},
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "started-in-month", Status: works.StatusStopped, StartedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC), StoppedAt: &stopAfterMonth},
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "stopped-in-month", Status: works.StatusStopped, StartedAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), StoppedAt: &stopInMonth},
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "cross-month", Status: works.StatusStopped, StartedAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), StoppedAt: &stopAfterMonth},
	}
	svc := NewService(repo, personRepo, reportWorkLister{records: records}, nil, 0)
	svc.SetNow(func() time.Time { return time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC) })
	result, err := svc.GenerateMonthly(context.Background(), "2026-06", TriggerTelegramCommand)
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"active", "started-in-month", "stopped-in-month", "cross-month"} {
		if !contains(result.Report.Content, title) {
			t.Fatalf("missing %s in report: %s", title, result.Report.Content)
		}
	}
}
