package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"work-status-bot/internal/config"
	"work-status-bot/internal/groups"
	applogger "work-status-bot/internal/logger"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/telegram"
)

type fakeCronReports struct {
	reports  map[string]reports.MonthlyReport
	sendErr  error
	lastSent reports.MonthlyReport
	targets  []groups.ReportDeliveryTarget
}

type fakeWebhookRegistrar struct {
	calls       int
	webhookURL  string
	secretToken string
	err         error
}

type fakeCommandRegistrar struct {
	calls    int
	commands []telegram.BotCommand
	err      error
}

type fakeCronTargets struct {
	targets []groups.ReportDeliveryTarget
	err     error
}

func (f fakeCronTargets) CronTargets(ctx context.Context) ([]groups.ReportDeliveryTarget, error) {
	return f.targets, f.err
}

func (f *fakeWebhookRegistrar) SetWebhook(ctx context.Context, webhookURL, secretToken string) error {
	f.calls++
	f.webhookURL = webhookURL
	f.secretToken = secretToken
	return f.err
}

func (f *fakeCommandRegistrar) SetMyCommands(ctx context.Context, commands []telegram.BotCommand) error {
	f.calls++
	f.commands = commands
	return f.err
}

func (f *fakeCronReports) GenerateMonthly(ctx context.Context, month, source string) (reports.GenerateResult, error) {
	if month == "" {
		month = "2026-06"
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}$`).MatchString(month) {
		return reports.GenerateResult{}, fmt.Errorf("invalid month")
	}
	if r, ok := f.reports[month]; ok {
		return reports.GenerateResult{Report: r, Duplicate: true}, nil
	}
	r := reports.MonthlyReport{Month: month, Content: "report"}
	f.reports[month] = r
	return reports.GenerateResult{Report: r}, nil
}

func (f *fakeCronReports) SendReportToGroup(ctx context.Context, report reports.MonthlyReport, duplicate bool) error {
	f.lastSent = report
	return f.sendErr
}

func (f *fakeCronReports) SendReportToTargets(ctx context.Context, report reports.MonthlyReport, duplicate bool, targets []groups.ReportDeliveryTarget) error {
	f.lastSent = report
	f.targets = targets
	return f.sendErr
}

func TestHealth(t *testing.T) {
	router := NewRouter("secret", http.NewServeMux(), &fakeCronReports{reports: map[string]reports.MonthlyReport{}})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/health", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
	}
}

func TestHealthRequestLogging(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	router := NewRouterWithLogger("secret", http.NewServeMux(), &fakeCronReports{reports: map[string]reports.MonthlyReport{}}, log)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/health", nil))
	out := buf.String()
	for _, want := range []string{`"event":"http.request"`, `"method":"GET"`, `"path":"/health"`, `"status":200`, `"duration_ms":`, `"remote_addr":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
}

func TestCronMonthlyReport(t *testing.T) {
	fake := &fakeCronReports{reports: map[string]reports.MonthlyReport{}}
	router := NewRouter("secret", http.NewServeMux(), fake)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`)))
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("missing secret status %d", res.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "secret")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assertCronStatus(t, res, "generated")
	req = httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "secret")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assertCronStatus(t, res, "duplicate")
	req = httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"bad"}`))
	req.Header.Set("X-Cron-Secret", "secret")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("invalid month status %d", res.Code)
	}
}

