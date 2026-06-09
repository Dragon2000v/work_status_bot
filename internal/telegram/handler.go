package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"work-status-bot/internal/datetime"
	"work-status-bot/internal/flows"
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

type WorkStartAtService interface {
	StartAt(ctx context.Context, firstName, lastName, title string, startedAt time.Time) (people.Person, works.WorkRecord, error)
}

type FlowService interface {
	Start(ctx context.Context, userID, chatID int64, flowType, step string, payload map[string]string) (flows.State, error)
	Get(ctx context.Context, userID, chatID int64) (flows.State, error)
	Advance(ctx context.Context, state flows.State, step string, payload map[string]string) (flows.State, error)
	Complete(ctx context.Context, userID, chatID int64) error
	Cancel(ctx context.Context, userID, chatID int64) error
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
	flows    FlowService
	telegram telegramAPI
	logger   *slog.Logger
	now      func() time.Time
}

type telegramAPI interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type keyboardSender interface {
	SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard any) error
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

func (h *Handler) SetFlows(flowSvc FlowService) {
	h.flows = flowSvc
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
	messageKind := h.cleanupKind(r.Context(), update.Message.Text, userID, update.Message.Chat.ID)
	response, err := h.HandleTextForChat(r.Context(), update.Message.Text, userID, username, update.Message.Chat)
	if err != nil {
		h.logger.Error("telegram command result", "event", "telegram.command_result", "operation", "telegram", "chat_id", update.Message.Chat.ID, "user_id", userID, "command", command, "outcome", "failure", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response != "" && h.telegram != nil {
		lang := h.languageForUser(r.Context(), userID)
		if keyboard := h.keyboardForResponse(response, lang); keyboard != nil {
			if err := h.sendMessageWithKeyboard(r.Context(), update.Message.Chat.ID, response, keyboard); err != nil {
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
	h.cleanupUserMessage(r.Context(), update.Message.Chat.ID, update.Message.MessageID, userID, command, messageKind)
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
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		if response, handled, err := h.handleActiveFlow(ctx, trimmed, userID, chat, lang); handled || err != nil {
			return response, err
		}
		if action, ok := CommandForKeyboardText(trimmed); ok {
			h.logger.Info("telegram reply keyboard action", "event", "telegram.reply_keyboard_action", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "action", action, "outcome", "received")
			return h.handleCanonicalAction(ctx, action, userID, chat, lang)
		}
		return i18n.T(lang, i18n.KeyUnsupportedAction), nil
	}
	if response, handled, err := h.handleBareFlowCommand(ctx, trimmed, userID, chat, lang); handled || err != nil {
		return response, err
	}
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
		return HelpMessageLang(lang), nil, "success", nil
	case CallbackMenuSettings:
		return SettingsMessage(lang), InlineSettingsKeyboard(lang), "success", nil
	case CallbackSettingsLanguage:
		return LanguageMessage(lang), InlineLanguageKeyboard(lang), "success", nil
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
		return i18n.T(selected, i18n.KeyLanguageChanged), nil, "success", nil
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction), nil, "unsupported", nil
	}
}

func (h *Handler) handleCanonicalAction(ctx context.Context, action string, userID int64, chat Chat, lang i18n.Language) (string, error) {
	decision := h.authorize(ctx, chat, action)
	if !decision.Allowed {
		h.logAuthorizationRejected(chat.ID, userID, action, decision.Reason)
		if decision.Reason == groups.ReasonPrivateChatSetup {
			return SetupGroupOnlyMessage(lang), nil
		}
		return SetupRequiredMessage(lang), nil
	}
	switch action {
	case CommandAddPerson:
		return h.startFlow(ctx, userID, chat.ID, flows.TypeAddPerson, flows.StepAddPersonEnterName, nil, i18n.T(lang, i18n.KeyAddPersonEnterName))
	case CommandStartWork:
		return h.startFlow(ctx, userID, chat.ID, flows.TypeStartWork, flows.StepStartWorkEnterPerson, nil, i18n.T(lang, i18n.KeyStartWorkEnterPerson))
	case CommandStopWork:
		return h.startFlow(ctx, userID, chat.ID, flows.TypeStopWork, flows.StepStopWorkEnterPersonOrReason, nil, i18n.T(lang, i18n.KeyStopWorkEnterInput))
	case "/settings":
		return h.startFlow(ctx, userID, chat.ID, flows.TypeSettings, flows.StepSettingsLanguageSelect, nil, LanguageMessage(lang))
	case CommandHelp:
		return HelpMessageLang(lang), nil
	case CommandStatus:
		persons, records, err := h.work.Status(ctx)
		if err != nil {
			return "", err
		}
		return StatusMessageLang(i18n.Ukrainian, persons, records, h.now()), nil
	case CommandReportMonth:
		if h.reports == nil {
			return "", errors.New("reports service not configured")
		}
		result, err := h.reports.GenerateMonthly(ctx, "", reports.TriggerTelegramCommand)
		if err != nil {
			return ErrorMessageLang(lang, err), nil
		}
		if h.telegram != nil {
			if err := h.telegram.SendMessage(ctx, chat.ID, ReportMessage(result)); err != nil {
				return "", err
			}
			return "", nil
		}
		return ReportMessage(result), nil
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction), nil
	}
}

func (h *Handler) handleBareFlowCommand(ctx context.Context, text string, userID int64, chat Chat, lang i18n.Language) (string, bool, error) {
	fields := splitCommand(text)
	if len(fields) != 1 {
		return "", false, nil
	}
	action := normalizeCommandName(fields[0])
	switch action {
	case CommandAddPerson, CommandStartWork, CommandStopWork:
		response, err := h.handleCanonicalAction(ctx, action, userID, chat, lang)
		return response, true, err
	default:
		return "", false, nil
	}
}

func (h *Handler) startFlow(ctx context.Context, userID, chatID int64, flowType, step string, payload map[string]string, response string) (string, error) {
	if h.flows == nil {
		return response, nil
	}
	if _, err := h.flows.Start(ctx, userID, chatID, flowType, step, payload); err != nil {
		return "", err
	}
	h.logger.Info("telegram flow started", "event", "telegram.flow_started", "operation", "telegram", "chat_id", chatID, "user_id", userID, "flow_type", flowType, "step", step, "outcome", "success")
	return response, nil
}

func (h *Handler) handleActiveFlow(ctx context.Context, text string, userID int64, chat Chat, lang i18n.Language) (string, bool, error) {
	if h.flows == nil {
		return "", false, nil
	}
	state, err := h.flows.Get(ctx, userID, chat.ID)
	if errors.Is(err, flows.ErrNotFound) {
		return "", false, nil
	}
	if errors.Is(err, flows.ErrExpired) {
		h.logger.Info("telegram flow expired", "event", "telegram.flow_expired", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "outcome", "expired")
		return i18n.T(lang, i18n.KeyFlowExpired), true, nil
	}
	if err != nil {
		return "", true, err
	}
	if isCancelText(text) {
		_ = h.flows.Cancel(ctx, userID, chat.ID)
		h.logger.Info("telegram flow cancelled", "event", "telegram.flow_cancelled", "operation", "telegram", "chat_id", chat.ID, "user_id", userID, "flow_type", state.FlowType, "outcome", "success")
		return i18n.T(lang, i18n.KeyFlowCancelled), true, nil
	}
	switch state.Step {
	case flows.StepAddPersonEnterName:
		return h.handleAddPersonFlow(ctx, state, text, lang)
	case flows.StepStopWorkEnterPersonOrReason:
		return h.handleStopWorkFlow(ctx, state, text, chat, lang)
	case flows.StepSettingsLanguageSelect:
		return h.handleSettingsFlow(ctx, state, text, lang)
	case flows.StepStartWorkEnterPerson, flows.StepStartWorkEnterTitle, flows.StepStartWorkSelectTimeMode, flows.StepStartWorkEnterTime, flows.StepStartWorkEnterManualDateTime:
		return h.handleStartWorkFlow(ctx, state, text, lang)
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction), true, nil
	}
}

