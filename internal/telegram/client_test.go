package telegram

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestClientAdditionalBotAPIPayloads(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
	})}
	client := NewClient("secret-token", httpClient)
	client.baseURL = "https://telegram.test"
	keyboard := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{{Text: "Статус", CallbackData: CallbackMenuStatus}}}}
	if err := client.SendMessageWithKeyboard(context.Background(), 10, "menu", keyboard); err != nil {
		t.Fatal(err)
	}
	if err := client.AnswerCallbackQuery(context.Background(), "cb1", ""); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteMessage(context.Background(), 10, 55); err != nil {
		t.Fatal(err)
	}
	if err := client.SetMyCommands(context.Background(), NativeBotCommands()); err != nil {
		t.Fatal(err)
	}
	if err := client.EditMessageText(context.Background(), 10, 56, "edit", keyboard); err != nil {
		t.Fatal(err)
	}
	wantPaths := []string{"/botsecret-token/sendMessage", "/botsecret-token/answerCallbackQuery", "/botsecret-token/deleteMessage", "/botsecret-token/setMyCommands", "/botsecret-token/editMessageText"}
	for i := range wantPaths {
		if paths[i] != wantPaths[i] {
			t.Fatalf("path %d = %s", i, paths[i])
		}
	}
	if _, ok := bodies[0]["reply_markup"]; !ok {
		t.Fatalf("sendMessage missing keyboard: %#v", bodies[0])
	}
	if bodies[1]["callback_query_id"] != "cb1" {
		t.Fatalf("answer body: %#v", bodies[1])
	}
	if bodies[2]["message_id"].(float64) != 55 {
		t.Fatalf("delete body: %#v", bodies[2])
	}
	if len(bodies[3]["commands"].([]any)) != 9 {
		t.Fatalf("commands body: %#v", bodies[3])
	}
	if _, ok := bodies[4]["reply_markup"]; !ok {
		t.Fatalf("edit missing keyboard: %#v", bodies[4])
	}
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
