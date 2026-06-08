# Tasks: Structured Application Logging

**Input**: Design documents from `/specs/002-structured-logging/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Focused Go test tasks are included because the feature specification and
quickstart require validation of structured log fields and secret redaction behavior.

**Organization**: Tasks are grouped by user story so each story can be implemented and
validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency
  on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add logging configuration surface and documentation prerequisites.

- [X] T001 Add APP_ENV and LOG_LEVEL fields with allowed values local|production and debug|info|warn|error to internal/config/config.go
- [X] T002 Add APP_ENV=local and LOG_LEVEL=info to .env.example
- [X] T003 Add APP_ENV=local and LOG_LEVEL=debug placeholder values to .env
- [X] T004 Update README.md environment section with APP_ENV and LOG_LEVEL source, allowed values, and secret logging warning

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared slog setup, safe logging helpers, and HTTP request middleware needed by all stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T005 [P] Add logger config tests for APP_ENV format selection and LOG_LEVEL parsing in internal/logger/logger_test.go
- [X] T006 [P] Add HTTP middleware tests for method, path, status, duration_ms, and remote_addr fields in internal/logger/middleware_test.go
- [X] T007 [P] Add secret redaction tests for representative MongoDB URI, Telegram token, webhook secret, and cron secret strings in internal/logger/logger_test.go
- [X] T008 Create internal/logger package with slog setup, level parsing, APP_ENV handler selection, and safe attribute helpers in internal/logger/logger.go
- [X] T009 Create chi-compatible HTTP logging middleware with response status capture and duration_ms in internal/logger/middleware.go
- [X] T010 Wire logger initialization into cmd/bot/main.go using APP_ENV and LOG_LEVEL without logging full config or secrets
- [X] T011 Apply HTTP logging middleware to all routes in cmd/bot/main.go

**Checkpoint**: Structured logger and request middleware exist, tests pass, and user story logging can be added safely.

---

## Phase 3: User Story 1 - Diagnose Request Flow (Priority: P1) MVP

**Goal**: Operators can inspect request logs for health, Telegram webhook, and cron routes with method, path, status, duration, and remote address.

**Independent Test**: Send health, webhook, and cron requests in tests and confirm structured request logs exist without secrets.

### Tests for User Story 1

- [X] T012 [P] [US1] Add health route request logging test asserting event=http.request and status=200 in cmd/bot/main_test.go
- [X] T013 [P] [US1] Add webhook request logging test asserting event=http.request and path=/telegram/webhook in internal/telegram/handler_test.go
- [X] T014 [P] [US1] Add cron request logging test asserting unauthorized and successful cron request status fields in cmd/bot/main_test.go

### Implementation for User Story 1

- [X] T015 [US1] Ensure HTTP middleware logs method, path, status, duration_ms, and remote_addr for all routes in internal/logger/middleware.go
- [X] T016 [US1] Ensure cmd/bot/main.go installs HTTP logging middleware before GET /health, POST /telegram/webhook, and POST /cron/monthly-report route registration
- [X] T017 [US1] Ensure request logging never includes request bodies, X-Cron-Secret, or query strings in internal/logger/middleware.go

**Checkpoint**: Request flow logging is independently functional and safe.

---

## Phase 4: User Story 2 - Trace Bot Operations (Priority: P2)

**Goal**: Operators can follow Telegram webhook receipt, chat validation, command handling, and Telegram send outcomes from logs.

**Independent Test**: Send supported and malformed webhook updates, then confirm logs include chat_id, user_id when present, command name, success/failure, and send result without tokens or raw payloads.

### Tests for User Story 2

- [X] T018 [P] [US2] Add Telegram webhook event logging tests for webhook received, configured chat accepted, rejected chat, and malformed JSON in internal/telegram/handler_logging_test.go
- [X] T019 [P] [US2] Add Telegram command logging tests for command name and success/failure outcomes in internal/telegram/handler_logging_test.go
- [X] T020 [P] [US2] Add Telegram client logging tests for send success and failure without token-bearing URL exposure in internal/telegram/client_test.go
- [X] T021 [P] [US2] Add stop alert logging tests for alert success and alert failure warning flow in internal/telegram/handler_stop_test.go

### Implementation for User Story 2

- [X] T022 [US2] Extend Telegram Update model with user id field needed for safe logging in internal/telegram/handler.go
- [X] T023 [US2] Inject *slog.Logger or logger-compatible dependency into Telegram handler constructor in internal/telegram/handler.go
- [X] T024 [US2] Log telegram.webhook_received, telegram.chat_rejected, telegram.command, and telegram.command_result with chat_id, user_id, command, and outcome in internal/telegram/handler.go
- [X] T025 [US2] Inject *slog.Logger or logger-compatible dependency into Telegram client constructor in internal/telegram/client.go
- [X] T026 [US2] Log telegram.send and telegram.alert_send success/failure without token-bearing URLs or outgoing message text in internal/telegram/client.go and internal/telegram/handler.go
- [X] T027 [US2] Update Telegram handler and client tests to use captured slog output helpers in internal/telegram/handler_logging_test.go

**Checkpoint**: Telegram event and command logs are independently functional and safe.

---

## Phase 5: User Story 3 - Confirm Startup And Data Connectivity (Priority: P3)

**Goal**: Operators can confirm startup, configuration, MongoDB connection, index setup, route serving, and cron report outcomes from logs.

**Independent Test**: Start application components in tests or invoke route setup and cron trigger tests, then confirm startup/MongoDB/cron logs show safe outcomes without URI passwords or secrets.

### Tests for User Story 3

- [X] T028 [P] [US3] Add config tests for APP_ENV and LOG_LEVEL default and invalid value behavior in internal/config/config_test.go
- [X] T029 [P] [US3] Add startup logging test for config loaded and app listen events without secret values in cmd/bot/main_test.go
- [X] T030 [P] [US3] Add MongoDB connection status logging tests for success/failure log fields without full URI in internal/database/mongo_test.go
- [X] T031 [P] [US3] Add cron logging tests for triggered, generated, duplicate, unauthorized, invalid, send success, and send failure outcomes in cmd/bot/main_test.go

### Implementation for User Story 3

- [X] T032 [US3] Log app.start, config.loaded, mongodb.connect, mongodb.indexes, and app.listen events in cmd/bot/main.go
- [X] T033 [US3] Add safe MongoDB connection status logging around Connect and EnsureIndexes calls without logging MONGODB_URI in cmd/bot/main.go
- [X] T034 [US3] Log cron.monthly_report_triggered, cron.monthly_report_result, and cron.monthly_report_send outcomes in cmd/bot/main.go
- [X] T035 [US3] Update config loader to parse APP_ENV and LOG_LEVEL explicitly while preserving existing required env variable names in internal/config/config.go
- [X] T036 [US3] Ensure startup and cron logs never include MONGODB_URI, TELEGRAM_BOT_TOKEN, TELEGRAM_WEBHOOK_SECRET, CRON_SECRET, or X-Cron-Secret in cmd/bot/main.go

**Checkpoint**: Startup, MongoDB, and cron logs are independently functional and safe.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate full logging behavior, documentation, and project constraints.

- [X] T037 [P] Update quickstart validation notes for APP_ENV, LOG_LEVEL, local text logs, production JSON logs, and secret checks in README.md
- [X] T038 [P] Add logging examples and field contract summary to README.md
- [X] T039 Run gofmt on all Go files in cmd/ and internal/
- [X] T040 Run go test ./... and fix any failing tests in cmd/ and internal/
- [X] T041 Verify captured logs and source code do not include secret logging of MONGODB_URI, MongoDB password, TELEGRAM_BOT_TOKEN, TELEGRAM_WEBHOOK_SECRET, CRON_SECRET, X-Cron-Secret, request bodies, or raw Telegram payloads in cmd/, internal/, README.md, and .env.example
- [X] T042 Validate specs/002-structured-logging/quickstart.md scenarios against implemented logs, including local readable output, production JSON output, request logs, Telegram command logs, cron logs, and secret absence checks
- [X] T043 Verify no implementation adds Redis, queues, WebSockets, Telegram polling, in-memory timers, background infinite loops, or web UI files in cmd/, internal/, go.mod, and README.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational. This is MVP for request logging.
- **User Story 2 (Phase 4)**: Depends on Foundational and can start after request logging middleware exists.
- **User Story 3 (Phase 5)**: Depends on Foundational and can start after logger initialization exists.
- **Polish (Phase 6)**: Depends on selected user stories.

### User Story Dependencies

- **US1**: Independent after Foundational.
- **US2**: Independent after Foundational, but uses the shared logger from Phase 2.
- **US3**: Independent after Foundational, but uses the shared logger from Phase 2.

### Within Each User Story

- Tests before implementation.
- Logger setup before middleware and operation logging.
- Middleware before route-specific request log validation.
- Operation logging before documentation validation.

---

## Parallel Opportunities

- Setup documentation/env tasks T002-T004 can run after T001.
- Foundational tests T005-T007 can run in parallel.
- US1 tests T012-T014 can run in parallel.
- US2 tests T018-T021 can run in parallel.
- US3 tests T028-T031 can run in parallel.
- Polish documentation tasks T037-T038 can run in parallel.

## Parallel Example: User Story 1

```bash
Task: "T012 [P] [US1] Add health route request logging test asserting event=http.request and status=200 in cmd/bot/main_test.go"
Task: "T013 [P] [US1] Add webhook request logging test asserting event=http.request and path=/telegram/webhook in internal/telegram/handler_test.go"
Task: "T014 [P] [US1] Add cron request logging test asserting unauthorized and successful cron request status fields in cmd/bot/main_test.go"
```

## Parallel Example: User Story 2

```bash
Task: "T018 [P] [US2] Add Telegram webhook event logging tests for webhook received, configured chat accepted, rejected chat, and malformed JSON in internal/telegram/handler_logging_test.go"
Task: "T020 [P] [US2] Add Telegram client logging tests for send success and failure without token-bearing URL exposure in internal/telegram/client_test.go"
Task: "T021 [P] [US2] Add stop alert logging tests for alert success and alert failure warning flow in internal/telegram/handler_stop_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T028 [P] [US3] Add config tests for APP_ENV and LOG_LEVEL default and invalid value behavior in internal/config/config_test.go"
Task: "T030 [P] [US3] Add MongoDB connection status logging tests for success/failure log fields without full URI in internal/database/mongo_test.go"
Task: "T031 [P] [US3] Add cron logging tests for triggered, generated, duplicate, unauthorized, invalid, send success, and send failure outcomes in cmd/bot/main_test.go"
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete User Story 1 request logging.
3. Validate health, webhook, and cron request logs with no secret exposure.
4. Stop and confirm request flow diagnostics work before adding deeper Telegram/startup logs.

### Incremental Delivery

1. Add User Story 2 for Telegram webhook, command, and send-result logs.
2. Add User Story 3 for startup, MongoDB, and cron report logs.
3. Run full quickstart validation and `go test ./...`.

### Notes

- Do not log full request bodies, raw Telegram payloads, outgoing Telegram message text,
  full MongoDB URI, bot token, webhook secret, cron secret, or `X-Cron-Secret`.
- Do not add logging backends, queues, Redis, polling, timers, background loops, or web UI.
- Keep logging explicit and small; prefer `log/slog` standard library only.
