package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"
)

type Update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
		From *struct {
			ID int64 `json:"id"`
		} `json:"from,omitempty"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type PeopleService interface {
	Add(ctx context.Context, firstName, lastName string) (people.Person, error)
}

type WorkService interface {
	Start(ctx context.Context, firstName, lastName, title string) (people.Person, works.WorkRecord, error)
	Status(ctx context.Context) ([]people.Person, []works.WorkRecord, error)
	Stop(ctx context.Context, firstName, lastName, reason string) (people.Person, works.WorkRecord, error)
}

type Handler struct {
	groupChat int64
	people    PeopleService
	work      WorkService
	reports   *reports.Service
	telegram  interface {
		SendMessage(ctx context.Context, chatID int64, text string) error
	}
	logger *slog.Logger
	now    func() time.Time
}

func NewHandler(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}) *Handler {
	return NewHandlerWithLogger(groupChat, people, work, reportSvc, tg, slog.Default())
}

func NewHandlerWithLogger(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{groupChat: groupChat, people: people, work: work, reports: reportSvc, telegram: tg, logger: log, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		h.logger.Warn("telegram webhook malformed", "event", "telegram.webhook_received", "operation", "telegram", "outcome", "invalid", "error", err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if update.Message == nil {
		h.logger.Info("telegram webhook received", "event", "telegram.webhook_received", "operation", "telegram", "outcome", "skipped")
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	userID := int64(0)
	if update.Message.From != nil {
		userID = update.Message.From.ID
	}
	command := commandName(update.Message.Text)
	h.logger.Info("telegram webhook received", "event", "telegram.webhook_received", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "received")
	if update.Message.Chat.ID != h.groupChat {
		h.logger.Warn("telegram chat rejected", "event", "telegram.chat_rejected", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "rejected")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	response, err := h.HandleText(r.Context(), update.Message.Text)
	if err != nil {
		h.logger.Error("telegram command result", "event", "telegram.command_result", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "failure", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response != "" && h.telegram != nil {
		if err := h.telegram.SendMessage(r.Context(), h.groupChat, response); err != nil {
			h.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "failure", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Info("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "success")
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleText(ctx context.Context, text string) (string, error) {
	cmd, err := ParseCommand(text)
	if err != nil {
		h.logger.Warn("telegram command", "event", "telegram.command", "operation", "telegram", "command", commandName(text), "outcome", "invalid", "error", err.Error())
		return ErrorMessage(err), nil
	}
	h.logger.Info("telegram command", "event", "telegram.command", "operation", "telegram", "command", cmd.Name, "outcome", "received")
	outcome := "success"
	defer func() {
		h.logger.Info("telegram command result", "event", "telegram.command_result", "operation", "telegram", "command", cmd.Name, "outcome", outcome)
	}()
	switch cmd.Name {
	case CommandHelp:
		return HelpMessage(), nil
	case CommandAddPerson:
		p, err := h.people.Add(ctx, cmd.FirstName, cmd.LastName)
		if err != nil {
			outcome = "failure"
			return ErrorMessage(err), nil
		}
		return PersonAddedMessage(p), nil
	case CommandStartWork:
		p, record, err := h.work.Start(ctx, cmd.FirstName, cmd.LastName, cmd.Title)
		if err != nil {
			outcome = "failure"
			return ErrorMessage(err), nil
		}
		return WorkStartedMessage(p, record), nil
	case CommandStatus:
		persons, records, err := h.work.Status(ctx)
		if err != nil {
			outcome = "failure"
			return "", err
		}
		return StatusMessage(persons, records, h.now()), nil
	case CommandStopWork:
		p, record, err := h.work.Stop(ctx, cmd.FirstName, cmd.LastName, cmd.Reason)
		if err != nil {
			outcome = "failure"
			return ErrorMessage(err), nil
		}
		occurred := h.now()
		if record.StoppedAt != nil {
			occurred = *record.StoppedAt
		}
		if h.reports != nil {
			if err := h.reports.RecordStopped(ctx, p.ID, record.ID, cmd.Reason, occurred); err != nil {
				outcome = "failure"
				return "", err
			}
		}
		warning := ""
		if h.telegram != nil {
			if err := h.telegram.SendMessage(ctx, h.groupChat, StopAlertMessage(p, record)); err != nil {
				warning = "alert delivery failed"
				h.logger.Error("telegram alert send", "event", "telegram.alert_send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "failure", "error", err.Error())
				if h.reports != nil {
					if recErr := h.reports.RecordAlertFailed(ctx, p.ID, record.ID, err.Error(), h.now()); recErr != nil {
						outcome = "failure"
						return "", recErr
					}
				}
			} else {
				h.logger.Info("telegram alert send", "event", "telegram.alert_send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "success")
			}
		}
		return WorkStoppedMessage(p, record, warning), nil
	case CommandReportMonth:
		if h.reports == nil {
			outcome = "failure"
			return "", errors.New("reports service not configured")
		}
		result, err := h.reports.GenerateMonthly(ctx, cmd.Month, reports.TriggerTelegramCommand)
		if err != nil {
			outcome = "failure"
			return ErrorMessage(err), nil
		}
		return ReportMessage(result), nil
	default:
		outcome = "failure"
		return ErrorMessage(ErrMalformedCommand), nil
	}
}

func commandName(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
