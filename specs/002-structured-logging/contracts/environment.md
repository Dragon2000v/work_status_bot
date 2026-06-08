# Environment Contract

Logging adds two environment variables and keeps all existing bot configuration names.

## New Variables

```bash
APP_ENV=local
LOG_LEVEL=info
```

## APP_ENV

Allowed values:
- `local`
- `production`

Expected behavior:
- `local`: readable structured text logs for local development.
- `production`: structured JSON logs for hosting and production log collection.

## LOG_LEVEL

Allowed values:
- `debug`
- `info`
- `warn`
- `error`

Expected behavior:
- Logs below the selected level are omitted.
- Invalid values must be handled explicitly and safely.

## Existing Variables

Existing variables remain unchanged:

```bash
APP_ADDR=:8080
MONGODB_URI=mongodb+srv://<username>:<password>@<cluster-url>/?retryWrites=true&w=majority
MONGODB_DATABASE=work_status_bot
TELEGRAM_BOT_TOKEN=<telegram-bot-token>
TELEGRAM_GROUP_CHAT_ID=<telegram-group-chat-id>
TELEGRAM_WEBHOOK_SECRET=<random-secret>
CRON_SECRET=<random-secret>
```

## Secret Logging Rule

The application must never log:
- `MONGODB_URI`
- MongoDB password
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_WEBHOOK_SECRET`
- `CRON_SECRET`
- `X-Cron-Secret`
- full environment dumps
