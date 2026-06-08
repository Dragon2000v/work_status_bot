# Implementation Plan: Structured Application Logging

**Branch**: `002-structured-logging` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-structured-logging/spec.md`

## Summary

Add structured application logging for the existing Telegram webhook bot. Use Go
standard library `log/slog`, add a small `internal/logger` package, add `APP_ENV` and
`LOG_LEVEL` configuration, and log startup, HTTP requests, MongoDB connection status,
Telegram webhook and command events, cron report events, and Telegram send outcomes
without logging secrets.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: Go standard library `log/slog`, chi HTTP router, official
MongoDB Go Driver, Go standard library HTTP client for Telegram Bot API requests

**Storage**: MongoDB Atlas with UTC timestamps; logs are emitted to process output and
are not persisted by the application

**Testing**: Go tests (`go test ./...`)

**Target Platform**: Server-hosted Telegram webhook service suitable for free hosting
where the process may sleep between requests

**Project Type**: Go Telegram bot service

**Performance Goals**: Logging must not add noticeable latency for MVP-sized requests;
health responses remain under 1 second, Telegram command handling and monthly report
triggers remain under 5 seconds

**Constraints**: No Telegram polling; no in-memory timers; no background infinite loops;
no in-process cron loop; no Redis; no queues; no WebSockets; no MVP web UI; no secret
logging; no full MongoDB URI logging; elapsed time still derives from stored timestamps;
database timestamps stored in UTC; user-facing dates shown in Europe/Kyiv timezone

**Scale/Scope**: One configured Telegram group for MVP; dozens of people; thousands of
work records and incidents per month; request-driven logs for startup, HTTP, Telegram,
MongoDB, cron, and alert send events

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries remain explicit: `internal/logger` owns logger setup and
  HTTP logging middleware; existing `internal/telegram`, `internal/database`,
  `internal/people`, `internal/works`, and `internal/reports` keep their responsibilities.
- Go code plan stays idiomatic and explicit: use standard library `log/slog`, small
  helper package, direct log calls at operation boundaries, no generic observability
  framework.
- Telegram integration remains webhook only: logging observes webhook requests and
  command handling; it does not add polling or background receivers.
- No in-memory timers or background infinite loops: duration is measured per request or
  operation only, and no scheduler is introduced.
- MongoDB Atlas remains the database of record; logging does not replace durable state.
- Database timestamps remain UTC; user-facing dates still display in Europe/Kyiv.
- MVP scope remains Telegram-only with health, webhook, and cron-trigger routes; no web UI.

**Initial Gate Result**: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-structured-logging/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── logging-events.md
│   └── environment.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
cmd/
└── bot/
    ├── main.go
    └── main_test.go

internal/
├── config/
│   └── config.go
├── logger/
│   ├── logger.go
│   ├── middleware.go
│   └── logger_test.go
├── database/
│   ├── mongo.go
│   └── indexes.go
├── telegram/
│   ├── client.go
│   ├── handler.go
│   └── *_test.go
└── reports/
    └── service.go

.env.example
README.md
```

**Structure Decision**: Add one small `internal/logger` package for `slog` setup,
log-level parsing, safe attributes, and chi-compatible HTTP middleware. Keep Telegram,
cron, MongoDB, and startup log calls at the existing operation boundaries so feature
behavior remains explicit and testable.

## Phase 0: Research Summary

See [research.md](./research.md).

Key decisions:
- Use standard library `log/slog`.
- Use JSON logs in production and text logs locally.
- Use `APP_ENV=local|production` and `LOG_LEVEL=debug|info|warn|error`.
- Redact secret-bearing values by policy: never log raw config, full MongoDB URI,
  Telegram API URLs, request bodies, authorization headers, or secret headers.
- Use response-wrapping HTTP middleware to capture status and duration.

## Phase 1: Design Summary

See [data-model.md](./data-model.md) and [contracts](./contracts/).

Primary logging model:
- `LogEvent`
- `LoggerConfig`
- `SafeContext`
- `SecretValue`

External contracts:
- Environment variables: `APP_ENV`, `LOG_LEVEL`
- Logging event fields for startup, HTTP requests, MongoDB, Telegram webhook/commands,
  cron monthly reports, and Telegram send outcomes

## Post-Design Constitution Check

- Clean boundaries remain simple: only logging setup/middleware goes into
  `internal/logger`; domain logic does not move.
- `log/slog` avoids new third-party dependencies and keeps Go code idiomatic.
- Logging is request-driven and startup-driven only; no polling, in-memory timers,
  background loops, queues, Redis, WebSockets, or UI are introduced.
- MongoDB Atlas persistence, UTC storage, Europe/Kyiv display, and timestamp-derived
  elapsed time remain unchanged.
- Secret-redaction rules are explicit in contracts and quickstart validation.

**Post-Design Gate Result**: PASS. No constitution violations.

## Complexity Tracking

No constitution violations.