func (h *Handler) handleAddPersonFlow(ctx context.Context, state flows.State, text string, lang i18n.Language) (string, bool, error) {
	first, last, _, ok := parseNameAndRest(text)
	if !ok {
		return i18n.T(lang, i18n.KeyAddPersonInvalidName), true, nil
	}
	p, err := h.people.Add(ctx, first, last)
	if err != nil {
		return ErrorMessageLang(lang, err), true, nil
	}
	_ = h.flows.Complete(ctx, state.TelegramUserID, state.ChatID)
	h.logger.Info("telegram flow completed", "event", "telegram.flow_completed", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "outcome", "success")
	return PersonAddedMessageLang(lang, p), true, nil
}

func (h *Handler) handleStopWorkFlow(ctx context.Context, state flows.State, text string, chat Chat, lang i18n.Language) (string, bool, error) {
	first, last, reason, ok := parseNameAndRest(text)
	if !ok {
		return i18n.T(lang, i18n.KeyAddPersonInvalidName), true, nil
	}
	cmd := Command{Name: CommandStopWork, FirstName: first, LastName: last, Reason: reason}
	outcome := "success"
	response, err := h.executeFlowCommand(ctx, cmd, chat, lang, &outcome)
	if err == nil && outcome == "success" {
		_ = h.flows.Complete(ctx, state.TelegramUserID, state.ChatID)
		h.logger.Info("telegram flow completed", "event", "telegram.flow_completed", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "outcome", "success")
	}
	return response, true, err
}

