package telegram

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	applogger "work-status-bot/internal/logger"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type okTelegram struct {
	messages []string
}

func (t *okTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	t.messages = append(t.messages, text)
	return nil
}

func TestHandlerStopAlertSuccess(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	personRepo := newTelegramPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	workRepo := &telegramWorkRepo{records: []works.WorkRecord{{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: now}}}
	reportRepo := newTelegramReportsRepo()
	reportSvc := reports.NewService(reportRepo, personRepo, telegramReportWorkLister{}, nil, 0)
	tg := &okTelegram{}
	handler := NewHandlerWithLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, now}, reportSvc, tg, log)
	got, err := handler.HandleText(context.Background(), "/stop_work Ivan Petrenko done")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "зупинено") || strings.Contains(got, "Попередження") {
		t.Fatalf("bad stop response: %q", got)
	}
	if len(tg.messages) != 1 || !strings.Contains(tg.messages[0], "Увага:") {
		t.Fatalf("alert not sent: %#v", tg.messages)
	}
	if len(reportRepo.incidents) != 1 || reportRepo.incidents[0].Type != reports.IncidentStopped {
		t.Fatalf("stopped incident missing: %#v", reportRepo.incidents)
	}
	if !strings.Contains(logs.String(), `"event":"telegram.alert_send"`) || !strings.Contains(logs.String(), `"outcome":"success"`) {
		t.Fatalf("alert success log missing: %s", logs.String())
	}
}
