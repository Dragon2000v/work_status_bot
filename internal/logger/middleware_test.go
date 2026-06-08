package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPMiddlewareLogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	log, err := New(&buf, EnvProduction, "info")
	if err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})
	handler := HTTPMiddleware(log)(next)
	req := httptest.NewRequest(http.MethodPost, "/health?secret=value", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	out := buf.String()
	for _, want := range []string{`"event":"http.request"`, `"method":"POST"`, `"path":"/health"`, `"status":201`, `"remote_addr":"127.0.0.1:1234"`, `"duration_ms":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, "secret=value") {
		t.Fatalf("query string leaked: %s", out)
	}
}
