# Implementation Plan: Multi-Group Setup

**Branch**: `004-multi-group-setup` | **Date**: 2026-06-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-multi-group-setup/spec.md`

## Summary

Add durable multi-group registration to the existing Telegram webhook bot. Introduce a
`configured_groups` collection and `internal/groups` package, route Telegram commands
through a group authorization service, keep `TELEGRAM_GROUP_CHAT_ID` as a fallback
allowed group, deliver command-driven reports and stop alerts to the current chat, and
send cron monthly reports to every enabled configured group plus the fallback group when
not duplicated.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: chi HTTP router, official MongoDB Go Driver, Go standard
library HTTP client for Telegram Bot API requests, Go standard library `log/slog`

**Storage**: MongoDB Atlas with UTC timestamps; new `configured_groups` collection stores
allowed Telegram groups

**Testing**: Go tests (`go test ./...`)

**Target Platform**: Server-hosted Telegram webhook service suitable for free hosting
where the process may sleep between requests

**Project Type**: Go Telegram bot service

**Performance Goals**: Group authorization lookup should add no noticeable latency for
MVP-sized group usage; common Telegram command handling and cron monthly report trigger
handling should remain under 5 seconds

**Constraints**: Use Telegram webhook only; no Telegram polling; no in-memory timers; no
background infinite loops; no Redis; no queues; no WebSockets; no MVP web UI; durable
group setup in MongoDB; UTC database timestamps; Europe/Kyiv user-facing dates;
timestamp-derived elapsed time

**Scale/Scope**: One bot instance and one MongoDB database serving multiple Telegram
groups; dozens of configured groups for MVP; fallback `TELEGRAM_GROUP_CHAT_ID` remains
supported for local development and backward compatibility

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries remain explicit: `internal/groups` owns configured group
  model, repository, service, and authorization decisions; `internal/telegram` owns update
  handling; `internal/reports` owns report generation and delivery fan-out.
- Go code plan stays idiomatic and explicit: direct repository/service methods, simple
  command constants, and no generic authorization framework.
- Telegram integration remains webhook only; group setup is a command flow inside the
  existing webhook handler.
- No in-memory timers or background infinite loops are introduced; cron remains externally
  triggered.
- MongoDB Atlas remains the database of record; configured groups are durable.
- Database timestamps remain UTC; user-facing dates still display in Europe/Kyiv.
- MVP scope remains Telegram-only with no web UI, Redis, queues, or WebSockets.

**Initial Gate Result**: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/004-multi-group-setup/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── telegram-commands.md
│   ├── group-storage.md
│   ├── cron-report-delivery.md
│   └── logging-events.md
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
├── database/
│   ├── indexes.go
│   └── mongo.go
├── groups/
│   ├── model.go
│   ├── repository.go
│   ├── service.go
│   └── *_test.go
├── logger/
├── people/
├── reports/
│   ├── service.go
│   └── *_test.go
├── telegram/
│   ├── commands.go
│   ├── handler.go
│   ├── messages.go
│   └── *_test.go
├── users/
└── works/
```

**Structure Decision**: Add `internal/groups` for configured group persistence and
authorization. Extend `internal/telegram` with `/setup`, `/groups`, `/disable_group`,
chat title/user metadata extraction, and authorization checks before operational
commands. Extend `internal/reports` so cron delivery can send to all enabled configured
groups plus the fallback group, while command-triggered reports and alerts use the
current chat id.

## Phase 0: Research Summary

See [research.md](./research.md).

Key decisions:
- Store allowed groups in `configured_groups` with a unique `telegram_chat_id` index.
- Treat the env fallback group as allowed even when no stored group exists.
- Implement authorization as a small service decision, not middleware, because Telegram
  command exceptions (`/setup`, `/help`) depend on parsed command and chat type.
- Skip admin verification in MVP unless it is cheap to add; document the limitation.
- Send cron reports to every enabled configured group and include fallback chat once when
  present and not already in configured groups.
- Send command-driven alerts and reports to the current Telegram chat.

## Phase 1: Design Summary

See [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), and
[contracts](./contracts/).

Primary design model:
- `ConfiguredGroup`
- `GroupAuthorizationDecision`
- `GroupSetupRequest`
- `ReportDeliveryTarget`

External contracts:
- Telegram command behavior for `/setup`, `/groups`, and `/disable_group`
- MongoDB `configured_groups` shape and indexes
- Cron report delivery fan-out behavior
- Structured logging fields for setup, authorization, cron per-group delivery, and alert
  target chat ids

## Post-Design Constitution Check

- Boundaries remain simple: group authorization is isolated in `internal/groups`;
  Telegram handler coordinates commands and reports; report service handles reusable
  delivery targets.
- No new third-party dependencies are planned.
- Webhook-only Telegram behavior is preserved; no polling or background report scheduler
  is introduced.
- MongoDB Atlas stores configured groups with UTC timestamps.
- Command-driven elapsed time, reports, alerts, and Europe/Kyiv display rules remain
  timestamp-based and unchanged.
- No web UI, Redis, queues, WebSockets, in-memory timers, or background loops are
  introduced.

**Post-Design Gate Result**: PASS. No constitution violations.

## Complexity Tracking

No constitution violations.
