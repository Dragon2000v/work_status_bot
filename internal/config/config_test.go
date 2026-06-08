package config

import "testing"

func TestLoadLoggingDefaults(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.net")
	t.Setenv("MONGODB_DATABASE", "work_status_bot")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_GROUP_CHAT_ID", "-100")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "webhook")
	t.Setenv("CRON_SECRET", "cron")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppEnv != "local" || cfg.LogLevel != "info" {
		t.Fatalf("bad defaults: %#v", cfg)
	}
	if cfg.TelegramAutoSetWebhook {
		t.Fatalf("auto webhook should default false")
	}
}

func TestLoadLoggingInvalidValues(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.net")
	t.Setenv("MONGODB_DATABASE", "work_status_bot")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_GROUP_CHAT_ID", "-100")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "webhook")
	t.Setenv("CRON_SECRET", "cron")
	t.Setenv("APP_ENV", "staging")
	if _, err := Load(); err == nil {
		t.Fatalf("want APP_ENV error")
	}
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "trace")
	if _, err := Load(); err == nil {
		t.Fatalf("want LOG_LEVEL error")
	}
}

func TestLoadTelegramWebhookConfig(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.net")
	t.Setenv("MONGODB_DATABASE", "work_status_bot")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_GROUP_CHAT_ID", "-100")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "webhook")
	t.Setenv("CRON_SECRET", "cron")
	t.Setenv("TELEGRAM_AUTO_SET_WEBHOOK", "true")
	t.Setenv("TELEGRAM_WEBHOOK_URL", "https://example.com/telegram/webhook")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TelegramAutoSetWebhook || cfg.TelegramWebhookURL != "https://example.com/telegram/webhook" {
		t.Fatalf("bad webhook config: %#v", cfg)
	}
}

func TestLoadTelegramWebhookInvalidConfig(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.net")
	t.Setenv("MONGODB_DATABASE", "work_status_bot")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_GROUP_CHAT_ID", "-100")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "webhook")
	t.Setenv("CRON_SECRET", "cron")
	t.Setenv("TELEGRAM_AUTO_SET_WEBHOOK", "maybe")
	if _, err := Load(); err == nil {
		t.Fatalf("want invalid auto webhook error")
	}
	t.Setenv("TELEGRAM_AUTO_SET_WEBHOOK", "true")
	if _, err := Load(); err == nil {
		t.Fatalf("want missing webhook url error")
	}
}
