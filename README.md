# Work Status Bot

Go Telegram webhook bot for team work status tracking.

## Constraints

- Telegram webhook only. No Telegram polling.
- MongoDB Atlas is database of record.
- UTC timestamps in database.
- Europe/Kyiv dates in Telegram messages and reports.
- Elapsed time calculated from stored timestamps only.
- No Redis, queues, WebSockets, in-memory timers, background infinite loops, or MVP web UI.

## Environment

Copy `.env.example` to `.env` for local development and replace placeholders with local values. Do not commit real secrets.

```bash
APP_ADDR=:8080
MONGODB_URI=mongodb+srv://<username>:<password>@<cluster-url>/?retryWrites=true&w=majority
MONGODB_DATABASE=work_status_bot
TELEGRAM_BOT_TOKEN=<telegram-bot-token>
TELEGRAM_GROUP_CHAT_ID=<telegram-group-chat-id>
TELEGRAM_WEBHOOK_SECRET=<random-secret>
CRON_SECRET=<random-secret>
```

Env variable source:

- `APP_ADDR`: local bind address for this HTTP service.
- `MONGODB_URI`: MongoDB Atlas connection string from Atlas cluster connect flow.
- `MONGODB_DATABASE`: database name to use in MongoDB Atlas.
- `TELEGRAM_BOT_TOKEN`: token from BotFather.
- `TELEGRAM_GROUP_CHAT_ID`: Telegram group chat ID for allowed commands and alerts.
- `TELEGRAM_WEBHOOK_SECRET`: random secret used when configuring Telegram webhook delivery.
- `CRON_SECRET`: random secret required in `X-Cron-Secret` for external monthly report trigger.

## Commands

```text
/help
/add_person Ivan Petrenko
/start_work Ivan Petrenko "API fix"
/status
/stop_work Ivan Petrenko Blocked by dependency
/report_month
/report_month 2026-06
```

## HTTP Routes

- `GET /health`
- `POST /telegram/webhook`
- `POST /cron/monthly-report` with `X-Cron-Secret`

## Run

```bash
go mod tidy
go test ./...
go run ./cmd/bot
```

## Validation Notes

- Configure Telegram webhook to `https://<host>/telegram/webhook`.
- Create MongoDB Atlas database and let startup ensure indexes for people, work records, incidents, monthly reports, and group config.
- Trigger monthly reports from external scheduler with `POST /cron/monthly-report`; do not run in-process cron.
- `GET /health` should respond under 1 second.
- Telegram command handlers and report generation should respond under 5 seconds for MVP-sized data.
