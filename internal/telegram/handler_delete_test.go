package telegram

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	applogger "work-status-bot/internal/logger"
)

func TestCommandDeleteBestEffort(t *testing.T) {
	tg := &fullTelegram{}
	handler := newCallbackHandler(tg)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"message_id":55,"text":"/help","from":{"id":77},"chat":{"id":10}}}`)))
	if res.Code != http.StatusOK || len(tg.deletes) != 1 || tg.deletes[0] != 55 {
		t.Fatalf("code=%d deletes=%v", res.Code, tg.deletes)
	}
	if len(tg.messages) == 0 || len(tg.deletes) != 1 {
		t.Fatalf("bot response should not be deleted: messages=%v deletes=%v", tg.messages, tg.deletes)
	}
}

func TestCommandDeleteFailureLoggedAndCommandSucceeds(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	tg := &fullTelegram{failDelete: true}
	handler := newCallbackHandler(tg)
	handler.logger = log
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"message_id":55,"text":"/help","from":{"id":77},"chat":{"id":10}}}`)))
	if res.Code != http.StatusOK {
		t.Fatalf("code=%d", res.Code)
	}
	out := logs.String()
	if !strings.Contains(out, `"event":"telegram.command_delete"`) || !strings.Contains(out, `"outcome":"failure"`) {
		t.Fatalf("missing delete failure log: %s", out)
	}
}
