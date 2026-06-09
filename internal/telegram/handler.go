package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"work-status-bot/internal/groups"
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
	ID       int64  `json:"id"`
	Username string `json:"username,omitempty"`
}

type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
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
	groups    *groups.Service
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
	return NewHandlerWithGroupsUsersAndLogger(groupChat, people, work, reportSvc, groups.NewService(nil, groupChat), userSvc, tg, log)
}

func NewHandlerWithGroupsUsersAndLogger(groupChat int64, people PeopleService, work WorkService, reportSvc *reports.Service, groupSvc *groups.Service, userSvc interface {
	LookupLanguage(ctx context.Context, telegramUserID int64) (i18n.Language, error)
	SetLanguage(ctx context.Context, telegramUserID int64, lang i18n.Language) (users.UserSetting, error)
}, tg interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	if groupSvc == nil {
		groupSvc = groups.NewService(nil, groupChat)
	}
	return &Handler{groupChat: groupChat, people: people, work: work, reports: reportSvc, groups: groupSvc, users: userSvc, telegram: tg, logger: log, now: func() time.Time { return time.Now().UTC() }}
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
	username := ""
	if update.Message.From != nil {
		username = update.Message.From.Username
	}
	response, err := h.HandleTextForChat(r.Context(), update.Message.Text, userID, username, update.Message.Chat)
	if err != nil {
		h.logger.Error("telegram command result", "event", "telegram.command_result", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "failure", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response != "" && h.telegram != nil {
		if command == CommandHelp {
			if err := h.sendMessageWithKeyboard(r.Context(), update.Message.Chat.ID, response, MainMenuKeyboard(h.languageForUser(r.Context(), userID))); err != nil {
				h.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", update.Message.Chat.ID, "outcome", "failure", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err := h.telegram.SendMessage(r.Context(), update.Message.Chat.ID, response); err != nil {
			h.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", update.Message.Chat.ID, "outcome", "failure", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Info("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", update.Message.Chat.ID, "outcome", "success")
	}
	h.deleteCommandMessage(r.Context(), update.Message.Chat.ID, update.Message.MessageID, userID, command)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleText(ctx context.Context, text string) (string, error) {
	return h.HandleTextForUser(ctx, text, 0)
}

func (h *Handler) HandleTextForUser(ctx context.Context, text string, userID int64) (string, error) {
	return h.handleText(ctx, text, userID, "", Chat{ID: h.groupChat, Type: "group"})
}

func (h *Handler) HandleTextForChat(ctx context.Context, text string, userID int64, username string, chat Chat) (string, error) {
	return h.handleText(ctx, text, userID, username, chat)
}

func (h *Handler) handleText(ctx context.Context, text string, userID int64, username string, chat Chat) (string, error) {
	lang := h.languageForUser(ctx, userID)
	cmd, err := ParseCommand(text)
	if err != nil {
		h.logger.Warn("telegram command", "event", "telegram.command", "operation", "telegram", "command", commandName(text), "outcome", "invalid", "error", err.Error())
		return ErrorMessageLang(lang, err), nil
	}
	decision := h.authorize(ctx, chat, cmd.Name)
	if !decision.Allowed {
		h.logAuthorizationRejected(chat.ID, userID, cmd.Name, decision.Reason)
		if decision.Reason == groups.ReasonPrivateChatSetup {
			return SetupGroupOnlyMessage(lang), nil
		}
		return SetupRequiredMessage(lang), nil
	}
	h.logger.Info("telegram command", "event", "telegram.command", "operation", "telegram", "chat_id", chat.ID, "command", cmd.Name, "outcome", "received")
	outcome := "success"
	defer func() {
		h.logger.Info("telegram command result", "event", "telegram.command_result", "operation", "telegram", "chat_id", chat.ID, "command", cmd.Name, "outcome", outcome)
	}()
	switch cmd.Name {
	case CommandHelp:
		return HelpMessageLang(lang), nil
	case CommandSetup:
		if chat.Type == "private" {
			outcome = "failure"
			h.logGroupSetup(chat.ID, userID, "rejected", groups.ReasonPrivateChatSetup)
			return SetupGroupOnlyMessage(lang), nil
		}
		h.logGroupSetup(chat.ID, userID, "attempt", "")
		result, err := h.groups.Setup(ctx, groups.GroupSetupRequest{
			TelegramChatID: chat.ID,
			Title:          chat.Title,
			SetupUserID:    userID,
			SetupUsername:  username,
			OccurredAt:     h.now(),
		})
		if err != nil {
			outcome = "failure"
			h.logGroupSetup(chat.ID, userID, "failure", err.Error())
			return ErrorMessageLang(lang, err), nil
		}
		h.logGroupSetup(chat.ID, userID, result.Outcome, "")
		return SetupMessage(lang, result.Outcome), nil
	case CommandGroups:
		configured, err := h.groups.ListAll(ctx)
		if err != nil {
			outcome = "failure"
			h.logger.Error("telegram group list", "event", "telegram.group_list", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", "failure", "error", err.Error())
			return "", err
		}
		h.logger.Info("telegram group list", "event", "telegram.group_list", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", "success", "group_count", len(configured))
		return GroupsMessage(lang, configured, h.groupChat), nil
	case CommandDisableGroup:
		group, disableOutcome, err := h.groups.Disable(ctx, chat.ID)
		if errors.Is(err, groups.ErrNotFound) {
			h.logger.Warn("telegram group disable", "event", "telegram.group_disable", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", "rejected")
			return NoStoredGroupMessage(lang), nil
		}
		if err != nil {
			outcome = "failure"
			h.logger.Error("telegram group disable", "event", "telegram.group_disable", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", "failure", "error", err.Error())
			return "", err
		}
		_ = group
		h.logger.Info("telegram group disable", "event", "telegram.group_disable", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", disableOutcome)
		return DisableGroupMessage(lang, disableOutcome), nil
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
			if err := h.telegram.SendMessage(ctx, chat.ID, StopAlertMessage(p, record)); err != nil {
				warning = "alert delivery failed"
				h.logger.Error("telegram alert send", "event", "telegram.alert_send", "operation", "telegram_send", "chat_id", chat.ID, "outcome", "failure", "error", err.Error())
				if h.reports != nil {
					if recErr := h.reports.RecordAlertFailed(ctx, p.ID, record.ID, err.Error(), h.now()); recErr != nil {
						outcome = "failure"
						return "", recErr
					}
				}
			} else {
				h.logger.Info("telegram alert send", "event", "telegram.alert_send", "operation", "telegram_send", "chat_id", chat.ID, "outcome", "success")
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
		if h.telegram != nil {
			if err := h.telegram.SendMessage(ctx, chat.ID, ReportMessage(result)); err != nil {
				outcome = "failure"
				return "", err
			}
			return "", nil
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
	lang := h.languageForUser(r.Context(), userID)
	answerErr := h.answerCallback(r.Context(), cb.ID, "", userID, action)
	command := commandForCallback(action)
	decision := h.authorize(r.Context(), Chat{ID: chatID, Type: chatTypeFromID(chatID)}, command)
	if !decision.Allowed {
		h.logAuthorizationRejected(chatID, userID, command, decision.Reason)
		if h.telegram != nil {
			if err := h.telegram.SendMessage(r.Context(), chatID, SetupRequiredMessage(lang)); err != nil {
				h.logger.Error("telegram callback result", "event", "telegram.callback_result", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", "failure", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		outcome := "success"
		if answerErr != nil {
			outcome = "failure"
		}
		h.logger.Info("telegram callback result", "event", "telegram.callback_result", "operation", "telegram", "chat_id", chatID, "user_id", userID, "callback_action", action, "outcome", outcome)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	text, keyboard, outcome, err := h.HandleCallbackForChat(r.Context(), action, userID, messageID, chatID)
	if err != nil {
		outcome = "failure"
		text = ErrorMessageLang(lang, err)
	}
	if text != "" && h.telegram != nil {
		if err := h.sendMessageWithKeyboard(r.Context(), chatID, text, keyboard); err != nil {
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
	return h.HandleCallbackForChat(ctx, action, userID, messageID, h.groupChat)
}

func (h *Handler) HandleCallbackForChat(ctx context.Context, action string, userID int64, messageID int, chatID int64) (string, *InlineKeyboardMarkup, string, error) {
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

func (h *Handler) authorize(ctx context.Context, chat Chat, command string) groups.GroupAuthorizationDecision {
	if h.groups == nil {
		if h.groupChat != 0 && chat.ID == h.groupChat {
			return groups.GroupAuthorizationDecision{ChatID: chat.ID, Command: command, Allowed: true, Reason: groups.ReasonFallback}
		}
		return groups.GroupAuthorizationDecision{ChatID: chat.ID, Command: command, Reason: groups.ReasonUnknownGroup}
	}
	return h.groups.Authorize(ctx, chat.ID, chat.Type, command)
}

func (h *Handler) logAuthorizationRejected(chatID, userID int64, command, reason string) {
	h.logger.Warn("telegram group authorization", "event", "telegram.group_authorization", "operation", "telegram", "chat_id", chatID, "user_id", userID, "command", command, "outcome", "rejected", "reason", reason)
}

func (h *Handler) logGroupSetup(chatID, userID int64, outcome, reason string) {
	attrs := []any{"event", "telegram.group_setup", "operation", "telegram", "chat_id", chatID, "user_id", userID, "outcome", outcome}
	if reason != "" {
		attrs = append(attrs, "reason", reason)
	}
	h.logger.Info("telegram group setup", attrs...)
}

func commandForCallback(action string) string {
	switch action {
	case CallbackMenuAddPerson:
		return CommandAddPerson
	case CallbackMenuStartWork:
		return CommandStartWork
	case CallbackMenuStatus:
		return CommandStatus
	case CallbackMenuStopWork:
		return CommandStopWork
	case CallbackMenuReportMonth:
		return CommandReportMonth
	case CallbackMenuHelp:
		return CommandHelp
	case CallbackMenuSettings, CallbackSettingsLanguage, CallbackLanguageUK, CallbackLanguageEN, CallbackLanguageRU:
		return "/settings"
	default:
		return "/callback"
	}
}

func chatTypeFromID(chatID int64) string {
	return ""
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
