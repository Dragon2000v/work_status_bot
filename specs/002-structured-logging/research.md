# Research: Structured Application Logging

## Decision: Use Go Standard Library `log/slog`

**Rationale**: `log/slog` provides structured logging without another dependency. It
supports typed attributes, log levels, text output for local development, and JSON output
for production log collection.

**Alternatives considered**:
- Existing `log` package: simple, but not structured enough for consistent fields.
- Third-party logging libraries: useful, but unnecessary for MVP and add dependency
  surface.

## Decision: Add `internal/logger`

**Rationale**: A small package can centralize log-level parsing, environment-based
handler selection, safe common attributes, and HTTP middleware. This keeps logging setup
out of `cmd/bot/main.go` while avoiding broad abstractions.

**Alternatives considered**:
- Put all logging setup in `cmd/bot/main.go`: simpler initially, but makes tests and
  middleware harder to reuse.
- Add a full observability package: too broad for MVP.

## Decision: `APP_ENV` Controls Log Format

**Rationale**: `APP_ENV=local` should produce readable text logs for local debugging.
`APP_ENV=production` should produce JSON logs for hosting and production log pipelines.

**Alternatives considered**:
- Always JSON: production-friendly but less pleasant locally.
- Always text: readable but weaker for production filtering and aggregation.

## Decision: `LOG_LEVEL` Controls Minimum Level

**Rationale**: `debug|info|warn|error` is enough for local and production operation.
Local debugging can enable `debug`; production can default to `info`.

**Alternatives considered**:
- Boolean debug flag: less flexible than four standard levels.
- Numeric levels: less clear for operators.

## Decision: HTTP Logging Middleware Wraps Response Writer

**Rationale**: Middleware can capture method, path, remote address, status, and duration
for every route in one place. A small response writer wrapper is enough.

**Alternatives considered**:
- Route-level logging only: duplicates code and risks missing routes.
- Third-party middleware: unnecessary because chi and standard HTTP APIs are enough.

## Decision: Log Telegram Event Metadata, Not Raw Payloads

**Rationale**: Chat ID, user ID, command name, and result are enough to debug command
flow. Raw payloads and message bodies can expose personal or sensitive content.

**Alternatives considered**:
- Log complete Telegram updates: useful for debugging but violates least-exposure.
- Log only success/failure counts: too little context for production issues.

## Decision: Secret Redaction Is Avoidance-First

**Rationale**: The safest policy is not to pass secrets into log attributes. Code should
log safe derived values such as database name, environment, command name, status code,
month, and error category. It must not log full MongoDB URI, bot token, webhook secret,
cron secret, secret headers, or token-bearing Telegram URLs.

**Alternatives considered**:
- Generic string sanitizer: useful as a backstop, but brittle for every possible secret
  shape.
- Manual log review only: not reliable enough.
