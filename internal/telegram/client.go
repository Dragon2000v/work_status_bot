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
	body, err := json.Marshal(map[string]any{"chat_id": chatID, "text": text})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errors.New("telegram sendMessage request build failed")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", chatID, "outcome", "failure", "error", "request failed")
		return errors.New("telegram sendMessage request failed")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.logger.Error("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", chatID, "outcome", "failure", "status", res.StatusCode)
		return fmt.Errorf("telegram sendMessage failed: %s", res.Status)
	}
	c.logger.Info("telegram send", "event", "telegram.send", "operation", "telegram_send", "chat_id", chatID, "outcome", "success")
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
