# Data Model: Structured Application Logging

## LogEvent

Represents one structured observation emitted by the application.

**Fields**:
- `time`: event timestamp assigned by the logger
- `level`: `debug`, `info`, `warn`, or `error`
- `event`: stable event name, such as `app.start`, `http.request`,
  `telegram.webhook_received`, or `cron.monthly_report`
- `operation`: high-level area, such as `startup`, `http`, `mongodb`, `telegram`,
  `cron`, or `telegram_send`
- `outcome`: success, failure, skipped, duplicate, unauthorized, invalid, or rejected
- `duration_ms`: elapsed time for request or operation events when applicable
- `error`: safe error category or message when applicable; must not contain secrets

**Validation**:
- Every log event must include `event` and `outcome` where an outcome exists.
- Request and operation duration fields must be measured from current request handling,
  not stored as business elapsed time.
- Log events must not include raw request bodies, raw Telegram update payloads, raw
  environment config, or secret values.

## LoggerConfig

Represents logging runtime configuration.

**Fields**:
- `app_env`: `local` or `production`
- `log_level`: `debug`, `info`, `warn`, or `error`

**Validation**:
- Unknown `app_env` values are rejected or defaulted explicitly.
- Unknown `log_level` values are rejected or defaulted explicitly.
- Configuration logs may include `app_env` and `log_level`; they must not include
  secrets or full connection strings.

## SafeContext

Represents non-secret context attached to logs.

**Fields**:
- `method`: HTTP method
- `path`: HTTP route path
- `status`: HTTP status code
- `remote_addr`: request remote address
- `chat_id`: Telegram chat ID
- `user_id`: Telegram user ID if present
- `command`: Telegram command name
- `month`: report month key
- `database`: MongoDB database name

**Validation**:
- Safe context must be enough to debug flow without exposing message bodies or secrets.
- `chat_id` and `user_id` are allowed operational identifiers.
- Full MongoDB URI and Telegram API URL are not safe context.

## SecretValue

Represents values that must never be logged.

**Examples**:
- Telegram bot token
- MongoDB password
- Full MongoDB URI
- Telegram webhook secret
- Cron secret
- `X-Cron-Secret` header
- Secret-bearing Telegram Bot API URL
- Raw environment variable dump

**Validation**:
- Secret values must not appear in log attributes, error messages, request logs, or
  startup logs.
- Tests should include representative secret strings and assert they are absent from
  captured logs.

## State Transitions

Logging has no durable business state transitions. Runtime events move through a simple
observation lifecycle:

```text
operation begins -> log event emitted -> operation outcome known -> outcome log emitted
```

No log event may mutate people, work records, incidents, or reports.
