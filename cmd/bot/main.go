package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"

	"work-status-bot/internal/config"
	"work-status-bot/internal/database"
	"work-status-bot/internal/groups"
	applogger "work-status-bot/internal/logger"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/telegram"
	"work-status-bot/internal/users"
	"work-status-bot/internal/works"
)

type cronReportService interface {
	GenerateMonthly(ctx context.Context, month, source string) (reports.GenerateResult, error)
	SendReportToGroup(ctx context.Context, report reports.MonthlyReport, duplicate bool) error
	SendReportToTargets(ctx context.Context, report reports.MonthlyReport, duplicate bool, targets []groups.ReportDeliveryTarget) error
}

type cronTargetService interface {
	CronTargets(ctx context.Context) ([]groups.ReportDeliveryTarget, error)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Default().Error("config load", "event", "config.loaded", "operation", "startup", "outcome", "failure", "error", err.Error())
		os.Exit(1)
	}
	log, err := applogger.New(os.Stdout, cfg.AppEnv, cfg.LogLevel)
	if err != nil {
		slog.Default().Error("logger setup", "event", "app.start", "operation", "startup", "outcome", "failure", "error", err.Error())
		os.Exit(1)
	}
	logAppStart(log)
	logConfigLoaded(log, cfg)

	ctx := context.Background()
	database.LogMongoConnect(log, cfg.MongoDBDatabase, "attempt", nil)
	client, err := database.Connect(ctx, cfg.MongoDBURI)
	if err != nil {
		database.LogMongoConnect(log, cfg.MongoDBDatabase, "failure", err)
		os.Exit(1)
	}
	database.LogMongoConnect(log, cfg.MongoDBDatabase, "success", nil)
	defer func() { _ = database.Shutdown(context.Background(), client) }()

	db := client.Database(cfg.MongoDBDatabase)
	if err := database.EnsureIndexes(ctx, db); err != nil {
		database.LogMongoIndexes(log, cfg.MongoDBDatabase, "failure", err)
		os.Exit(1)
	}
	database.LogMongoIndexes(log, cfg.MongoDBDatabase, "success", nil)
	peopleRepo := people.NewRepository(db)
	workRepo := works.NewRepository(db)
	reportRepo := reports.NewRepository(db)
	userRepo := users.NewRepository(db)
	groupRepo := groups.NewRepository(db)
	peopleSvc := people.NewService(peopleRepo)
	workSvc := works.NewService(peopleRepo, workRepo)
	userSvc := users.NewService(userRepo)
	groupSvc := groups.NewService(groupRepo, cfg.TelegramGroupChatID)
	tg := telegram.NewClientWithLogger(cfg.TelegramBotToken, http.DefaultClient, log)
	if err := registerTelegramWebhook(ctx, cfg, tg, log); err != nil {
		os.Exit(1)
	}
	reportSvc := reports.NewServiceWithLogger(reportRepo, peopleRepo, workSvc, tg, cfg.TelegramGroupChatID, log)
	handler := telegram.NewHandlerWithGroupsUsersAndLogger(cfg.TelegramGroupChatID, peopleSvc, workSvc, reportSvc, groupSvc, userSvc, tg, log)

	router := NewRouterWithGroupTargets(cfg.CronSecret, handler, reportSvc, groupSvc, log)
	log.Info("application listen", "event", "app.listen", "operation", "startup", "addr", cfg.AppAddr, "outcome", "success")
	if err := http.ListenAndServe(cfg.AppAddr, router); err != nil {
		log.Error("application listen", "event", "app.listen", "operation", "startup", "addr", cfg.AppAddr, "outcome", "failure", "error", err.Error())
		os.Exit(1)
	}
}

type webhookRegistrar interface {
	SetWebhook(ctx context.Context, webhookURL, secretToken string) error
}