func (h *Handler) handleSettingsFlow(ctx context.Context, state flows.State, text string, lang i18n.Language) (string, bool, error) {
	if isBackText(text) {
		return SettingsMessage(lang), true, nil
	}
	selected, ok := languageFromText(text)
	if !ok {
		return LanguageMessage(lang), true, nil
	}
	if h.users != nil {
		if _, err := h.users.SetLanguage(ctx, state.TelegramUserID, selected); err != nil {
			h.logger.Error("telegram user language changed", "event", "telegram.user_language_changed", "operation", "telegram", "user_id", state.TelegramUserID, "language", selected, "outcome", "failure", "error", err.Error())
			return "", true, err
		}
	}
	_ = h.flows.Complete(ctx, state.TelegramUserID, state.ChatID)
	h.logger.Info("telegram user language changed", "event", "telegram.user_language_changed", "operation", "telegram", "user_id", state.TelegramUserID, "language", selected, "outcome", "success")
	return i18n.T(selected, i18n.KeyLanguageChanged), true, nil
}

func (h *Handler) handleStartWorkFlow(ctx context.Context, state flows.State, text string, lang i18n.Language) (string, bool, error) {
	payload := state.Payload
	if payload == nil {
		payload = map[string]string{}
	}
	switch state.Step {
	case flows.StepStartWorkEnterPerson:
		first, last, _, ok := parseNameAndRest(text)
		if !ok {
			return i18n.T(lang, i18n.KeyAddPersonInvalidName), true, nil
		}
		payload["first_name"], payload["last_name"] = first, last
		if _, err := h.flows.Advance(ctx, state, flows.StepStartWorkEnterTitle, payload); err != nil {
			return "", true, err
		}
		h.logger.Info("telegram flow step advanced", "event", "telegram.flow_step_advanced", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "step", flows.StepStartWorkEnterTitle, "outcome", "success")
		return i18n.T(lang, i18n.KeyStartWorkEnterTitle), true, nil
	case flows.StepStartWorkEnterTitle:
		title := strings.TrimSpace(text)
		if title == "" {
			return i18n.T(lang, i18n.KeyStartWorkEnterTitle), true, nil
		}
		payload["title"] = title
		if _, err := h.flows.Advance(ctx, state, flows.StepStartWorkSelectTimeMode, payload); err != nil {
			return "", true, err
		}
		h.logger.Info("telegram flow step advanced", "event", "telegram.flow_step_advanced", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "step", flows.StepStartWorkSelectTimeMode, "outcome", "success")
		return i18n.T(lang, i18n.KeyStartWorkSelectTime), true, nil
	case flows.StepStartWorkSelectTimeMode:
		mode, prompt, ok := startModeFromText(text, lang)
		if !ok {
			return i18n.T(lang, i18n.KeyStartWorkSelectTime), true, nil
		}
		payload["time_mode"] = string(mode)
		if mode == datetime.ModeNow {
			return h.completeStartWork(ctx, state, payload, h.now(), lang)
		}
		next := flows.StepStartWorkEnterTime
		if mode == datetime.ModeManual {
			next = flows.StepStartWorkEnterManualDateTime
		}
		if _, err := h.flows.Advance(ctx, state, next, payload); err != nil {
			return "", true, err
		}
		h.logger.Info("telegram flow step advanced", "event", "telegram.flow_step_advanced", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "step", next, "outcome", "success")
		return prompt, true, nil
	case flows.StepStartWorkEnterTime, flows.StepStartWorkEnterManualDateTime:
		mode := datetime.Mode(payload["time_mode"])
		startedAt, err := datetime.ParseStart(text, mode, h.now())
		if err != nil {
			h.logDateParse(state, "failure", err)
			return DateTimeErrorMessage(lang, err), true, nil
		}
		h.logDateParse(state, "success", nil)
		return h.completeStartWork(ctx, state, payload, startedAt, lang)
	default:
		return i18n.T(lang, i18n.KeyUnsupportedAction), true, nil
	}
}

