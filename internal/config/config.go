package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	AppEnv                 string
	AppAddr                string
	LogLevel               string
	MongoDBURI             string
	MongoDBDatabase        string
	TelegramBotToken       string
	TelegramGroupChatID    int64
	TelegramWebhookSecret  string
	TelegramWebhookURL     string
	TelegramAutoSetWebhook bool
	CronSecret             string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                getenv("APP_ENV", "local"),
		AppAddr:               getenv("APP_ADDR", ":8080"),
		LogLevel:              getenv("LOG_LEVEL", "info"),
		MongoDBURI:            os.Getenv("MONGODB_URI"),
		MongoDBDatabase:       os.Getenv("MONGODB_DATABASE"),
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		TelegramWebhookURL:    os.Getenv("TELEGRAM_WEBHOOK_URL"),
		CronSecret:            os.Getenv("CRON_SECRET"),
	}

	chatID := os.Getenv("TELEGRAM_GROUP_CHAT_ID")
	if chatID != "" {
		v, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return Config{}, err
		}
		cfg.TelegramGroupChatID = v
	}
	autoSetWebhook := getenv("TELEGRAM_AUTO_SET_WEBHOOK", "false")
	switch autoSetWebhook {
	case "true":
		cfg.TelegramAutoSetWebhook = true
	case "false":
		cfg.TelegramAutoSetWebhook = false
	default:
		return Config{}, errors.New("TELEGRAM_AUTO_SET_WEBHOOK must be true or false")
	}

	if cfg.MongoDBURI == "" || cfg.MongoDBDatabase == "" || cfg.TelegramBotToken == "" || cfg.TelegramGroupChatID == 0 || cfg.TelegramWebhookSecret == "" || cfg.CronSecret == "" {
		return Config{}, errors.New("missing required environment configuration")
	}
	if cfg.TelegramAutoSetWebhook && cfg.TelegramWebhookURL == "" {
		return Config{}, errors.New("TELEGRAM_WEBHOOK_URL is required when TELEGRAM_AUTO_SET_WEBHOOK is true")
	}
	if cfg.AppEnv != "local" && cfg.AppEnv != "production" {
		return Config{}, errors.New("APP_ENV must be local or production")
	}
	if cfg.LogLevel != "debug" && cfg.LogLevel != "info" && cfg.LogLevel != "warn" && cfg.LogLevel != "error" {
		return Config{}, errors.New("LOG_LEVEL must be debug, info, warn, or error")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