func registerTelegramWebhook(ctx context.Context, cfg config.Config, tg webhookRegistrar, log *slog.Logger) error {
	if !cfg.TelegramAutoSetWebhook {
		log.Info("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "skipped")
		return nil
	}
	if err := tg.SetWebhook(ctx, cfg.TelegramWebhookURL, cfg.TelegramWebhookSecret); err != nil {
		return err
	}
	return nil
}

func NewRouter(cronSecret string, telegramHandler http.Handler, reportSvc cronReportService) http.Handler {
	return NewRouterWithLogger(cronSecret, telegramHandler, reportSvc, slog.Default())
}

func NewRouterWithLogger(cronSecret string, telegramHandler http.Handler, reportSvc cronReportService, log *slog.Logger) http.Handler {
	return NewRouterWithGroupTargets(cronSecret, telegramHandler, reportSvc, nil, log)
}

func NewRouterWithGroupTargets(cronSecret string, telegramHandler http.Handler, reportSvc cronReportService, targetSvc cronTargetService, log *slog.Logger) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	r := chi.NewRouter()
	r.Use(applogger.HTTPMiddleware(log))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/telegram/webhook", telegramHandler.ServeHTTP)
	r.Post("/cron/monthly-report", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Cron-Secret") != cronSecret {
			log.Warn("cron monthly report", "event", "cron.monthly_report_triggered", "operation", "cron", "outcome", "unauthorized")
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var req struct {
			Month string `json:"month"`
		}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				log.Warn("cron monthly report", "event", "cron.monthly_report_triggered", "operation", "cron", "outcome", "invalid", "error", err.Error())
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
		}
		if req.Month != "" && !validMonth(req.Month) {
			log.Warn("cron monthly report", "event", "cron.monthly_report_triggered", "operation", "cron", "month", req.Month, "outcome", "invalid")
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid month"})
			return
		}
		start := time.Now()
		log.Info("cron monthly report", "event", "cron.monthly_report_triggered", "operation", "cron", "month", req.Month, "outcome", "received")
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := reportSvc.GenerateMonthly(ctx, req.Month, reports.TriggerExternalHTTP)
		if err != nil {
			log.Error("cron monthly report", "event", "cron.monthly_report_result", "operation", "cron", "month", req.Month, "outcome", "failed", "duration_ms", time.Since(start).Milliseconds(), "error", err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if targetSvc != nil {
			targets, err := targetSvc.CronTargets(ctx)
			if err != nil {
				log.Error("cron monthly report send", "event", "cron.monthly_report_send", "operation", "cron", "month", result.Report.Month, "outcome", "failed", "duration_ms", time.Since(start).Milliseconds(), "error", err.Error())
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			if err := reportSvc.SendReportToTargets(ctx, result.Report, result.Duplicate, targets); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		} else if err := reportSvc.SendReportToGroup(ctx, result.Report, result.Duplicate); err != nil {
			log.Error("cron monthly report send", "event", "cron.monthly_report_send", "operation", "cron", "month", result.Report.Month, "outcome", "failed", "duration_ms", time.Since(start).Milliseconds(), "error", err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		status := "generated"
		if result.Duplicate {
			status = "duplicate"
		}
		log.Info("cron monthly report", "event", "cron.monthly_report_result", "operation", "cron", "month", result.Report.Month, "outcome", status, "duration_ms", time.Since(start).Milliseconds())
		if targetSvc == nil {
			log.Info("cron monthly report send", "event", "cron.monthly_report_send", "operation", "cron", "month", result.Report.Month, "outcome", "success", "duration_ms", time.Since(start).Milliseconds())
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": status, "month": result.Report.Month})
	})
	return r
}

func logAppStart(log *slog.Logger) {
	log.Info("application start", "event", "app.start", "operation", "startup", "outcome", "success")
}

func logConfigLoaded(log *slog.Logger, cfg config.Config) {
	log.Info("config loaded", "event", "config.loaded", "operation", "startup", "app_env", cfg.AppEnv, "log_level", cfg.LogLevel, "outcome", "success")
}

var monthPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)

func validMonth(month string) bool {
	return monthPattern.MatchString(month)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