func (h *Handler) completeStartWork(ctx context.Context, state flows.State, payload map[string]string, startedAt time.Time, lang i18n.Language) (string, bool, error) {
	var p people.Person
	var r works.WorkRecord
	var err error
	if starter, ok := h.work.(WorkStartAtService); ok {
		p, r, err = starter.StartAt(ctx, payload["first_name"], payload["last_name"], payload["title"], startedAt)
	} else {
		p, r, err = h.work.Start(ctx, payload["first_name"], payload["last_name"], payload["title"])
	}
	if err != nil {
		return ErrorMessageLang(lang, err), true, nil
	}
	_ = h.flows.Complete(ctx, state.TelegramUserID, state.ChatID)
	h.logger.Info("telegram flow completed", "event", "telegram.flow_completed", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "outcome", "success")
	return WorkStartedMessageLang(lang, p, r), true, nil
}

func (h *Handler) executeFlowCommand(ctx context.Context, cmd Command, chat Chat, lang i18n.Language, outcome *string) (string, error) {
	switch cmd.Name {
	case CommandStopWork:
		p, record, err := h.work.Stop(ctx, cmd.FirstName, cmd.LastName, cmd.Reason)
		if err != nil {
			*outcome = "failure"
			return ErrorMessageLang(lang, err), nil
		}
		occurred := h.now()
		if record.StoppedAt != nil {
			occurred = *record.StoppedAt
		}
		if h.reports != nil {
			if err := h.reports.RecordStopped(ctx, p.ID, record.ID, cmd.Reason, occurred); err != nil {
				*outcome = "failure"
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
						*outcome = "failure"
						return "", recErr
					}
				}
			} else {
				h.logger.Info("telegram alert send", "event", "telegram.alert_send", "operation", "telegram_send", "chat_id", chat.ID, "outcome", "success")
			}
		}
		return WorkStoppedMessageLang(lang, p, record, warning), nil
	default:
		*outcome = "failure"
		return ErrorMessageLang(lang, ErrMalformedCommand), nil
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

func (h *Handler) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard any) error {
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

func (h *Handler) keyboardForResponse(response string, lang i18n.Language) any {
	switch response {
	case HelpMessageLang(lang), MenuMessage(lang), i18n.T(lang, i18n.KeyLanguageChanged), i18n.T(lang, i18n.KeyFlowCancelled), i18n.T(lang, i18n.KeyFlowExpired):
		return MainMenuKeyboard(i18n.Ukrainian)
	case SettingsMessage(lang):
		return SettingsKeyboard(lang)
	case LanguageMessage(lang):
		return LanguageKeyboard(lang)
	case i18n.T(lang, i18n.KeyStartWorkSelectTime):
		return StartTimeModeKeyboard(lang)
	default:
		return nil
	}
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

func (h *Handler) cleanupKind(ctx context.Context, text string, userID, chatID int64) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "/") {
		return "slash_command"
	}
	if _, ok := CommandForKeyboardText(text); ok {
		return "reply_keyboard"
	}
	if h.flows != nil {
		if _, err := h.flows.Get(ctx, userID, chatID); err == nil || errors.Is(err, flows.ErrExpired) {
			return "flow_input"
		}
	}
	return ""
}

