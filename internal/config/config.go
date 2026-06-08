package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	AppAddr               string
	MongoDBURI            string
	MongoDBDatabase       string
	TelegramBotToken      string
	TelegramGroupChatID   int64
	TelegramWebhookSecret string
	CronSecret            string
}

func Load() (Config, error) {
	cfg := Config{
		AppAddr:               getenv("APP_ADDR", ":8080"),
		MongoDBURI:            os.Getenv("MONGODB_URI"),
		MongoDBDatabase:       os.Getenv("MONGODB_DATABASE"),
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
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

	if cfg.MongoDBURI == "" || cfg.MongoDBDatabase == "" || cfg.TelegramBotToken == "" || cfg.TelegramGroupChatID == 0 || cfg.TelegramWebhookSecret == "" || cfg.CronSecret == "" {
		return Config{}, errors.New("missing required environment configuration")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
