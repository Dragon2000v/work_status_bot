package telegram

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	applogger "work-status-bot/internal/logger"
	"work-status-bot/internal/people"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTelegramWebhookLogging(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	handler := NewHandlerWithLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, nil, log)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"text":"/help","from":{"id":77},"chat":{"id":10}}}`)))
	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
	}

	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"text":"/status","from":{"id":78},"chat":{"id":11}}}`)))
	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
	}

	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad json`)))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status %d", res.Code)
	}

	out := buf.String()
	for _, want := range []string{`"event":"telegram.webhook_received"`, `"event":"telegram.group_authorization"`, `"chat_id":10`, `"chat_id":11`, `"user_id":77`, `"command":"/help"`, `"command":"/status"`, `"outcome":"rejected"`, `"outcome":"invalid"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, `"text"`) || strings.Contains(out, "bot-token") {
		t.Fatalf("raw payload leaked: %s", out)
	}
}

func TestTelegramCommandLogging(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	personRepo := newTelegramPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	workRepo := &telegramWorkRepo{records: []works.WorkRecord{{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: time.Now().UTC()}}}
	handler := NewHandlerWithLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, nil, log)

	if _, err := handler.HandleText(context.Background(), "/status"); err != nil {
		t.Fatal(err)
	}
	if _, err := handler.HandleText(context.Background(), "/unknown"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"event":"telegram.command"`, `"event":"telegram.command_result"`, `"command":"/status"`, `"command":"/unknown"`, `"outcome":"success"`, `"outcome":"invalid"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, "API") {
		t.Fatalf("message detail leaked: %s", out)
	}
}
