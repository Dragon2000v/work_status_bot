# Implementation Plan: Telegram Reply Keyboard, Languages, And Start Work Flow

**Branch**: `005-reply-keyboard-start-flow` | **Date**: 2026-06-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/005-reply-keyboard-start-flow/spec.md`

## Summary

Replace the current main-menu inline keyboard with a persistent Telegram reply keyboard,
register native Telegram commands at startup, preserve slash-command and callback-query
fallback behavior, persist per-user language and multi-step flow state in MongoDB, and
add a button-driven Start Work flow with Europe/Kyiv parsing and UTC storage.

The implementation will extend the existing Go codebase rather than introduce a new
architecture: `internal/telegram` remains the webhook delivery boundary,
`internal/users` owns user language settings, new `internal/flows` owns expiring
conversation state, new `internal/datetime` owns user date/time parsing, and existing
people/work/report/group services stay the business boundary.

## Technical Context

**Language/Version**: Go 1.22

**Primary Dependencies**: Existing Telegram webhook HTTP client, `github.com/go-chi/chi/v5`, official MongoDB Go Driver

**Storage**: MongoDB Atlas; existing collections plus `user_settings` and `user_flow_states`; all timestamps stored in UTC

**Testing**: Go tests with `go test ./...`

**Target Platform**: Server-hosted Telegram webhook service

**Project Type**: Go Telegram bot service

**Performance Goals**: Telegram webhook update handling should remain under 5 seconds for MVP-sized groups; `/health` remains under 1 second; no local scheduler or timer work is added

**Constraints**: Telegram webhook only; no polling; no web UI; no Redis; no queues; no WebSockets; no in-memory timers; no background infinite loops; no secrets in logs

**Scale/Scope**: MVP team/group bot with multiple authorized groups, per-user settings by Telegram user id, and one active flow state per user/chat; flow state expires 15 minutes after `updated_at`

## Constitution Check

*GATE: Passed before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries are explicit: Telegram delivery (`internal/telegram`), i18n/date/flow services (`internal/i18n`, `internal/datetime`, `internal/flows`), domain services (`internal/people`, `internal/works`, `internal/reports`, `internal/groups`, `internal/users`), and persistence repositories.
- Go code plan stays idiomatic, readable, and explicit without framework-heavy abstractions.
- Telegram integration remains webhook-only; no polling and no background infinite loops.
- No in-memory timers are introduced; flow expiration is checked from persisted `expires_at` during message handling, with `expires_at = updated_at + 15 minutes`.
- MongoDB Atlas remains the database of record through the official MongoDB Go Driver.
- Database timestamps are UTC; user-facing dates display in Europe/Kyiv timezone.
- MVP scope has no web UI.

## Project Structure

### Documentation (this feature)

```text
specs/005-reply-keyboard-start-flow/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── logging-events.md
│   ├── telegram-api.md
│   ├── telegram-interactions.md
│   └── storage.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/bot/
├── main.go                  # startup wiring, setWebhook, setMyCommands, HTTP routes
└── main_test.go

internal/
├── config/                  # existing environment/config loading
├── database/                # MongoDB connection and indexes
├── datetime/                # new Europe/Kyiv parser and validation errors
├── flows/                   # new user_flow_states model, repository, service
├── groups/                  # existing multi-group authorization/setup
├── i18n/                    # in-code uk/en/ru dictionaries and language helpers
├── logger/                  # existing structured logging helpers
├── people/                  # existing people domain
├── reports/                 # existing reports, incidents, alerts, cron outputs
├── telegram/                # Telegram API client, webhook handler, keyboards, command routing
├── users/                   # user_settings language persistence
└── works/                   # work records, UTC timestamps, Kyiv display helpers
```

**Structure Decision**: Keep the existing single Go service layout. Add only
`internal/flows` and `internal/datetime`; update existing `internal/telegram`,
`internal/i18n`, `internal/users`, `internal/database`, `internal/works`, and
`cmd/bot` where the current responsibilities already live.

## Phase 0: Research

Research decisions are recorded in [research.md](./research.md). All planning unknowns
are resolved; no `NEEDS CLARIFICATION` items remain.

## Phase 1: Design

Design artifacts:

- [data-model.md](./data-model.md)
- [contracts/telegram-api.md](./contracts/telegram-api.md)
- [contracts/telegram-interactions.md](./contracts/telegram-interactions.md)
- [contracts/storage.md](./contracts/storage.md)
- [contracts/logging-events.md](./contracts/logging-events.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

*GATE: Passed after Phase 1 design.*

- The design keeps Telegram delivery, services, and persistence separated.
- Startup command registration uses the existing Telegram client boundary and structured logs; `setMyCommands` failure is logged without secrets and does not stop application startup.
- Reply keyboard, callback fallback, deleteMessage, and sendMessage changes stay inside the Telegram delivery layer.
- User settings and flow states are persisted in MongoDB; no in-memory flow map is used.
- Flow expiration is timestamp-based (`updated_at + 15 minutes`) and evaluated on inbound webhook messages, not by timers.
- Date/time parsing stores UTC and displays Europe/Kyiv.
- No banned runtime technology or web UI is introduced.

## Complexity Tracking

No constitution violations.
