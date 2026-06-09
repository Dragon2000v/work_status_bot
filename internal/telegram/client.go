package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
	IsPersistent    bool               `json:"is_persistent,omitempty"`
}

type KeyboardButton struct {
	Text string `json:"text"`
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

func NewClient(token string, httpClient *http.Client) *Client {
	return NewClientWithLogger(token, httpClient, slog.Default())
}

func NewClientWithLogger(token string, httpClient *http.Client, log *slog.Logger) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if log == nil {
		log = slog.Default()
	}
	return &Client{token: token, httpClient: httpClient, baseURL: "https://api.telegram.org", logger: log}
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.SendMessageWithKeyboard(ctx, chatID, text, nil)
}

func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard any) error {
	payload := map[string]any{"chat_id": chatID, "text": text}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "sendMessage", body, "telegram.send", "telegram_send", "chat_id", chatID)
}

func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	payload := map[string]any{"callback_query_id": callbackQueryID}
	if text != "" {
		payload["text"] = text
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "answerCallbackQuery", body, "telegram.callback_answer", "telegram_send")
}

func (c *Client) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	body, err := json.Marshal(map[string]any{"chat_id": chatID, "message_id": messageID})
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "deleteMessage", body, "telegram.command_delete", "telegram_send", "chat_id", chatID, "message_id", messageID)
}

func (c *Client) SetMyCommands(ctx context.Context, commands []BotCommand) error {
	c.logger.Info("telegram command menu registration", "event", "telegram.command_menu_registration", "operation", "telegram_command_menu", "outcome", "attempt", "command_count", len(commands))
	body, err := json.Marshal(map[string]any{"commands": commands})
	if err != nil {
		c.logger.Error("telegram command menu registration", "event", "telegram.command_menu_registration", "operation", "telegram_command_menu", "outcome", "failure", "error", err.Error())
		return err
	}
	if err := c.postJSON(ctx, "setMyCommands", body, "telegram.command_menu_registration", "telegram_command_menu", "command_count", len(commands)); err != nil {
		return err
	}
	return nil
}

func (c *Client) EditMessageText(ctx context.Context, chatID int64, messageID int, text string, keyboard *InlineKeyboardMarkup) error {
	payload := map[string]any{"chat_id": chatID, "message_id": messageID, "text": text}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "editMessageText", body, "telegram.edit_message", "telegram_send", "chat_id", chatID, "message_id", messageID)
}

func (c *Client) postJSON(ctx context.Context, method string, body []byte, event, operation string, attrs ...any) error {
	url := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram %s request build failed", method)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	logAttrs := []any{"event", event, "operation", operation, "outcome", "failure"}
	logAttrs = append(logAttrs, attrs...)
	if err != nil {
		c.logger.Error("telegram api", append(logAttrs, "error", "request failed")...)
		return fmt.Errorf("telegram %s request failed", method)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.logger.Error("telegram api", append(logAttrs, "status", res.StatusCode)...)
		return fmt.Errorf("telegram %s failed: %s", method, res.Status)
	}
	successAttrs := []any{"event", event, "operation", operation, "outcome", "success"}
	successAttrs = append(successAttrs, attrs...)
	c.logger.Info("telegram api", successAttrs...)
	return nil
}

func (c *Client) SetWebhook(ctx context.Context, webhookURL, secretToken string) error {
	c.logger.Info("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "attempt")
	body, err := json.Marshal(map[string]any{"url": webhookURL, "secret_token": secretToken})
	if err != nil {
		c.logger.Error("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "failure", "error", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/bot%s/setWebhook", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		c.logger.Error("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "failure", "error", "request build failed")
		return errors.New("telegram setWebhook request build failed")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "failure", "error", "request failed")
		return errors.New("telegram setWebhook request failed")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.logger.Error("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "failure", "status", res.StatusCode)
		return fmt.Errorf("telegram setWebhook failed: %s", res.Status)
	}
	c.logger.Info("telegram webhook registration", "event", "telegram.webhook_registration", "operation", "telegram_webhook", "outcome", "success")
	return nil
}
