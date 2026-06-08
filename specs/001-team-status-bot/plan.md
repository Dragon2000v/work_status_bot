# Implementation Plan: Telegram Team Work Status Tracker

**Branch**: `001-team-status-bot` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-team-status-bot/spec.md`

## Summary

Build a Telegram-only team work status tracker as a small Go webhook service. The
service exposes health, Telegram webhook, and externally triggered monthly report HTTP
routes; stores durable people, work records, incidents, alert failures, and monthly reports in MongoDB
Atlas; and sends Telegram messages through direct Telegram Bot API HTTP requests. No web
UI, polling, queues, Redis, WebSockets, in-memory timers, or background infinite loops
are included.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: chi HTTP router, official MongoDB Go Driver, Go standard
library HTTP client for Telegram Bot API requests

**Storage**: MongoDB Atlas with UTC timestamps

**Testing**: Go tests (`go test ./...`)

**Target Platform**: Server-hosted Telegram webhook service suitable for free hosting
where the process may sleep between requests

**Project Type**: Go Telegram bot service

**Performance Goals**: Health responses under 1 second; Telegram command responses and
monthly report trigger responses under 5 seconds for MVP-scale data; all successful stop
actions attempt alert delivery during the request that changed the work record and return
a visible warning if alert delivery fails

**Constraints**: No Telegram polling; no in-memory timers; no background infinite loops;
no in-process cron loop; no Redis; no queues; no WebSockets; no MVP web UI; no separate
reset command; elapsed time calculated only from stored timestamps; database timestamps
stored in UTC; dates shown to users in Europe/Kyiv timezone; `POST /cron/monthly-report`
requires `X-Cron-Secret`; MVP commands require no user roles or authentication beyond
configured Telegram group/chat validation

**Scale/Scope**: One configured Telegram group for MVP; dozens of people; thousands of
work records and incidents per month; one stored monthly report per month

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries are explicit: `internal/telegram` handles Telegram and
  HTTP delivery, `internal/people`, `internal/works`, and `internal/reports` own business
  services, and `internal/database` owns MongoDB connection setup.
- Go code plan stays idiomatic, readable, and explicit: small packages, direct
  interfaces at persistence/Telegram boundaries, context-aware operations, no framework
  beyond chi routing.
- Telegram integration uses webhook only: `POST /telegram/webhook`; no polling.
- No background infinite loops and no in-memory timers: monthly reports are triggered by
  `POST /cron/monthly-report` with `X-Cron-Secret`, and elapsed time is calculated per
  request from stored timestamps.
- MongoDB Atlas uses the official MongoDB Go Driver.
- Database timestamps are UTC; user-facing dates display in Europe/Kyiv timezone.
- MVP scope has no web UI; only health, webhook, and cron-trigger routes exist.

**Initial Gate Result**: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-team-status-bot/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── http-api.md
│   └── telegram-commands.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
cmd/
└── bot/
    └── main.go

internal/
├── config/
│   └── config.go
├── database/
│   ├── mongo.go
│   └── indexes.go
├── telegram/
│   ├── client.go
│   ├── handler.go
│   ├── commands.go
│   └── messages.go
├── people/
│   ├── model.go
│   ├── repository.go
│   └── service.go
├── works/
│   ├── model.go
│   ├── repository.go
│   └── service.go
└── reports/
    ├── model.go
    ├── repository.go
    └── service.go
```

**Structure Decision**: Use the user-provided Go package layout. Keep HTTP routing in
`cmd/bot/main.go` and Telegram delivery in `internal/telegram`. Put domain behavior near
the relevant package (`people`, `works`, `reports`) and keep MongoDB connection/index
setup in `internal/database`. Avoid generic `service` or `repository` catch-all packages.

## Phase 0: Research Summary

See [research.md](./research.md).

Key decisions:
- Use chi for three explicit HTTP routes.
- Use direct Telegram Bot API HTTP calls.
- Use MongoDB Atlas with official MongoDB Go Driver.
- Use external scheduled HTTP requests for monthly reports.
- Use UTC persistence and Europe/Kyiv formatting.

## Phase 1: Design Summary

See [data-model.md](./data-model.md) and [contracts](./contracts/).

Primary data model:
- `Person`
- `WorkRecord`
- `Incident`
- `MonthlyReport`
- `TelegramGroupConfig`

External contracts:
- `GET /health`
- `POST /telegram/webhook`
- `POST /cron/monthly-report`
- Telegram commands: `/help`, `/add_person`, `/start_work`, `/status`, `/stop_work`,
  `/report_month`

## Post-Design Constitution Check

- Clean architecture boundaries remain explicit in package structure.
- Go + chi + official MongoDB driver are minimal and aligned with project constraints.
- Direct Telegram HTTP avoids extra bot frameworks while keeping webhook-only behavior.
- External scheduled HTTP trigger avoids in-process cron loops, timers, and background
  loops.
- No Redis, queues, WebSockets, or web UI are introduced.
- No separate reset command is introduced; interruption is handled by `/stop_work`.
- UTC/Kyiv and stored timestamp elapsed-time rules are represented in data model,
  contracts, and quickstart validation.

**Post-Design Gate Result**: PASS. No constitution violations.

## Complexity Tracking

No constitution violations.
