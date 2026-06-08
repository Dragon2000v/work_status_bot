# Quickstart: Structured Application Logging

## Prerequisites

- Go 1.22+
- Existing bot environment from `.env.example`
- New logging environment values:

```bash
APP_ENV=local
LOG_LEVEL=debug
```

## Run Tests

```bash
go test ./...
```

Expected:
- Tests pass.
- Logging tests confirm selected log level, HTTP request fields, Telegram event fields,
  cron event fields, and secret redaction behavior.

## Validate Local Startup Logs

Run locally with placeholder-safe development values:

```bash
go run ./cmd/bot
```

Expected logs include:
- application startup event
- config loaded event with `APP_ENV` and `LOG_LEVEL`
- MongoDB connection attempt and outcome
- MongoDB index setup outcome
- listen address

Expected logs do not include:
- full MongoDB URI
- MongoDB password
- Telegram bot token
- webhook secret
- cron secret

## Validate HTTP Request Logs

```bash
curl http://localhost:8080/health
```

Expected log fields:
- `event=http.request`
- `method=GET`
- `path=/health`
- `status=200`
- `duration_ms`
- `remote_addr`

## Validate Telegram Webhook Logs

Send a test webhook update for a supported command from the configured chat.

Expected log fields:
- `event=telegram.webhook_received`
- `chat_id`
- `user_id` when present
- `command`
- command success or failure outcome

Expected logs do not include:
- full raw Telegram payload
- full message body beyond command name
- Telegram bot token

## Validate Cron Logs

```bash
curl -X POST http://localhost:8080/cron/monthly-report \
  -H "Content-Type: application/json" \
  -H "X-Cron-Secret: $CRON_SECRET" \
  -d '{"month":"2026-06"}'
```

Expected log fields:
- monthly report trigger event
- `month=2026-06`
- generated, duplicate, skipped, invalid, unauthorized, or failed outcome
- report send outcome when the trigger sends to Telegram

Expected logs do not include:
- cron secret
- `X-Cron-Secret` value

## Validate Production Format

Set:

```bash
APP_ENV=production
LOG_LEVEL=info
```

Run the service and trigger `/health`.

Expected:
- Logs are structured JSON.
- Request fields are present and parseable.
- Debug logs are omitted when `LOG_LEVEL=info`.
