package database

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	applogger "work-status-bot/internal/logger"
)

func TestMongoLoggingDoesNotExposeURI(t *testing.T) {
	var buf bytes.Buffer
	log, err := applogger.New(&buf, applogger.EnvProduction, "info")
	if err != nil {
		t.Fatal(err)
	}
	LogMongoConnect(log, "work_status_bot", "failure", errors.New("connection failed"))
	LogMongoIndexes(log, "work_status_bot", "success", nil)
	out := buf.String()
	for _, want := range []string{`"event":"mongodb.connect"`, `"event":"mongodb.indexes"`, `"database":"work_status_bot"`, `"outcome":"failure"`, `"outcome":"success"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, "mongodb+srv://") || strings.Contains(out, "password") {
		t.Fatalf("mongo secret leaked: %s", out)
	}
}
