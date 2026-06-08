package telegram

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	applogger "work-status-bot/internal/logger"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type telegramPeopleRepo struct {
	people map[string]people.Person
}

func newTelegramPeopleRepo() *telegramPeopleRepo {
	return &telegramPeopleRepo{people: map[string]people.Person{}}
}

func (r *telegramPeopleRepo) Create(ctx context.Context, person people.Person) (people.Person, error) {
	r.people[person.NormalizedFullName] = person
	return person, nil
}

func (r *telegramPeopleRepo) FindByNormalizedName(ctx context.Context, normalized string) (people.Person, error) {
	person, ok := r.people[normalized]
	if !ok {
		return people.Person{}, people.ErrNotFound
	}
	return person, nil
}

func (r *telegramPeopleRepo) List(ctx context.Context) ([]people.Person, error) {
	var out []people.Person
	for _, person := range r.people {
		out = append(out, person)
	}
	return out, nil
}

type telegramWorkRepo struct {
	records []works.WorkRecord
}

func (r *telegramWorkRepo) CreateActive(ctx context.Context, record works.WorkRecord) (works.WorkRecord, error) {
	r.records = append(r.records, record)
	return record, nil
}

func (r *telegramWorkRepo) FindActiveByPerson(ctx context.Context, personID primitive.ObjectID) (works.WorkRecord, error) {
	for _, rec := range r.records {
		if rec.PersonID == personID && rec.Status == works.StatusActive {
			return rec, nil
		}
	}
	return works.WorkRecord{}, works.ErrNoActiveWork
}

func (r *telegramWorkRepo) StopActive(ctx context.Context, personID primitive.ObjectID, stoppedAt time.Time, reason string) (works.WorkRecord, error) {
	for i := range r.records {
		if r.records[i].PersonID == personID && r.records[i].Status == works.StatusActive {
			r.records[i].Status = works.StatusStopped
			r.records[i].StoppedAt = &stoppedAt
			r.records[i].StopReason = reason
			return r.records[i], nil
		}
	}
	return works.WorkRecord{}, works.ErrNoActiveWork
}

func (r *telegramWorkRepo) ListStatus(ctx context.Context) ([]works.WorkRecord, error) {
	return r.records, nil
}

func (r *telegramWorkRepo) ListOverlappingMonth(ctx context.Context, startUTC, endUTC time.Time) ([]works.WorkRecord, error) {
	return r.records, nil
}

type fakePeopleService struct{ repo *telegramPeopleRepo }

func (s fakePeopleService) Add(ctx context.Context, firstName, lastName string) (people.Person, error) {
	return people.NewService(s.repo).Add(ctx, firstName, lastName)
}

type fakeWorkService struct {
	repo   *telegramWorkRepo
	people *telegramPeopleRepo
	now    time.Time
}

func (s fakeWorkService) Start(ctx context.Context, firstName, lastName, title string) (people.Person, works.WorkRecord, error) {
	svc := works.NewService(s.people, s.repo)
	svc.SetNow(func() time.Time { return s.now })
	return svc.Start(ctx, firstName, lastName, title)
}

func (s fakeWorkService) Status(ctx context.Context) ([]people.Person, []works.WorkRecord, error) {
	persons, err := s.people.List(ctx)
	return persons, s.repo.records, err
}

func (s fakeWorkService) Stop(ctx context.Context, firstName, lastName, reason string) (people.Person, works.WorkRecord, error) {
	svc := works.NewService(s.people, s.repo)
	svc.SetNow(func() time.Time { return s.now.Add(time.Hour) })
	return svc.Stop(ctx, firstName, lastName, reason)
}

func TestHandlerHappyPaths(t *testing.T) {
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	handler := NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, now}, nil, nil)
	cases := []string{"/help", "/add_person Ivan Petrenko", "/start_work Ivan Petrenko \"API fix\"", "/status", "/stop_work Ivan Petrenko done"}
	for _, text := range cases {
		got, err := handler.HandleText(context.Background(), text)
		if err != nil {
			t.Fatalf("%s err: %v", text, err)
		}
		if got == "" || strings.HasPrefix(got, "Error:") && text != "/stop_work Ivan Petrenko done" {
			t.Fatalf("%s bad response: %q", text, got)
		}
	}
}

func TestHandlerErrors(t *testing.T) {
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	handler := NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, nil)
	for _, text := range []string{"/add_person Ivan", "/start_work Ivan Petrenko", "/bogus", "/stop_work Ivan Petrenko"} {
		got, err := handler.HandleText(context.Background(), text)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !strings.HasPrefix(got, "Error:") {
			t.Fatalf("want error for %s, got %q", text, got)
		}
	}
	_, _ = personRepo.Create(context.Background(), people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")})
	if _, err := handler.HandleText(context.Background(), "/start_work Ivan Petrenko \"API\""); err != nil {
		t.Fatal(err)
	}
	got, _ := handler.HandleText(context.Background(), "/start_work Ivan Petrenko \"Other\"")
	if !strings.Contains(got, "active") {
		t.Fatalf("want already active, got %q", got)
	}
}

type failTelegram struct{}

func (failTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	return errors.New("send failed")
}

type telegramReportsRepo struct {
	incidents []reports.Incident
	reports   map[string]reports.MonthlyReport
}

func newTelegramReportsRepo() *telegramReportsRepo {
	return &telegramReportsRepo{reports: map[string]reports.MonthlyReport{}}
}

func (r *telegramReportsRepo) CreateIncident(ctx context.Context, incident reports.Incident) error {
	r.incidents = append(r.incidents, incident)
	return nil
}

func (r *telegramReportsRepo) ListIncidentsByMonth(ctx context.Context, startUTC, endUTC time.Time) ([]reports.Incident, error) {
	return r.incidents, nil
}

func (r *telegramReportsRepo) CreateMonthlyReport(ctx context.Context, report reports.MonthlyReport) (reports.MonthlyReport, error) {
	if _, ok := r.reports[report.Month]; ok {
		return reports.MonthlyReport{}, reports.ErrReportExists
	}
	r.reports[report.Month] = report
	return report, nil
}

func (r *telegramReportsRepo) FindMonthlyReport(ctx context.Context, month string) (reports.MonthlyReport, error) {
	report, ok := r.reports[month]
	if !ok {
		return reports.MonthlyReport{}, reports.ErrReportNotFound
	}
	return report, nil
}

type telegramReportWorkLister struct{}

func (telegramReportWorkLister) ListOverlappingMonth(ctx context.Context, month string) ([]works.WorkRecord, error) {
	return nil, nil
}

func TestHandlerStopAlertWarning(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	personRepo := newTelegramPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	workRepo := &telegramWorkRepo{records: []works.WorkRecord{{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: now}}}
	reportRepo := newTelegramReportsRepo()
	reportSvc := reports.NewService(reportRepo, personRepo, telegramReportWorkLister{}, nil, 0)
	handler := NewHandlerWithLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, now}, reportSvc, failTelegram{}, log)
	got, err := handler.HandleText(context.Background(), "/stop_work Ivan Petrenko blocked")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Warning") || len(reportRepo.incidents) != 2 {
		t.Fatalf("want warning and incidents, got %q %#v", got, reportRepo.incidents)
	}
	if !strings.Contains(logs.String(), `"event":"telegram.alert_send"`) || !strings.Contains(logs.String(), `"outcome":"failure"`) {
		t.Fatalf("alert failure log missing: %s", logs.String())
	}
}