func TestCronLogging(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	fake := &fakeCronReports{reports: map[string]reports.MonthlyReport{}}
	router := NewRouterWithLogger("super-secret-cron", http.NewServeMux(), fake, log)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`)))
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("want unauthorized")
	}

	req := httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"bad"}`))
	req.Header.Set("X-Cron-Secret", "super-secret-cron")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("want invalid month")
	}

	req = httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "super-secret-cron")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assertCronStatus(t, res, "generated")

	req = httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "super-secret-cron")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assertCronStatus(t, res, "duplicate")

	out := buf.String()
	for _, want := range []string{`"event":"cron.monthly_report_triggered"`, `"event":"cron.monthly_report_result"`, `"event":"cron.monthly_report_send"`, `"outcome":"unauthorized"`, `"outcome":"invalid"`, `"outcome":"generated"`, `"outcome":"duplicate"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, "super-secret-cron") || strings.Contains(out, "X-Cron-Secret") {
		t.Fatalf("cron secret leaked: %s", out)
	}
}

func TestCronSendFailureLogging(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	fake := &fakeCronReports{reports: map[string]reports.MonthlyReport{}, sendErr: errors.New("send failed")}
	router := NewRouterWithLogger("secret", http.NewServeMux(), fake, log)
	req := httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "secret")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", res.Code)
	}
	out := buf.String()
	if !strings.Contains(out, `"event":"cron.monthly_report_send"`) || !strings.Contains(out, `"outcome":"failed"`) {
		t.Fatalf("missing send failure log: %s", out)
	}
}

func TestCronMonthlyReportFanOutTargets(t *testing.T) {
	fake := &fakeCronReports{reports: map[string]reports.MonthlyReport{}}
	targets := fakeCronTargets{targets: []groups.ReportDeliveryTarget{{TelegramChatID: -1001}, {TelegramChatID: -1002}}}
	router := NewRouterWithGroupTargets("secret", http.NewServeMux(), fake, targets, nil)
	req := httptest.NewRequest(http.MethodPost, "/cron/monthly-report", bytes.NewBufferString(`{"month":"2026-06"}`))
	req.Header.Set("X-Cron-Secret", "secret")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assertCronStatus(t, res, "generated")
	if len(fake.targets) != 2 || fake.targets[0].TelegramChatID != -1001 || fake.targets[1].TelegramChatID != -1002 {
		t.Fatalf("fan-out targets not passed: %#v", fake.targets)
	}
}

func TestStartupLoggingDoesNotExposeSecrets(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	cfg := config.Config{
		AppEnv:                "production",
		LogLevel:              "info",
		AppAddr:               ":8080",
		MongoDBURI:            "mongodb+srv://user:password@cluster.example.net/?retryWrites=true",
		MongoDBDatabase:       "work_status_bot",
		TelegramBotToken:      "telegram-token",
		TelegramWebhookSecret: "webhook-secret",
		CronSecret:            "cron-secret",
	}
	logAppStart(log)
	logConfigLoaded(log, cfg)
	out := buf.String()
	for _, want := range []string{`"event":"app.start"`, `"event":"config.loaded"`, `"app_env":"production"`, `"log_level":"info"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	for _, secret := range []string{cfg.MongoDBURI, "password", cfg.TelegramBotToken, cfg.TelegramWebhookSecret, cfg.CronSecret} {
		if strings.Contains(out, secret) {
			t.Fatalf("secret leaked: %s in %s", secret, out)
		}
	}
}

func TestRegisterTelegramWebhookSkipAndCall(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	registrar := &fakeWebhookRegistrar{}
	cfg := config.Config{TelegramAutoSetWebhook: false, TelegramWebhookURL: "https://example.com/telegram/webhook", TelegramWebhookSecret: "webhook-secret"}
	if err := registerTelegramWebhook(context.Background(), cfg, registrar, log); err != nil {
		t.Fatal(err)
	}
	if registrar.calls != 0 || !strings.Contains(buf.String(), `"outcome":"skipped"`) {
		t.Fatalf("skip failed calls=%d logs=%s", registrar.calls, buf.String())
	}

	cfg.TelegramAutoSetWebhook = true
	if err := registerTelegramWebhook(context.Background(), cfg, registrar, log); err != nil {
		t.Fatal(err)
	}
	if registrar.calls != 1 || registrar.webhookURL != cfg.TelegramWebhookURL || registrar.secretToken != cfg.TelegramWebhookSecret {
		t.Fatalf("registration not called correctly: %#v", registrar)
	}
	if strings.Contains(buf.String(), cfg.TelegramWebhookSecret) {
		t.Fatalf("webhook secret leaked: %s", buf.String())
	}
}

func TestRegisterTelegramWebhookFailure(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	registrar := &fakeWebhookRegistrar{err: errors.New("telegram failed")}
	cfg := config.Config{TelegramAutoSetWebhook: true, TelegramWebhookURL: "https://example.com/telegram/webhook", TelegramWebhookSecret: "webhook-secret"}
	if err := registerTelegramWebhook(context.Background(), cfg, registrar, log); err == nil {
		t.Fatalf("want registration error")
	}
}

func TestRegisterTelegramCommandsLogsSuccessAndFailureWithoutStopping(t *testing.T) {
	var buf bytes.Buffer
	log, _ := applogger.New(&buf, applogger.EnvProduction, "info")
	ok := &fakeCommandRegistrar{}
	registerTelegramCommands(context.Background(), ok, log)
	if ok.calls != 1 || len(ok.commands) != 9 {
		t.Fatalf("commands not registered: %#v", ok)
	}
	fail := &fakeCommandRegistrar{err: errors.New("telegram-token webhook-secret cron-secret failed")}
	registerTelegramCommands(context.Background(), fail, log)
	if fail.calls != 1 {
		t.Fatalf("failure path did not call registrar")
	}
	out := buf.String()
	for _, want := range []string{`"event":"telegram.command_menu_registration"`, `"outcome":"attempt"`, `"outcome":"success"`, `"outcome":"failure"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	for _, secret := range []string{"telegram-token", "webhook-secret", "cron-secret"} {
		if strings.Contains(out, secret) {
			t.Fatalf("secret leaked: %s", out)
		}
	}
}

func assertCronStatus(t *testing.T, res *httptest.ResponseRecorder, want string) {
	t.Helper()
	if res.Code != http.StatusOK {
		t.Fatalf("status %d body %s", res.Code, res.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != want {
		t.Fatalf("want %s got %s", want, body["status"])
	}
}
