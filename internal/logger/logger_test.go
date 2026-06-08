package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewSelectsLocalTextAndProductionJSON(t *testing.T) {
	var local bytes.Buffer
	localLogger, err := New(&local, EnvLocal, "info")
	if err != nil {
		t.Fatal(err)
	}
	localLogger.Info("test", "event", "local.event")
	if !strings.Contains(local.String(), "event=local.event") {
		t.Fatalf("local text log missing field: %s", local.String())
	}

	var prod bytes.Buffer
	prodLogger, err := New(&prod, EnvProduction, "info")
	if err != nil {
		t.Fatal(err)
	}
	prodLogger.Info("test", "event", "prod.event")
	if !strings.Contains(prod.String(), `"event":"prod.event"`) {
		t.Fatalf("production json log missing field: %s", prod.String())
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for input, want := range cases {
		got, err := ParseLevel(input)
		if err != nil {
			t.Fatalf("%s: %v", input, err)
		}
		if got != want {
			t.Fatalf("%s: want %v got %v", input, want, got)
		}
	}
	if _, err := ParseLevel("trace"); err == nil {
		t.Fatalf("want invalid level error")
	}
}

func TestLogLevelFiltersDebug(t *testing.T) {
	var buf bytes.Buffer
	log, err := New(&buf, EnvProduction, "info")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("hidden", "event", "hidden")
	log.Info("visible", "event", "visible")
	out := buf.String()
	if strings.Contains(out, "hidden") || !strings.Contains(out, "visible") {
		t.Fatalf("bad level filtering: %s", out)
	}
}

func TestRedactSecrets(t *testing.T) {
	mongoURI := "mongodb+srv://user:password@cluster.example.net/?retryWrites=true"
	token := "123456:telegram-token"
	webhookSecret := "webhook-secret"
	cronSecret := "cron-secret"
	input := strings.Join([]string{mongoURI, token, webhookSecret, cronSecret}, " ")
	out := RedactSecrets(input, mongoURI, "password", token, webhookSecret, cronSecret)
	for _, secret := range []string{mongoURI, "password", token, webhookSecret, cronSecret} {
		if strings.Contains(out, secret) {
			t.Fatalf("secret leaked: %s in %s", secret, out)
		}
	}
}