func (h *Handler) cleanupUserMessage(ctx context.Context, chatID int64, messageID int, userID int64, command, kind string) {
	if messageID == 0 || kind == "" {
		return
	}
	deleter, ok := h.telegram.(messageDeleter)
	if !ok {
		return
	}
	if kind == "slash_command" {
		h.logger.Info("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "attempt")
	}
	h.logger.Info("telegram message delete", "event", "telegram.message_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "message_kind", kind, "outcome", "attempt")
	if err := deleter.DeleteMessage(ctx, chatID, messageID); err != nil {
		if kind == "slash_command" {
			h.logger.Warn("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "failure", "error", err.Error())
		}
		h.logger.Warn("telegram message delete", "event", "telegram.message_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "message_kind", kind, "outcome", "failure", "error", err.Error())
		return
	}
	if kind == "slash_command" {
		h.logger.Info("telegram command delete", "event", "telegram.command_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "outcome", "success")
	}
	h.logger.Info("telegram message delete", "event", "telegram.message_delete", "operation", "telegram_send", "chat_id", chatID, "user_id", userID, "message_id", messageID, "command", command, "message_kind", kind, "outcome", "success")
}

func commandName(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return ""
	}
	return normalizeCommandName(fields[0])
}

func parseNameAndRest(text string) (string, string, string, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) < 2 {
		return "", "", "", false
	}
	rest := ""
	if len(fields) > 2 {
		rest = strings.Join(fields[2:], " ")
	}
	return fields[0], fields[1], rest, true
}

func languageFromText(text string) (i18n.Language, bool) {
	for _, lang := range []i18n.Language{i18n.Ukrainian, i18n.English, i18n.Russian} {
		switch text {
		case i18n.T(lang, i18n.KeyLanguageUkrainian):
			return i18n.Ukrainian, true
		case i18n.T(lang, i18n.KeyLanguageEnglish):
			return i18n.English, true
		case i18n.T(lang, i18n.KeyLanguageRussian):
			return i18n.Russian, true
		}
	}
	return i18n.Ukrainian, false
}

func isCancelText(text string) bool {
	for _, lang := range []i18n.Language{i18n.Ukrainian, i18n.English, i18n.Russian} {
		if text == i18n.T(lang, i18n.KeyCancel) {
			return true
		}
	}
	return false
}

func isBackText(text string) bool {
	for _, lang := range []i18n.Language{i18n.Ukrainian, i18n.English, i18n.Russian} {
		if text == i18n.T(lang, i18n.KeyBack) {
			return true
		}
	}
	return false
}

func startModeFromText(text string, lang i18n.Language) (datetime.Mode, string, bool) {
	switch text {
	case i18n.T(lang, i18n.KeyStartNow), i18n.T(i18n.Ukrainian, i18n.KeyStartNow), i18n.T(i18n.English, i18n.KeyStartNow), i18n.T(i18n.Russian, i18n.KeyStartNow):
		return datetime.ModeNow, "", true
	case i18n.T(lang, i18n.KeyStartToday), i18n.T(i18n.Ukrainian, i18n.KeyStartToday), i18n.T(i18n.English, i18n.KeyStartToday), i18n.T(i18n.Russian, i18n.KeyStartToday):
		return datetime.ModeToday, i18n.T(lang, i18n.KeyStartWorkEnterTime), true
	case i18n.T(lang, i18n.KeyStartYesterday), i18n.T(i18n.Ukrainian, i18n.KeyStartYesterday), i18n.T(i18n.English, i18n.KeyStartYesterday), i18n.T(i18n.Russian, i18n.KeyStartYesterday):
		return datetime.ModeYesterday, i18n.T(lang, i18n.KeyStartWorkEnterTime), true
	case i18n.T(lang, i18n.KeyStartManual), i18n.T(i18n.Ukrainian, i18n.KeyStartManual), i18n.T(i18n.English, i18n.KeyStartManual), i18n.T(i18n.Russian, i18n.KeyStartManual):
		return datetime.ModeManual, i18n.T(lang, i18n.KeyStartWorkEnterManual), true
	default:
		return "", "", false
	}
}

func (h *Handler) logDateParse(state flows.State, outcome string, err error) {
	attrs := []any{"event", "telegram.datetime_parse", "operation", "telegram", "chat_id", state.ChatID, "user_id", state.TelegramUserID, "flow_type", state.FlowType, "outcome", outcome}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	h.logger.Info("telegram datetime parse", attrs...)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
