package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"work-status-bot/internal/config"
	"work-status-bot/internal/database"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/telegram"
	"work-status-bot/internal/works"
)

type cronReportService interface {
	GenerateMonthly(ctx context.Context, month, source string) (reports.GenerateResult, error)
	SendReportToGroup(ctx context.Context, report reports.MonthlyReport, duplicate bool) error
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	client, err := database.Connect(ctx, cfg.MongoDBURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Shutdown(context.Background(), client) }()

	db := client.Database(cfg.MongoDBDatabase)
	if err := database.EnsureIndexes(ctx, db); err != nil {
		log.Fatal(err)
	}
	peopleRepo := people.NewRepository(db)
	workRepo := works.NewRepository(db)
	reportRepo := reports.NewRepository(db)
	peopleSvc := people.NewService(peopleRepo)
	workSvc := works.NewService(peopleRepo, workRepo)
	tg := telegram.NewClient(cfg.TelegramBotToken, http.DefaultClient)
	reportSvc := reports.NewService(reportRepo, peopleRepo, workSvc, tg, cfg.TelegramGroupChatID)
	handler := telegram.NewHandler(cfg.TelegramGroupChatID, peopleSvc, workSvc, reportSvc, tg)

	router := NewRouter(cfg.CronSecret, handler, reportSvc)
	log.Fatal(http.ListenAndServe(cfg.AppAddr, router))
}

func NewRouter(cronSecret string, telegramHandler http.Handler, reportSvc cronReportService) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/telegram/webhook", telegramHandler.ServeHTTP)
	r.Post("/cron/monthly-report", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Cron-Secret") != cronSecret {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var req struct {
			Month string `json:"month"`
		}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := reportSvc.GenerateMonthly(ctx, req.Month, reports.TriggerExternalHTTP)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := reportSvc.SendReportToGroup(ctx, result.Report, result.Duplicate); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		status := "generated"
		if result.Duplicate {
			status = "duplicate"
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": status, "month": result.Report.Month})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
