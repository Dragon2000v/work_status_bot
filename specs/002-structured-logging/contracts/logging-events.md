# Logging Events Contract

All log entries are structured and use stable field names. Production logs are expected
to be machine-parseable; local logs are expected to remain readable.

## Common Fields

- `time`: logger-provided timestamp
- `level`: `debug`, `info`, `warn`, or `error`
- `event`: stable event name
- `operation`: high-level operation group
- `outcome`: operation result when applicable
- `duration_ms`: elapsed handling time when applicable
- `error`: safe error category or safe error message when applicable

## HTTP Request Event

**Event**: `http.request`

**Fields**:
- `method`
- `path`
- `status`
- `duration_ms`
- `remote_addr`

**Secret rules**:
- Do not log request body.
- Do not log request headers containing secrets.
- Do not log query parameters that may contain secrets.

## Startup Events

**Events**:
- `app.start`
- `config.loaded`
- `mongodb.connect`
- `mongodb.indexes`
- `app.listen`

**Fields**:
- `app_env`
- `log_level`
- `database` when safe
- `addr` when safe
- `outcome`
- `error` when failed

**Secret rules**:
- Do not log `MONGODB_URI`.
- Do not log Telegram token, webhook secret, cron secret, or full environment config.

## Telegram Webhook And Command Events

**Events**:
- `telegram.webhook_received`
- `telegram.chat_rejected`
- `telegram.command`
- `telegram.command_result`

**Fields**:
- `chat_id`
- `user_id` when present
- `command` when parsed
- `outcome`
- `error` when failed

**Secret rules**:
- Do not log raw Telegram payload.
- Do not log full message body.
- Do not log bot token or Telegram API URL.

## Telegram Send Events

**Events**:
- `telegram.send`
- `telegram.alert_send`

**Fields**:
- `chat_id`
- `outcome`
- `error` when failed

**Secret rules**:
- Do not log token-bearing URLs.
- Do not log outgoing message text unless a future spec explicitly allows it.

## Cron Report Events

**Events**:
- `cron.monthly_report_triggered`
- `cron.monthly_report_result`
- `cron.monthly_report_send`

**Fields**:
- `month` when provided or resolved
- `outcome`: `generated`, `duplicate`, `skipped`, `unauthorized`, `invalid`, or `failed`
- `duration_ms` when applicable
- `error` when failed

**Secret rules**:
- Do not log `X-Cron-Secret`.
- Do not log request body beyond safe `month`.
