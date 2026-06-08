# Tasks: Telegram Team Work Status Tracker

**Input**: Design documents from `/specs/001-team-status-bot/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Focused Go test tasks are included because the specification defines
independent tests for each user story and quickstart validation requires `go test ./...`.

**Organization**: Tasks are grouped by user story so each story can be implemented and
validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency
  on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go service skeleton and shared runtime configuration.

- [ ] T001 Create Go module definition in go.mod
- [ ] T002 Create application entrypoint with chi router skeleton in cmd/bot/main.go
- [ ] T003 [P] Create environment config loader for APP_ADDR, MONGODB_URI, MONGODB_DATABASE, TELEGRAM_BOT_TOKEN, TELEGRAM_GROUP_CHAT_ID, TELEGRAM_WEBHOOK_SECRET, and CRON_SECRET in internal/config/config.go
- [ ] T004 [P] Create MongoDB client setup and shutdown helpers using the official MongoDB Go Driver in internal/database/mongo.go
- [ ] T005 [P] Create MongoDB index setup function for people, work_records, incidents, monthly_reports, and telegram_group_config in internal/database/indexes.go
- [ ] T006 [P] Create Telegram Bot API HTTP client skeleton using direct HTTP requests in internal/telegram/client.go
- [ ] T007 Create GET /health route returning {"status":"ok"} in cmd/bot/main.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared domain models, time handling, HTTP security, and routing needed by all stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T008 [P] Create Person model with normalized full name fields in internal/people/model.go
- [ ] T009 [P] Create WorkRecord model with active/stopped statuses and UTC timestamp fields in internal/works/model.go
- [ ] T010 [P] Create Incident and MonthlyReport models in internal/reports/model.go
- [ ] T011 [P] Create Europe/Kyiv date formatting and elapsed time helpers from stored timestamps in internal/works/time.go
- [ ] T012 Create Telegram command parsing types for /help, /add_person, /start_work, /status, /stop_work, and /report_month in internal/telegram/commands.go
- [ ] T013 Create Telegram message formatting helpers for confirmations, errors, warnings, status, alerts, and reports in internal/telegram/messages.go
- [ ] T014 Create POST /telegram/webhook route that rejects unconfigured chats and dispatches parsed commands without polling in internal/telegram/handler.go
- [ ] T015 Create POST /cron/monthly-report route that requires X-Cron-Secret matching CRON_SECRET in cmd/bot/main.go
- [ ] T016 Document banned technologies and runtime constraints in README.md

**Checkpoint**: Foundation ready. All routes exist, shared models exist, and user story work can proceed.

---

## Phase 3: User Story 1 - Track Work In Group Chat (Priority: P1) MVP

**Goal**: A team member can add a person, start work, view status, and stop work from the Telegram group.

**Independent Test**: In a Telegram group, add a person, start work, check status, then stop work and verify the status changes.

### Tests for User Story 1

- [ ] T017 [P] [US1] Add person service tests for valid person and duplicate normalized name in internal/people/service_test.go
- [ ] T018 [P] [US1] Add work service tests for start, status elapsed time from stored timestamps, stop, and no active work errors in internal/works/service_test.go
- [ ] T019 [P] [US1] Add Telegram command handler tests for /help, /add_person, /start_work, /status, and /stop_work happy paths in internal/telegram/handler_test.go
- [ ] T020 [P] [US1] Add Telegram command handler error tests for malformed command format, missing arguments, unknown person, no active work record, already stopped work record, and any second active work for the same person regardless of title in internal/telegram/handler_errors_test.go
- [ ] T021 [P] [US1] Add Telegram command handler test verifying MVP commands require no user role or authentication beyond configured chat validation in internal/telegram/handler_auth_test.go

### Implementation for User Story 1

- [ ] T022 [P] [US1] Implement PeopleRepository with create and find-by-normalized-name operations in internal/people/repository.go
- [ ] T023 [P] [US1] Implement WorkRepository with create active, find active by person, stop active, and list status operations in internal/works/repository.go
- [ ] T024 [US1] Implement PeopleService add person and duplicate validation in internal/people/service.go
- [ ] T025 [US1] Implement WorkService start, status, and stop operations using stored UTC timestamps in internal/works/service.go
- [ ] T026 [US1] Wire /help, /add_person, /start_work, /status, and /stop_work commands to services in internal/telegram/handler.go
- [ ] T027 [US1] Format all user-facing dates in Europe/Kyiv timezone in internal/telegram/messages.go

**Checkpoint**: User Story 1 is independently functional and covers MVP daily tracking.

---

## Phase 4: User Story 2 - Alert On Stop (Priority: P2)

**Goal**: Stopping work creates an incident, attempts Telegram alert delivery, and records alert failure with a visible command warning.

**Independent Test**: Stop an active work record, verify group alert attempt. Simulate alert delivery failure and verify warning response plus alert failure event.

### Tests for User Story 2

- [ ] T028 [P] [US2] Add report incident service tests for stopped and alert_failed events in internal/reports/service_test.go
- [ ] T029 [P] [US2] Add Telegram client tests for send message success and failure handling in internal/telegram/client_test.go
- [ ] T030 [P] [US2] Add stop command handler tests for alert success and alert failure warning in internal/telegram/handler_stop_test.go

### Implementation for User Story 2

- [ ] T031 [P] [US2] Implement IncidentRepository create and list-by-month operations in internal/reports/repository.go
- [ ] T032 [US2] Implement ReportsService incident creation for stopped and alert_failed events in internal/reports/service.go
- [ ] T033 [US2] Implement TelegramClient SendMessage through direct Telegram Bot API HTTP request in internal/telegram/client.go
- [ ] T034 [US2] Update WorkService stop flow to save stopped incident before alert attempt in internal/works/service.go
- [ ] T035 [US2] Update /stop_work handler to attempt alert, record alert_failed on send failure, and return visible warning in internal/telegram/handler.go

**Checkpoint**: Stop alerts and alert failure semantics are independently testable.

---

## Phase 5: User Story 3 - Generate Monthly Report (Priority: P3)

**Goal**: Generate one report per month from Telegram or external HTTP trigger, including all records active at any moment during the selected month.

**Independent Test**: Create records that start before, during, and stop during a month; generate report manually and by cron endpoint; verify overlap inclusion and duplicate prevention.

### Tests for User Story 3

- [ ] T036 [P] [US3] Add report service tests for month overlap inclusion across active, started-in-month, stopped-in-month, and cross-month records in internal/reports/service_report_test.go
- [ ] T037 [P] [US3] Add duplicate monthly report repository tests for unique month behavior in internal/reports/repository_test.go
- [ ] T038 [P] [US3] Add /report_month command handler tests for default current month, YYYY-MM argument, invalid month format, and duplicate notice in internal/telegram/handler_report_test.go
- [ ] T039 [P] [US3] Add POST /cron/monthly-report HTTP tests for X-Cron-Secret success, missing header, invalid header, generated response, duplicate response, and invalid month format in cmd/bot/main_test.go

### Implementation for User Story 3

- [ ] T040 [P] [US3] Implement MonthlyReport repository create-if-absent and find-by-month operations in internal/reports/repository.go
- [ ] T041 [P] [US3] Implement WorkRepository list records overlapping selected month query in internal/works/repository.go
- [ ] T042 [US3] Implement ReportsService monthly report generation with people, overlapping work records, elapsed time, incidents, and duplicate prevention in internal/reports/service.go
- [ ] T043 [US3] Implement /report_month command parsing and report response formatting in internal/telegram/handler.go
- [ ] T044 [US3] Implement POST /cron/monthly-report request parsing, X-Cron-Secret validation, and response bodies in cmd/bot/main.go
- [ ] T045 [US3] Send generated report or duplicate notice to configured Telegram group for external cron trigger in internal/reports/service.go

**Checkpoint**: Monthly reporting works manually and from an external scheduled HTTP request without internal timers.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate full behavior, remove rough edges, and document operation.

- [ ] T046 [P] Add command usage examples and environment variable reference in README.md
- [ ] T047 [P] Add quickstart validation notes for webhook setup, cron secret, and MongoDB Atlas indexes in README.md
- [ ] T048 Run gofmt on all Go files in cmd/ and internal/
- [ ] T049 Run go test ./... and fix any failing tests in cmd/ and internal/
- [ ] T050 Verify no implementation uses Redis, queues, WebSockets, Telegram polling, in-memory timers, background infinite loops, or web UI files in cmd/, internal/, go.mod, and README.md
- [ ] T051 Validate quickstart scenarios from specs/001-team-status-bot/quickstart.md against the implemented commands and HTTP routes, including local/manual timing checks that GET /health responds under 1 second and Telegram command handlers plus report generation respond under 5 seconds for normal MVP-sized data

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational. This is MVP.
- **User Story 2 (Phase 4)**: Depends on User Story 1 stop flow.
- **User Story 3 (Phase 5)**: Depends on User Story 1 records and User Story 2 incidents.
- **Polish (Phase 6)**: Depends on selected user stories.

### User Story Dependencies

- **US1**: Independent after Foundational.
- **US2**: Requires US1 stop operation and work records.
- **US3**: Requires people, work records, incidents, and Telegram client behavior from US1/US2.

### Within Each User Story

- Tests before implementation.
- Models and repositories before services.
- Services before Telegram handlers and HTTP routes.
- Command/HTTP wiring after service behavior exists.

---

## Parallel Opportunities

- Setup tasks T003-T006 can run in parallel after T001.
- Foundational model/helper tasks T008-T011 can run in parallel.
- US1 test tasks T017-T021 can run in parallel.
- US1 repository tasks T022-T023 can run in parallel.
- US2 test tasks T028-T030 can run in parallel.
- US3 test tasks T036-T039 can run in parallel.
- US3 repository/query tasks T040-T041 can run in parallel.
- Polish docs tasks T046-T047 can run in parallel.

## Parallel Example: User Story 1

```bash
Task: "T017 [P] [US1] Add person service tests for valid person and duplicate normalized name in internal/people/service_test.go"
Task: "T020 [P] [US1] Add Telegram command handler error tests for malformed command format, missing arguments, unknown person, no active work record, already stopped work record, and any second active work for the same person regardless of title in internal/telegram/handler_errors_test.go"
Task: "T021 [P] [US1] Add Telegram command handler test verifying MVP commands require no user role or authentication beyond configured chat validation in internal/telegram/handler_auth_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T036 [P] [US3] Add report service tests for month overlap inclusion across active, started-in-month, stopped-in-month, and cross-month records in internal/reports/service_report_test.go"
Task: "T038 [P] [US3] Add /report_month command handler tests for default current month, YYYY-MM argument, invalid month format, and duplicate notice in internal/telegram/handler_report_test.go"
Task: "T039 [P] [US3] Add POST /cron/monthly-report HTTP tests for X-Cron-Secret success, missing header, invalid header, generated response, duplicate response, and invalid month format in cmd/bot/main_test.go"
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete User Story 1.
3. Validate `/help`, `/add_person`, `/start_work`, `/status`, and `/stop_work`.
4. Stop and confirm MVP tracking works before alert/report expansion.

### Incremental Delivery

1. Add User Story 2 for stop alerts and alert failure recording.
2. Add User Story 3 for manual and externally triggered monthly reports.
3. Run quickstart validation and full tests.

### Notes

- Do not add `/reset_work` in MVP.
- Do not use Telegram polling.
- Do not add Redis, queues, WebSockets, in-memory timers, background infinite loops, or web UI.
- All elapsed time must be calculated from stored timestamps.
- Store timestamps in UTC and display dates in Europe/Kyiv timezone.
