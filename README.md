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
- `telegram.command_menu_registration`
- `telegram.reply_keyboard_action`
- `telegram.flow_started`
- `telegram.flow_step_advanced`
- `telegram.flow_completed`
- `telegram.flow_cancelled`
- `telegram.flow_expired`
- `telegram.callback_received`
- `telegram.callback_result`
- `telegram.callback_answer`
- `telegram.user_language_changed`
- `telegram.command_delete`
- `telegram.message_delete`
- `telegram.datetime_parse`
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
change their own language from Settings using the reply keyboard menu. Supported
user-specific languages are Ukrainian, English, and Russian. The bot does not ask for
language on every interaction.

Startup registers the native Telegram command menu with:

- `/setup`
- `/groups`
- `/disable_group`
- `/add_person`
- `/start_work`
- `/status`
- `/stop_work`
- `/report_month`
- `/help`

If command menu registration fails, startup continues and the failure is logged without
secrets.

The `/help` command shows the persistent Telegram ReplyKeyboardMarkup menu. Group-wide
keyboard labels are Ukrainian:

- `Додати людину`
- `Почати роботу`
- `Статус`
- `Зупинити роботу`
- `Місячний звіт`
- `Налаштування`
- `Допомога`

Telegram sends reply keyboard presses as text messages. The bot handles those labels as
the same actions as slash commands. Slash commands remain available as fallback:

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

Button-driven flows:

- `Додати людину`: asks for `first_name last_name`, creates the person, then clears the flow.
- `Зупинити роботу`: asks for `first_name last_name [reason]`, stops active work, records the incident, sends the group alert, then clears the flow.
- `Почати роботу`: asks for person, title, and start time mode.

Start Work time modes:

- `Почати зараз`
- `Сьогодні`: then enter `HH:mm`
- `Вчора`: then enter `HH:mm`
- `Ввести дату вручну`: enter `DD.MM HH:mm`, `DD.MM.YYYY HH:mm`, or `YYYY-MM-DD HH:mm`
- `Скасувати`

Date/time input is parsed in Europe/Kyiv, stored as UTC, and future start values are
rejected.

After processing a user slash command, reply keyboard text, or manual flow input, the bot
attempts to delete only that user's message. This cleanup is best-effort: if the bot has
no permission to delete messages, the action still completes and the failure is logged.
Bot responses, status output, reports, alerts, incident messages, and monthly reports are
not deleted.

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
- Startup also ensures the `user_settings` index used for per-user language preferences and the `user_flow_states` indexes used for multi-step flows.
- Trigger monthly reports from external scheduler with `POST /cron/monthly-report`; do not run in-process cron.
- `GET /health` should respond under 1 second.
- Telegram command handlers and report generation should respond under 5 seconds for MVP-sized data.
