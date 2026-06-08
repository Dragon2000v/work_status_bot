package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"
)

type Update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
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
	now func() time.Time
}

func NewHandler(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}) *Handler {
	return &Handler{groupChat: groupChat, people: people, work: work, reports: reportSvc, telegram: tg, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if update.Message == nil {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	if update.Message.Chat.ID != h.groupChat {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	response, err := h.HandleText(r.Context(), update.Message.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response != "" && h.telegram != nil {
		if err := h.telegram.SendMessage(r.Context(), h.groupChat, response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleText(ctx context.Context, text string) (string, error) {
	cmd, err := ParseCommand(text)
	if err != nil {
		return ErrorMessage(err), nil
	}
	switch cmd.Name {
	case CommandHelp:
		return HelpMessage(), nil
	case CommandAddPerson:
		p, err := h.people.Add(ctx, cmd.FirstName, cmd.LastName)
		if err != nil {
			return ErrorMessage(err), nil
		}
		return PersonAddedMessage(p), nil
	case CommandStartWork:
		p, record, err := h.work.Start(ctx, cmd.FirstName, cmd.LastName, cmd.Title)
		if err != nil {
			return ErrorMessage(err), nil
		}
		return WorkStartedMessage(p, record), nil
	case CommandStatus:
		persons, records, err := h.work.Status(ctx)
		if err != nil {
			return "", err
		}
		return StatusMessage(persons, records, h.now()), nil
	case CommandStopWork:
		p, record, err := h.work.Stop(ctx, cmd.FirstName, cmd.LastName, cmd.Reason)
		if err != nil {
			return ErrorMessage(err), nil
		}
		occurred := h.now()
		if record.StoppedAt != nil {
			occurred = *record.StoppedAt
		}
		if h.reports != nil {
			if err := h.reports.RecordStopped(ctx, p.ID, record.ID, cmd.Reason, occurred); err != nil {
				return "", err
			}
		}
		warning := ""
		if h.telegram != nil {
			if err := h.telegram.SendMessage(ctx, h.groupChat, StopAlertMessage(p, record)); err != nil {
				warning = "alert delivery failed"
				if h.reports != nil {
					if recErr := h.reports.RecordAlertFailed(ctx, p.ID, record.ID, err.Error(), h.now()); recErr != nil {
						return "", recErr
					}
				}
			}
		}
		return WorkStoppedMessage(p, record, warning), nil
	case CommandReportMonth:
		if h.reports == nil {
			return "", errors.New("reports service not configured")
		}
		result, err := h.reports.GenerateMonthly(ctx, cmd.Month, reports.TriggerTelegramCommand)
		if err != nil {
			return ErrorMessage(err), nil
		}
		return ReportMessage(result), nil
	default:
		return ErrorMessage(ErrMalformedCommand), nil
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
