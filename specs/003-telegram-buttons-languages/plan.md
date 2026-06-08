# Implementation Plan: Telegram Buttons And User Languages

**Branch**: `003-telegram-buttons-languages` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-telegram-buttons-languages/spec.md`

## Summary

Add Telegram inline button menus, callback query handling, and durable per-user language
preferences to the existing Go Telegram webhook bot. Keep text commands as fallback,
default all output to Ukrainian, localize user-specific interactions through a small
`internal/i18n` package, persist user settings in MongoDB Atlas, and add best-effort
command message deletion with structured logs.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: chi HTTP router, official MongoDB Go Driver, Go standard
library HTTP client for Telegram Bot API requests, Go standard library `log/slog`

**Storage**: MongoDB Atlas with UTC timestamps; new `user_settings` collection stores
Telegram user language preferences

**Testing**: Go tests (`go test ./...`)

**Target Platform**: Server-hosted Telegram webhook service suitable for free hosting
where the process may sleep between requests

**Project Type**: Go Telegram bot service

**Performance Goals**: Callback acknowledgements should be attempted during every
callback request; menu rendering and language lookup should add no noticeable latency for
MVP-sized group usage; normal webhook handling should remain under 5 seconds for common
actions

**Constraints**: Use Telegram webhook only; no Telegram polling; no in-memory timers; no
background infinite loops; no Redis; no queues; no WebSockets; no MVP web UI; Ukrainian
is the default language; public group reports and alerts remain Ukrainian; elapsed time
still derives from stored timestamps; database timestamps stored in UTC; user-facing dates
shown in Europe/Kyiv timezone

**Scale/Scope**: One configured Telegram group for MVP; dozens of users and people;
thousands of work records, incidents, and reports per month; three supported user
languages (`uk`, `en`, `ru`); group-level language settings explicitly out of scope

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries remain explicit: `internal/telegram` owns Telegram update
  handling and Bot API calls, `internal/i18n` owns message lookup, and persistence owns
  `user_settings`.
- Go code plan stays idiomatic and explicit: in-code dictionaries, simple callback action
  constants, direct repository/service methods, and no localization framework.
- Telegram integration remains webhook only: callback queries are handled as webhook
  update variants, not polling.
- No in-memory timers or background infinite loops are introduced; command cleanup and
  callback acknowledgement are request-scoped operations.
- MongoDB Atlas remains the database of record; user language preferences are durable.
- Database timestamps remain UTC; user-facing dates still display in Europe/Kyiv.
- MVP scope remains Telegram-only with no web UI, Redis, queues, or WebSockets.

**Initial Gate Result**: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/003-telegram-buttons-languages/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── telegram-interactions.md
│   ├── user-settings.md
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
│   └── config.go
├── database/
│   ├── indexes.go
│   └── mongo.go
├── i18n/
│   ├── i18n.go
│   └── i18n_test.go
├── logger/
│   ├── logger.go
│   └── middleware.go
├── people/
├── reports/
├── telegram/
│   ├── callbacks.go
│   ├── client.go
│   ├── commands.go
│   ├── handler.go
│   ├── keyboard.go
│   ├── messages.go
│   └── *_test.go
├── users/
│   ├── model.go
│   ├── repository.go
│   ├── service.go
│   └── *_test.go
└── works/
```

**Structure Decision**: Add a focused `internal/i18n` package for translation keys and
simple dictionaries, plus `internal/users` for durable language settings. Extend the
existing `internal/telegram` package with callback routing, inline keyboard rendering,
localized message formatting, and Telegram Bot API methods. Keep current `people`,
`works`, `reports`, `database`, `logger`, and `cmd/bot` responsibilities intact.

## Phase 0: Research Summary

See [research.md](./research.md).

Key decisions:
- Represent localized text with small in-code dictionaries keyed by stable message IDs.
- Store per-user settings in MongoDB collection `user_settings` with a unique
  `telegram_user_id` index.
- Extend the Telegram update model to include `callback_query` while preserving message
  command handling.
- Always call callback acknowledgement for every callback query and log both success and
  failure.
- Add Bot API methods for `answerCallbackQuery`, `deleteMessage`, `sendMessage` with
  inline keyboard, and optionally `editMessageText`.
- Keep public group reports and alerts Ukrainian for this feature.

## Phase 1: Design Summary

See [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), and
[contracts](./contracts/).

Primary design model:
- `UserSetting`
- `Language`
- `TranslationKey`
- `InlineMenu`
- `CallbackAction`
- `CommandDeletionAttempt`

External contracts:
- Telegram webhook update handling for message and callback query updates
- Telegram Bot API payloads used by this feature
- MongoDB `user_settings` collection shape and indexes
- Structured log events for callbacks, language changes, and command deletion

## Post-Design Constitution Check

- Clean boundaries remain simple: i18n lookup has no Telegram or MongoDB dependency;
  user settings persistence is isolated; Telegram handling coordinates callbacks,
  messages, and Bot API methods.
- No new third-party dependencies are planned; Go standard library maps and explicit
  structs are enough for translation and Telegram payloads.
- Callback handling is webhook-driven and request-scoped; no polling or background loop is
  introduced.
- MongoDB Atlas persists language settings with UTC timestamps.
- Reports, alerts, existing elapsed-time rules, and Europe/Kyiv display behavior remain
  unchanged except for Ukrainian output being explicit for public group output.
- No web UI, Redis, queues, WebSockets, in-memory timers, or internal schedulers are
  introduced.

**Post-Design Gate Result**: PASS. No constitution violations.

## Complexity Tracking

No constitution violations.
