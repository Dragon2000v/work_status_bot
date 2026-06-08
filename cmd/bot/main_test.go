package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"work-status-bot/internal/reports"
)

type fakeCronReports struct {
	reports map[string]reports.MonthlyReport
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
	return nil
}

func TestHealth(t *testing.T) {
	router := NewRouter("secret", http.NewServeMux(), &fakeCronReports{reports: map[string]reports.MonthlyReport{}})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/health", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
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
