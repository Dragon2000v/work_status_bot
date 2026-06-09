package telegram

import (
	"bytes"
	"context"
	"errors"
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

type fullTelegram struct {
	messages     []string
	keyboards    []any
	answers      []string
	deletes      []int
	failAnswer   bool
	failDelete   bool
	deletedChats []int64
}

func (t *fullTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	t.messages = append(t.messages, text)
	return nil
}

func (t *fullTelegram) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard any) error {
	t.messages = append(t.messages, text)
	t.keyboards = append(t.keyboards, keyboard)
	return nil
}

func (t *fullTelegram) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	t.answers = append(t.answers, callbackQueryID)
	if t.failAnswer {
		return errors.New("answer failed")
	}
	return nil
}

func (t *fullTelegram) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	t.deletes = append(t.deletes, messageID)
	t.deletedChats = append(t.deletedChats, chatID)
	if t.failDelete {
		return errors.New("delete failed")
	}
	return nil
}

func newCallbackHandler(tg *fullTelegram) *Handler {
	personRepo := newTelegramPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	workRepo := &telegramWorkRepo{records: []works.WorkRecord{{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)}}}
	return NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, tg)
}

func TestCallbackRoutingStatusHelpAndPrompts(t *testing.T) {
	tg := &fullTelegram{}
	handler := newCallbackHandler(tg)
	for _, action := range []string{CallbackMenuStatus, CallbackMenuHelp, CallbackMenuAddPerson, CallbackMenuStartWork, CallbackMenuStopWork, CallbackMenuReportMonth} {
		text, _, outcome, err := handler.HandleCallback(context.Background(), action, 77, 1)
		if err != nil || outcome != "success" || text == "" {
			t.Fatalf("%s text=%q outcome=%s err=%v", action, text, outcome, err)
		}
	}
}

func TestCallbackWebhookAnswersSupportedAndUnsupported(t *testing.T) {
	tg := &fullTelegram{}
	handler := newCallbackHandler(tg)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"callback_query":{"id":"cb1","from":{"id":77},"message":{"message_id":9,"chat":{"id":10}},"data":"menu:help"}}`)))
	if res.Code != http.StatusOK || len(tg.answers) != 1 || len(tg.messages) != 1 {
		t.Fatalf("supported code=%d answers=%v messages=%v", res.Code, tg.answers, tg.messages)
	}
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"callback_query":{"id":"cb2","from":{"id":77},"message":{"message_id":9,"chat":{"id":10}},"data":"bad:data"}}`)))
	if res.Code != http.StatusOK || len(tg.answers) != 2 || !strings.Contains(tg.messages[1], "Невідома дія") {
		t.Fatalf("unsupported code=%d answers=%v messages=%v", res.Code, tg.answers, tg.messages)
	}
}

func TestCallbackAnswerFailureIsLogged(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	tg := &fullTelegram{failAnswer: true}
	handler := newCallbackHandler(tg)
	handler.logger = log
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"callback_query":{"id":"cb1","from":{"id":77},"message":{"message_id":9,"chat":{"id":10}},"data":"menu:help"}}`)))
	if res.Code != http.StatusOK || len(tg.messages) != 1 {
		t.Fatalf("code=%d messages=%v", res.Code, tg.messages)
	}
	if !strings.Contains(logs.String(), `"event":"telegram.callback_answer"`) || !strings.Contains(logs.String(), `"outcome":"failure"`) {
		t.Fatalf("missing answer failure log: %s", logs.String())
	}
}
