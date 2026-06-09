package reports

import (
	"context"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/groups"
	"work-status-bot/internal/people"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memReportsRepo struct {
	incidents []Incident
	reports   map[string]MonthlyReport
}

func newMemReportsRepo() *memReportsRepo { return &memReportsRepo{reports: map[string]MonthlyReport{}} }

type reportPeopleRepo struct {
	people map[string]people.Person
}

func newReportPeopleRepo() *reportPeopleRepo {
	return &reportPeopleRepo{people: map[string]people.Person{}}
}

func (r *reportPeopleRepo) Create(ctx context.Context, person people.Person) (people.Person, error) {
	r.people[person.NormalizedFullName] = person
	return person, nil
}

func (r *reportPeopleRepo) FindByNormalizedName(ctx context.Context, normalized string) (people.Person, error) {
	p, ok := r.people[normalized]
	if !ok {
		return people.Person{}, people.ErrNotFound
	}
	return p, nil
}

func (r *reportPeopleRepo) List(ctx context.Context) ([]people.Person, error) {
	var out []people.Person
	for _, p := range r.people {
		out = append(out, p)
	}
	return out, nil
}

func (r *memReportsRepo) CreateIncident(ctx context.Context, incident Incident) error {
	r.incidents = append(r.incidents, incident)
	return nil
}

func (r *memReportsRepo) ListIncidentsByMonth(ctx context.Context, startUTC, endUTC time.Time) ([]Incident, error) {
	var out []Incident
	for _, inc := range r.incidents {
		if !inc.OccurredAt.Before(startUTC) && inc.OccurredAt.Before(endUTC) {
			out = append(out, inc)
		}
	}
	return out, nil
}

func (r *memReportsRepo) CreateMonthlyReport(ctx context.Context, report MonthlyReport) (MonthlyReport, error) {
	if _, ok := r.reports[report.Month]; ok {
		return MonthlyReport{}, ErrReportExists
	}
	r.reports[report.Month] = report
	return report, nil
}

func (r *memReportsRepo) FindMonthlyReport(ctx context.Context, month string) (MonthlyReport, error) {
	report, ok := r.reports[month]
	if !ok {
		return MonthlyReport{}, ErrReportNotFound
	}
	return report, nil
}

type reportWorkLister struct{ records []works.WorkRecord }

type recordingTelegram struct {
	chats []int64
	texts []string
}

func (t *recordingTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	t.chats = append(t.chats, chatID)
	t.texts = append(t.texts, text)
	return nil
}

func (l reportWorkLister) ListOverlappingMonth(ctx context.Context, month string) ([]works.WorkRecord, error) {
	start, end, _ := works.MonthBoundsKyiv(month)
	var out []works.WorkRecord
	for _, rec := range l.records {
		stopped := end
		if rec.StoppedAt != nil {
			stopped = *rec.StoppedAt
		}
		if rec.StartedAt.Before(end) && !stopped.Before(start) {
			out = append(out, rec)
		}
	}
	return out, nil
}

func TestReportsServiceIncidents(t *testing.T) {
	repo := newMemReportsRepo()
	svc := NewService(repo, newReportPeopleRepo(), reportWorkLister{}, nil, 0)
	personID := primitive.NewObjectID()
	recordID := primitive.NewObjectID()
	when := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	if err := svc.RecordStopped(context.Background(), personID, recordID, "done", when); err != nil {
		t.Fatal(err)
	}
	if err := svc.RecordAlertFailed(context.Background(), personID, recordID, "send failed", when); err != nil {
		t.Fatal(err)
	}
	if len(repo.incidents) != 2 || repo.incidents[0].Type != IncidentStopped || repo.incidents[1].Type != IncidentAlertFailed {
		t.Fatalf("bad incidents: %#v", repo.incidents)
	}
}

func TestReportsServiceMonthlyOverlapAndDuplicate(t *testing.T) {
	repo := newMemReportsRepo()
	personRepo := newReportPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	stop := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	records := []works.WorkRecord{
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "before", Status: works.StatusStopped, StartedAt: time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC), StoppedAt: &stop},
		{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "during", Status: works.StatusActive, StartedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)},
	}
	svc := NewService(repo, personRepo, reportWorkLister{records: records}, nil, 0)
	svc.now = func() time.Time { return time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC) }
	result, err := svc.GenerateMonthly(context.Background(), "2026-06", TriggerTelegramCommand)
	if err != nil {
		t.Fatal(err)
	}
	if result.Duplicate || result.Report.Month != "2026-06" {
		t.Fatalf("bad report result")
	}
	if result.Report.Content == "" || !contains(result.Report.Content, "before") || !contains(result.Report.Content, "during") {
		t.Fatalf("missing overlapping records: %s", result.Report.Content)
	}
	dupe, err := svc.GenerateMonthly(context.Background(), "2026-06", TriggerTelegramCommand)
	if err != nil {
		t.Fatal(err)
	}
	if !dupe.Duplicate {
		t.Fatalf("want duplicate")
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func TestReportsServiceSendsReportToChatAndTargets(t *testing.T) {
	tg := &recordingTelegram{}
	svc := NewService(newMemReportsRepo(), newReportPeopleRepo(), reportWorkLister{}, tg, 10)
	report := MonthlyReport{Month: "2026-06", Content: "report body"}
	if err := svc.SendReportToChat(context.Background(), -1001, report, false); err != nil {
		t.Fatal(err)
	}
	if len(tg.chats) != 1 || tg.chats[0] != -1001 || tg.texts[0] != "report body" {
		t.Fatalf("manual report target wrong: chats=%v texts=%v", tg.chats, tg.texts)
	}
	targets := []groups.ReportDeliveryTarget{{TelegramChatID: -1002}, {TelegramChatID: 10}}
	if err := svc.SendReportToTargets(context.Background(), report, true, targets); err != nil {
		t.Fatal(err)
	}
	if len(tg.chats) != 3 || tg.chats[1] != -1002 || tg.chats[2] != 10 {
		t.Fatalf("fan-out targets wrong: chats=%v", tg.chats)
	}
	if !strings.Contains(tg.texts[1], "Звіт уже існує") {
		t.Fatalf("duplicate prefix missing: %q", tg.texts[1])
	}
}
