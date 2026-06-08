# Quickstart: Telegram Team Work Status Tracker

## Prerequisites

- Go 1.22+
- MongoDB Atlas connection string
- Telegram bot token
- Telegram group chat ID for commands and alerts
- Public HTTPS URL for Telegram webhook delivery
- External scheduler capable of sending `POST /cron/monthly-report`

## Configuration

Set environment variables:

```bash
APP_ADDR=:8080
MONGODB_URI=mongodb+srv://...
MONGODB_DATABASE=work_status_bot
TELEGRAM_BOT_TOKEN=123456:token
TELEGRAM_GROUP_CHAT_ID=-1001234567890
TELEGRAM_WEBHOOK_SECRET=change-me
CRON_SECRET=change-me
APP_PUBLIC_URL=https://example.com
```

## Run Locally

```bash
go mod tidy
go test ./...
go run ./cmd/bot
```

## Validate Health

```bash
curl http://localhost:8080/health
```

Expected:

```json
{"status":"ok"}
```

## Configure Telegram Webhook

Use Telegram Bot API to point updates at:

```text
https://<public-host>/telegram/webhook
```

Webhook delivery is required. Do not use polling.

## Validate Telegram Commands

In the configured Telegram group:

```text
/help
/add_person Ivan Petrenko
/start_work Ivan Petrenko "API fix"
/status
/stop_work Ivan Petrenko Blocked by dependency
/report_month
```

Expected:
- `/help` lists commands.
- `/add_person` confirms the person.
- `/start_work` creates active work with Europe/Kyiv display time.
- `/status` shows elapsed time calculated from stored timestamps.
- `/stop_work` stores stop time, creates an incident, and sends an immediate group alert.
- `/report_month` lists people, active work records, elapsed time, and incidents.

## Validate Duplicate Monthly Report Prevention

Run:

```text
/report_month 2026-06
/report_month 2026-06
```

Expected:
- First command generates/stores the report.
- Second command returns existing report or duplicate notice.
- No second report record is created.

## Validate External Monthly Report Trigger

```bash
curl -X POST http://localhost:8080/cron/monthly-report \
  -H 'Content-Type: application/json' \
  -d '{"month":"2026-06"}'
```

Expected:
- Response is `generated` when no report exists.
- Response is `duplicate` when report already exists.
- Telegram group receives report or duplicate notice.
- No in-process cron loop, background infinite loop, or in-memory timer is required.

## Validate Sleep-Friendly Behavior

1. Start work from Telegram.
2. Let hosting sleep or stop the local process.
3. Restart service.
4. Run `/status` or `/report_month`.

Expected:
- Elapsed time still matches stored start/stop timestamps.
- Database timestamps remain UTC.
- User-facing dates display in Europe/Kyiv timezone.
