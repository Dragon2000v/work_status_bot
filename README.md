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
APP_ENV=local
LOG_LEVEL=info
MONGODB_URI=mongodb+srv://<username>:<password>@<cluster-url>/?retryWrites=true&w=majority
MONGODB_DATABASE=work_status_bot
TELEGRAM_BOT_TOKEN=<telegram-bot-token>
TELEGRAM_GROUP_CHAT_ID=<telegram-group-chat-id>
TELEGRAM_WEBHOOK_SECRET=<random-secret>
TELEGRAM_WEBHOOK_URL=https://<public-host>/telegram/webhook
TELEGRAM_AUTO_SET_WEBHOOK=false
CRON_SECRET=<random-secret>
```

Env variable source:

- `APP_ADDR`: local bind address for this HTTP service.
- `APP_ENV`: logging output mode; use `local` for readable text logs and `production` for JSON logs.
- `LOG_LEVEL`: minimum log level; allowed values are `debug`, `info`, `warn`, and `error`.
- `MONGODB_URI`: MongoDB Atlas connection string from Atlas cluster connect flow.
- `MONGODB_DATABASE`: database name to use in MongoDB Atlas.
- `TELEGRAM_BOT_TOKEN`: token from BotFather.
- `TELEGRAM_GROUP_CHAT_ID`: optional fallback Telegram group chat ID for local development and backward compatibility.
- `TELEGRAM_WEBHOOK_SECRET`: random secret used when configuring Telegram webhook delivery.
- `TELEGRAM_WEBHOOK_URL`: public HTTPS Telegram webhook URL, usually `https://<host>/telegram/webhook`.
- `TELEGRAM_AUTO_SET_WEBHOOK`: `true` to register the webhook on startup, `false` to skip registration.
- `CRON_SECRET`: random secret required in `X-Cron-Secret` for external monthly report trigger.

Never log real secret values. Logs must not include the Telegram bot token, MongoDB password, full MongoDB URI, webhook secret, cron secret, or `X-Cron-Secret`.

When `TELEGRAM_AUTO_SET_WEBHOOK=true`, startup calls Telegram `setWebhook` using `TELEGRAM_WEBHOOK_URL` and sends `TELEGRAM_WEBHOOK_SECRET` as the Telegram `secret_token`. Registration logs show attempt, success, failure, or skipped status without logging the bot token or webhook secret.

## Logging

The bot writes structured logs to process output. Local mode (`APP_ENV=local`) uses readable structured text logs. Production mode (`APP_ENV=production`) uses JSON logs.

Important fields include `event`, `operation`, `outcome`, `method`, `path`, `status`, `duration_ms`, `remote_addr`, `chat_id`, `user_id`, `command`, and `month`.

Example events:

- `app.start`
- `config.loaded`
- `mongodb.connect`
- `mongodb.indexes`
- `app.listen`
- `http.request`
- `telegram.webhook_received`
- `telegram.command`
- `telegram.command_result`
- `telegram.callback_received`
- `telegram.callback_result`
- `telegram.callback_answer`
- `telegram.user_language_changed`
- `telegram.command_delete`
- `telegram.group_setup`
- `telegram.group_authorization`
- `telegram.group_list`
- `telegram.group_disable`
- `telegram.send`
- `telegram.alert_send`
- `cron.monthly_report_triggered`
- `cron.monthly_report_result`
- `cron.monthly_report_send`

## Commands

Ukrainian is the default language for bot messages and public group output. Users can
change their own language from Settings using the inline button menu. Supported
user-specific languages are Ukrainian, English, and Russian. The bot does not ask for
language on every interaction.

The `/help` command shows the inline button menu. Menu buttons use Telegram
InlineKeyboardMarkup and support the same core actions as text commands:

- Add person
- Start work
- Status
- Stop work
- Monthly report
- Settings
- Help

```text
/help
/setup
/groups
/disable_group
/add_person Ivan Petrenko
/start_work Ivan Petrenko "API fix"
/status
/stop_work Ivan Petrenko Blocked by dependency
/report_month
/report_month 2026-06
```

Text commands remain available as fallback. After processing a user command, the bot
attempts to delete only that user's command message. This cleanup is best-effort: if the
bot has no permission to delete messages, the command still completes and the failure is
logged. Bot responses, reports, alerts, incident messages, and monthly reports are not
deleted by command cleanup.

## Multi-Group Setup

The bot can serve multiple Telegram groups with one deployment and one MongoDB database.
A group registers itself by running:

```text
/setup
```

`/setup` works only in Telegram group or supergroup chats. It stores the current chat id,
group title, setup user id, optional setup username, enabled status, and UTC timestamps in
the `configured_groups` collection. Re-running `/setup` is idempotent: it updates the
stored title/setup metadata and re-enables a disabled group.

Unknown groups may use only `/setup` and `/help`. Operational commands are rejected until
the group is configured. `TELEGRAM_GROUP_CHAT_ID` remains an allowed fallback group when
configured, but adding new groups no longer requires changing env vars or redeploying.

`/groups` can be run only from an enabled configured group or the fallback group. Its
output lists all stored configured groups, including disabled groups, and shows each
group's enabled/disabled status.

`/disable_group` disables the current stored group without deleting its history. A
disabled group cannot run `/groups` or operational commands unless it is re-enabled with
`/setup` or it matches the fallback group id.

MVP limitation: `/setup` and `/disable_group` do not verify Telegram admin status.

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
- Create MongoDB Atlas database and let startup ensure indexes for people, work records, incidents, monthly reports, user settings, configured groups, and legacy group config.
- Startup also ensures the `user_settings` index used for per-user language preferences.
- Trigger monthly reports from external scheduler with `POST /cron/monthly-report`; do not run in-process cron.
- `GET /health` should respond under 1 second.
- Telegram command handlers and report generation should respond under 5 seconds for MVP-sized data.
