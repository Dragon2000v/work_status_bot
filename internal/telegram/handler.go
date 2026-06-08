package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"work-status-bot/internal/i18n"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/users"
	"work-status-bot/internal/works"
)

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	From      *User  `json:"from,omitempty"`
	Chat      Chat   `json:"chat"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from,omitempty"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

type User struct {
	ID int64 `json:"id"`
}

type Chat struct {
	ID int64 `json:"id"`
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
	users     interface {
		LookupLanguage(ctx context.Context, telegramUserID int64) (i18n.Language, error)
		SetLanguage(ctx context.Context, telegramUserID int64, lang i18n.Language) (users.UserSetting, error)
	}
	telegram telegramAPI
	logger   *slog.Logger
	now      func() time.Time
}

type telegramAPI interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type keyboardSender interface {
	SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error
}

type callbackAnswerer interface {
	AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error
}

type messageDeleter interface {
	DeleteMessage(ctx context.Context, chatID int64, messageID int) error
}

func NewHandler(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}) *Handler {
	return NewHandlerWithLogger(groupChat, people, work, reportSvc, tg, slog.Default())
}

func NewHandlerWithLogger(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}, log *slog.Logger) *Handler {
	return NewHandlerWithUsersAndLogger(groupChat, people, work, reportSvc, nil, tg, log)
}

func NewHandlerWithUsersAndLogger(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, userSvc interface {
	LookupLanguage(ctx context.Context, telegramUserID int64) (i18n.Language, error)
	SetLanguage(ctx context.Context, telegramUserID int64, lang i18n.Language) (users.UserSetting, error)
}, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{groupChat: groupChat, people: people, work: work, reports: reportSvc, users: userSvc, telegram: tg, logger: log, now: func() time.Time { return time.Now().UTC() }}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		h.logger.Warn("telegram webhook malformed", "event", "telegram.webhook_received", "operation", "telegram", "outcome", "invalid", "error", err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if update.CallbackQuery != nil {
		h.handleCallbackHTTP(w, r, update.CallbackQuery)
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
	response, err := h.HandleTextForUser(r.Context(), update.Message.Text, userID)
	if err != nil {
		h.logger.Error("telegram command result", "event", "telegram.command_result", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "failure", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response != "" && h.telegram != nil {
		if command == CommandHelp {
			if err := h.sendMessageWithKeyboard(r.Context(), h.groupChat, response, MainMenuKeyboard(h.languageForUser(r.Context(), userID))); err != nil {
				h.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "failure", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err := h.telegram.SendMessage(r.Context(), h.groupChat, response); err != nil {
			h.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "failure", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Info("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", h.groupChat, "outcome", "success")
	}
	h.deleteCommandMessage(r.Context(), update.Message.Chat.ID, update.Message.MessageID, userID, command)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleText(ctx context.Context, text string) (string, error) {
	return h.HandleTextForUser(ctx, text, 0)
}

func (h *Handler) HandleTextForUser(ctx context.Context, text string, userID int64) (string, error) {
	lang := h.languageForUser(ctx, userID)
	cmd, err := ParseCommand(text)
	if err != nil {
		h.logger.Warn("telegram command", "event", "telegram.command", "operation", "telegram", "command", commandName(text), "outcome", "invalid", "error", err.Error())
		return ErrorMessageLang(lang, err), nil
	}
	h.logger.Info("telegram command", "event", "telegram.command", "operation", "telegram", "command", cmd.Name, "outcome", "received")
	outcome := "success"
	defer func() {
		h.logger.Info("telegram command result", "event", "telegram.command_result", "operation", "telegram", "command", cmd.Name, "outcome", outcome)
	}()
	switch cmd.Name {
	case CommandHelp:
		return HelpMessageLang(lang), nil
	case CommandAddPerson:
		p, err := h.people.Add(ctx, cmd.FirstName, cmd.LastName)
		if err != nil {
			outcome = "failure"
			return ErrorMessageLang(lang, err), nil
		}
		return PersonAddedMessageLang(lang, p), nil
	case CommandStartWork:
		p, record, err := h.work.Start(ctx, cmd.FirstName, cmd.LastName, cmd.Title)
		if err != nil {
			outcome = "failure"
			return ErrorMessageLang(lang, err), nil
		}
		return WorkStartedMessageLang(lang, p, record), nil
	case CommandStatus:
		persons, records, err := h.work.Status(ctx)
		if err != nil {
			outcome = "failure"
			return "", err
		}
		return StatusMessageLang(lang, persons, records, h.now()), nil
	case CommandStopWork:
		p, record, err := h.work.Stop(ctx, cmd.FirstName, cmd.LastName, cmd.Reason)
		if err != nil {
			outcome = "failure"
			return ErrorMessageLang(lang, err), nil
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
		return WorkStoppedMessageLang(lang, p, record, warning), nil
	case CommandReportMonth:
		if h.reports == nil {
			outcome = "failure"
			return "", errors.New("reports service not configured")
		}
		result, err := h.reports.GenerateMonthly(ctx, cmd.Month, reports.TriggerTelegramCommand)
		if err != nil {
			outcome = "failure"
			return ErrorMessageLang(lang, err), nil
		}
		return ReportMessage(result), nil
	default:
		outcome = "failure"
		return ErrorMessageLang(lang, ErrMalformedCommand), nil
	}
}

func (h *Handler) handleCallbackHTTP(w http.ResponseWriter, r *http.Request, cb *CallbackQuery) {
	chatID := int64(0)
	messageID := 0
	if cb.Message != nil {
		chatID = cb.Message.Chat.ID
		messageID = cb.Message.MessageID
	}
	userID := int64(0)
	if cb.From != nil {
		userID = cb.From.ID
	}
	action := cb.Data
	h.logger.Info("telegram callback received", "event", "telegram.callback_received", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", "received")
	if chatID != h.groupChat {
		_ = h.answerCallback(r.Context(), cb.ID, "", userID, action)
		h.logger.Warn("telegram chat rejected", "event", "telegram.chat_rejected", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", "rejected")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lang := h.languageForUser(r.Context(), userID)
	answerErr := h.answerCallback(r.Context(), cb.ID, "", userID, action)
	text, keyboard, outcome, err := h.HandleCallback(r.Context(), action, userID, messageID)
	if err != nil {
		outcome = "failure"
		text = ErrorMessageLang(lang, err)
	}
	if text != "" && h.telegram != nil {
		if err := h.sendMessageWithKeyboard(r.Context(), h.groupChat, text, keyboard); err != nil {
			h.logger.Error("telegram callback result", "event", "telegram.callback_result", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", "failure", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if answerErr != nil && outcome == "success" {
		outcome = "failure"
	}
	h.logger.Info("telegram callback result", "event", "telegram.callback_result", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", outcome)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleCallback(ctx context.Context, action string, userID int64, messageID int) (string, *InlineKeyboardMarkup, string, error) {
	lang := h.languageForUser(ctx, userID)
	switch action {
	case CallbackMenuHelp:
		return HelpMessageLang(lang), MainMenuKeyboard(lang), "success", nil
	case CallbackMenuSettings:
		return SettingsMessage(lang), SettingsKeyboard(lang), "success", nil
	case CallbackSettingsLanguage:
		return LanguageMessage(lang), LanguageKeyboard(lang), "success", nil
	case CallbackMenuStatus:
		persons, records, err := h.work.Status(ctx)
		if err != nil {
			return "", nil, "failure", err
		}
		return StatusMessageLang(lang, persons, records, h.now()), nil, "success", nil
	case CallbackMenuAddPerson, CallbackMenuStartWork, CallbackMenuStopWork, CallbackMenuReportMonth:
		return ActionPromptMessage(lang, action), nil, "success", nil
	case CallbackLanguageUK, CallbackLanguageEN, CallbackLanguageRU:
		selected, ok := languageFromCallback(action)
		if !ok {
			return i18n.T(lang, i18n.KeyUnsupportedAction), nil, "unsupported", nil
		}
		if h.users != nil {
			if _, err := h.users.SetLanguage(ctx, userID, selected); err != nil {
				h.logger.Error("telegram user language changed", "event", "telegram.user_language_changed", "operation", "telegram", "user_id", userID, "language", selected, "outcome", "failure", "error", err.Error())
				return "", nil, "failure", err
			}
		}
		h.logger.Info("telegram user language changed", "event", "telegram.user_language_changed", "operation", "telegram", "user_id", userID, "language", selected, "outcome", "success")
		return i18n.T(selected, i18n.KeyLanguageChanged), MainMenuKeyboard(selected), "success", nil
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction), nil, "unsupported", nil
	}
}

func (h *Handler) languageForUser(ctx context.Context, userID int64) i18n.Language {
	if h.users == nil {
		return i18n.Ukrainian
	}
	lang, err := h.users.LookupLanguage(ctx, userID)
	if err != nil {
		h.logger.Warn("telegram user language lookup", "event", "telegram.user_language_lookup", "operation", "telegram", "user_id", userID, "outcome", "failure", "error", err.Error())
		return i18n.Ukrainian
	}
	return lang
}

func (h *Handler) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error {
	if h.telegram == nil {
		return nil
	}
	if keyboard != nil {
		if sender, ok := h.telegram.(keyboardSender); ok {
			return sender.SendMessageWithKeyboard(ctx, chatID, text, keyboard)
		}
	}
	return h.telegram.SendMessage(ctx, chatID, text)
}

func (h *Handler) answerCallback(ctx context.Context, callbackID, text string, userID int64, action string) error {
	if callbackID == "" {
		return nil
	}
	answerer, ok := h.telegram.(callbackAnswerer)
	if !ok {
		h.logger.Info("telegram callback answer", "event", "telegram.callback_answer", "operation", "telegram_send", "user_id", userID, "callback_action", action, "outcome", "success")
		return nil
	}
	if err := answerer.AnswerCallbackQuery(ctx, callbackID, text); err != nil {
		h.logger.Error("telegram callback answer", "event", "telegram.callback_answer", "operation", "telegram_send", "user_id", userID, "callback_action", action, "outcome", "failure", "error", err.Error())
		return err
	}
	h.logger.Info("telegram callback answer", "event", "telegram.callback_answer", "operation", "telegram_send", "user_id", userID, "callback_action", action, "outcome", "success")
	return nil
}

func (h *Handler) deleteCommandMessage(ctx context.Context, chatID int64, messageID int, userID int64, command string) {
	if messageID == 0 || command == "" || !strings.HasPrefix(command, "/") {
		return
	}
	deleter, ok := h.telegram.(messageDeleter)
	if !ok {
		return
	}
	h.logger.Info("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "attempt")
	if err := deleter.DeleteMessage(ctx, chatID, messageID); err != nil {
		h.logger.Warn("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "failure", "error", err.Error())
		return
	}
	h.logger.Info("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "success")
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
