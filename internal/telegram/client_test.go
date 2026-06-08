package telegram

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	applogger "work-status-bot/internal/logger"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientSendMessageSuccessAndFailure(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	calls := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
		}
		return &http.Response{StatusCode: http.StatusBadGateway, Status: "502 Bad Gateway", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
	})}
	client := NewClientWithLogger("secret-token", httpClient, log)
	client.baseURL = "https://telegram.test"
	if err := client.SendMessage(context.Background(), 1, "ok"); err != nil {
		t.Fatalf("send ok: %v", err)
	}
	if err := client.SendMessage(context.Background(), 1, "fail"); err == nil {
		t.Fatalf("want failure")
	}
	out := logs.String()
	for _, want := range []string{`"event":"telegram.send"`, `"outcome":"success"`, `"outcome":"failure"`, `"chat_id":1`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, "secret-token") || strings.Contains(out, "telegram.test") || strings.Contains(out, "sendMessage") {
		t.Fatalf("secret url leaked: %s", out)
	}
}

func TestClientSetWebhookSuccessAndFailure(t *testing.T) {
	var logs bytes.Buffer
	log, _ := applogger.New(&logs, applogger.EnvProduction, "info")
	calls := 0
	var bodies []string
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		if calls == 1 {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
		}
		return &http.Response{StatusCode: http.StatusBadGateway, Status: "502 Bad Gateway", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
	})}
	client := NewClientWithLogger("webhook-secret-token", httpClient, log)
	client.baseURL = "https://telegram.test"
	if err := client.SetWebhook(context.Background(), "https://example.com/telegram/webhook", "super-webhook-secret"); err != nil {
		t.Fatalf("set webhook ok: %v", err)
	}
	if err := client.SetWebhook(context.Background(), "https://example.com/telegram/webhook", "super-webhook-secret"); err == nil {
		t.Fatalf("want set webhook failure")
	}
	if !strings.Contains(bodies[0], `"secret_token":"super-webhook-secret"`) {
		t.Fatalf("secret token not sent to Telegram: %s", bodies[0])
	}
	out := logs.String()
	for _, want := range []string{`"event":"telegram.webhook_registration"`, `"outcome":"attempt"`, `"outcome":"success"`, `"outcome":"failure"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	for _, secret := range []string{"webhook-secret-token", "super-webhook-secret", "telegram.test", "setWebhook"} {
		if strings.Contains(out, secret) {
			t.Fatalf("webhook secret leaked: %s", out)
		}
	}
}
