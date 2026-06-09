package telegram

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookAllowsHelpButRejectsOperationalUnknownChat(t *testing.T) {
	handler := NewHandler(10, nil, nil, nil, nil)
	good := httptest.NewRecorder()
	handler.ServeHTTP(good, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"text":"/help","chat":{"id":10}}}`)))
	if good.Code != http.StatusOK {
		t.Fatalf("configured chat status %d", good.Code)
	}
	help := httptest.NewRecorder()
	handler.ServeHTTP(help, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"text":"/help","chat":{"id":11}}}`)))
	if help.Code != http.StatusOK {
		t.Fatalf("unknown help status %d", help.Code)
	}
	bad := httptest.NewRecorder()
	handler.ServeHTTP(bad, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"message":{"text":"/status","chat":{"id":11}}}`)))
	if bad.Code != http.StatusOK {
		t.Fatalf("unconfigured operational status %d", bad.Code)
	}
}
