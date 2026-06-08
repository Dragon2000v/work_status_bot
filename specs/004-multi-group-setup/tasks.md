# Tasks: Multi-Group Setup

**Input**: Design documents from `/specs/004-multi-group-setup/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`

**Tests**: Required by feature plan and quickstart. Add focused Go tests for group setup, authorization, delivery targets, and existing command fallback behavior.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or isolated test cases
- **[Story]**: User story identifier (`US1`, `US2`, `US3`)
- Every task includes exact repository paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm current project shape and prepare shared package locations without changing behavior.

- [ ] T001 Inspect existing Telegram command, callback, report, cron, logging, and MongoDB index code in `internal/telegram/`, `internal/reports/`, `cmd/bot/`, and `internal/database/`
- [ ] T002 [P] Create `internal/groups/` package skeleton with `model.go`, `repository.go`, `service.go`, and test files
- [ ] T003 [P] Review current localized command/help/menu text in `internal/i18n/` and `internal/telegram/` to identify messages that need `/setup`, `/groups`, `/disable_group`, and setup-required responses

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add durable group data and shared authorization primitives required by all user stories.

**CRITICAL**: No user story work should begin until these tasks are complete.

- [ ] T004 Define `ConfiguredGroup`, `GroupSetupRequest`, `GroupAuthorizationDecision`, decision reason constants, and `ReportDeliveryTarget` in `internal/groups/model.go`
- [ ] T005 Implement MongoDB repository methods in `internal/groups/repository.go` for setup upsert, lookup by chat id, list enabled groups, list all groups, and disable current group
- [ ] T006 Add unique `{telegram_chat_id: 1}` and `{enabled: 1}` indexes for `configured_groups` in `internal/database/indexes.go`
- [ ] T007 Implement group service validation and UTC timestamp handling in `internal/groups/service.go`
- [ ] T008 Implement authorization decisions in `internal/groups/service.go` for enabled configured groups, fallback `TELEGRAM_GROUP_CHAT_ID`, `/setup`, `/help`, unknown groups, disabled groups, and private chat setup
- [ ] T009 [P] Add group repository and service unit tests for setup upsert, idempotent setup, disabled group re-enable, enabled listing, and disable behavior in `internal/groups/*_test.go`
- [ ] T010 [P] Add authorization unit tests for configured groups, disabled groups, unknown groups, setup/help exceptions, private chat setup rejection, and env fallback allowance in `internal/groups/*_test.go`
- [ ] T011 Wire `internal/groups.Service` and repository construction into `cmd/bot/main.go` without changing webhook startup or cron route registration
- [ ] T012 Update Telegram handler construction in `internal/telegram/handler.go` and related tests to accept the group service while preserving existing user settings, callback, and fallback command behavior

**Checkpoint**: Group storage, authorization decisions, indexes, and application wiring are available for feature stories.

---

## Phase 3: User Story 1 - Register A New Telegram Group (Priority: P1) MVP

**Goal**: A new Telegram group can run `/setup`, be stored in MongoDB, and immediately use existing commands without redeploy.

**Independent Test**: Send `/setup` from an unconfigured group, verify the stored enabled group details, then send `/status` from the same group and verify it is accepted; send `/setup` from a private chat and verify it is rejected.

### Tests for User Story 1

- [ ] T013 [P] [US1] Add Telegram handler tests for `/setup` in a group, private chat setup rejection, existing enabled group setup update, disabled group setup re-enable, and missing username handling in `internal/telegram/handler_test.go`
- [ ] T014 [P] [US1] Add integration-style service tests for idempotent setup preserving `created_at`, updating `updated_at`, title, setup user fields, `enabled=true`, and missing or blank group title normalization to `Unknown group` in `internal/groups/service_test.go`
- [ ] T015 [P] [US1] Add fallback group command acceptance test proving existing commands still work for `TELEGRAM_GROUP_CHAT_ID` without a stored group in `internal/telegram/handler_test.go`

### Implementation for User Story 1

- [ ] T016 [US1] Add `/setup` command constant and parsing support in `internal/telegram/commands.go`
- [ ] T017 [US1] Implement `/setup` handling in `internal/telegram/handler.go` using current chat id, group title, setup user id, optional username, and group service setup result
- [ ] T018 [US1] Add localized Ukrainian-default setup responses and help text entries for setup success, already completed, re-enabled, private-chat rejection, and setup failure in `internal/i18n/`
- [ ] T019 [US1] Add structured logs for `telegram.group_setup` attempt, created, updated, re_enabled, rejected, and failure outcomes in `internal/telegram/handler.go`
- [ ] T020 [US1] Ensure existing operational command flow in `internal/telegram/handler.go` accepts commands from newly configured enabled groups after setup
- [ ] T021 [US1] Document skipped admin verification as an MVP limitation in `README.md`

**Checkpoint**: User Story 1 is functional and independently testable.

---

## Phase 4: User Story 2 - Reject Unconfigured Groups Safely (Priority: P2)

**Goal**: Unknown or disabled groups cannot run operational commands, while `/setup` and `/help` remain available for onboarding.

**Independent Test**: Send `/status`, `/add_person`, `/start_work`, `/stop_work`, `/report_month`, `/groups`, and `/disable_group` from an unknown group and verify clear rejection; verify `/help` and `/setup` still work.

### Tests for User Story 2

- [ ] T022 [P] [US2] Add Telegram handler tests for setup-required rejection of `/status`, `/add_person`, `/start_work`, `/stop_work`, `/report_month`, `/groups`, and `/disable_group` from unknown groups in `internal/telegram/handler_test.go`
- [ ] T023 [P] [US2] Add Telegram handler tests proving `/help` and `/setup` are allowed from unknown groups in `internal/telegram/handler_test.go`
- [ ] T024 [P] [US2] Add tests proving disabled groups are rejected for operational commands unless they match fallback group id in `internal/groups/service_test.go`

### Implementation for User Story 2

- [ ] T025 [US2] Route every text command and callback action through group authorization in `internal/telegram/handler.go` before operational behavior executes
- [ ] T026 [US2] Add setup-required rejection responses with Ukrainian default localization and existing per-user language lookup for user-specific interactions in `internal/i18n/` and `internal/telegram/messages.go`
- [ ] T027 [US2] Keep `/help` available in unknown groups and update help output to mention `/setup`, `/groups`, and `/disable_group` in `internal/telegram/handler.go` and `internal/i18n/`
- [ ] T028 [US2] Add structured `telegram.group_authorization` rejection logs with chat id, user id, command, outcome, and safe reason in `internal/telegram/handler.go`
- [ ] T029 [US2] Verify command message deletion remains best-effort after authorization rejection and does not delete bot responses in `internal/telegram/handler.go`

**Checkpoint**: User Stories 1 and 2 work independently and preserve fallback command behavior.

---

## Phase 5: User Story 3 - View And Disable Configured Groups (Priority: P3)

**Goal**: Authorized groups can list configured groups and disable the current stored group without deleting history.

**Independent Test**: Configure two groups, disable one group, run `/groups` from an enabled configured group or fallback group, verify both stored groups appear with title/chat id/enabled/disabled status, then verify `/groups` and other operational commands from the disabled group are rejected until `/setup` re-enables it.

### Tests for User Story 3

- [ ] T030 [P] [US3] Add Telegram handler tests proving enabled configured groups can run `/groups`, fallback env group can run `/groups`, disabled configured groups cannot run `/groups`, unknown groups cannot run `/groups`, private chats cannot run `/groups`, no stored groups with fallback configured is handled, and allowed `/groups` output includes disabled groups with enabled/disabled status in `internal/telegram/handler_test.go`
- [ ] T031 [P] [US3] Add Telegram handler tests for `/disable_group` disabling current group, idempotent already-disabled response, fallback-only no stored group response, and unknown group rejection in `internal/telegram/handler_test.go`
- [ ] T032 [P] [US3] Add service tests for multiple configured groups and disable-without-delete behavior in `internal/groups/service_test.go`

### Implementation for User Story 3

- [ ] T033 [US3] Add `/groups` and `/disable_group` command constants and parsing support in `internal/telegram/commands.go`
- [ ] T034 [US3] Implement `/groups` handling in `internal/telegram/handler.go` so only enabled configured groups or fallback env group can execute it, and allowed output uses all stored group service list results with title, Telegram chat id, and enabled/disabled status
- [ ] T035 [US3] Implement `/disable_group` handling in `internal/telegram/handler.go` using group service disable result, idempotent responses, and fallback-only no stored group handling
- [ ] T036 [US3] Add structured logs for `telegram.group_list` and `telegram.group_disable` success, idempotent, rejected, and failure outcomes in `internal/telegram/handler.go`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Reports, Alerts, Cron, And Cross-Cutting Concerns

**Purpose**: Preserve existing report, alert, cron, logging, and documentation behavior across multiple groups.

- [ ] T037 [P] Add report service tests for command-driven monthly reports targeting the current chat instead of a single env group in `internal/reports/service_test.go` or `internal/telegram/handler_test.go`
- [ ] T038 [P] Add alert tests proving `/stop_work` sends alerts to the current group chat id in `internal/telegram/handler_test.go`
- [ ] T039 [P] Add cron tests proving monthly report fan-out sends to all enabled configured groups and fallback once when not duplicated in `cmd/bot/main_test.go` or `internal/reports/service_test.go`
- [ ] T040 Update `internal/reports/service.go` so manual report delivery can target the current chat and cron delivery can iterate enabled configured groups plus deduplicated fallback group
- [ ] T041 Update cron monthly report handler wiring in `cmd/bot/main.go` to obtain enabled group targets from `internal/groups.Service` while preserving existing secret validation and month validation
- [ ] T042 Update stop-work alert delivery in `internal/telegram/handler.go` to send alerts to the group where the stop action happened
- [ ] T043 Add structured logs for `cron.monthly_report_send` per target and `telegram.alert_send` target chat id with safe error categories in `internal/reports/service.go` and `internal/telegram/handler.go`
- [ ] T044 [P] Update contracts or README operator notes if implementation details differ, especially admin-verification MVP limitation and fallback group behavior in `README.md`
- [ ] T045 Run `gofmt` on changed Go files
- [ ] T046 Run `go test ./...` and fix regressions in existing webhook, MongoDB, structured logging, i18n, callback, command deletion, reports, alerts, and cron behavior, including verifying every `callback_query` still calls `answerCallbackQuery` after group authorization changes
- [ ] T047 Run lightweight manual validation from `specs/004-multi-group-setup/quickstart.md` confirming `/setup` completes in under 1 minute, normal command handlers complete under 5 seconds for MVP-sized data, and cron monthly report completes under 5 seconds for MVP-sized data

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No dependencies.
- **Phase 2**: Depends on Phase 1 and blocks all user stories.
- **Phase 3 (US1)**: Depends on Phase 2 and delivers the MVP path for group registration.
- **Phase 4 (US2)**: Depends on Phase 2; can run after or in parallel with US1 implementation once authorization primitives exist.
- **Phase 5 (US3)**: Depends on Phase 2; benefits from US2 authorization behavior but remains independently testable.
- **Phase 6**: Depends on selected user stories and validates cross-story report, alert, cron, and documentation behavior.

### User Story Dependencies

- **US1 (P1)**: Core MVP. Requires foundational storage and authorization primitives only.
- **US2 (P2)**: Requires foundational authorization decisions. Integrates with US1 setup but can be tested with seeded group states.
- **US3 (P3)**: Requires foundational storage/service methods. Integrates with US2 rejection behavior after disable.

### Within Each User Story

- Write tests before implementation tasks for that story.
- Add command constants before handler behavior.
- Implement service behavior before Telegram handler wiring where possible.
- Add localized responses and structured logs with the behavior they describe.
- Keep each story passing independently before moving to lower-priority work.

---

## Parallel Opportunities

- T002 and T003 can run in parallel after T001.
- T009 and T010 can run in parallel after T004-T008 are drafted.
- T013, T014, and T015 can run in parallel for US1 tests.
- T022, T023, and T024 can run in parallel for US2 tests.
- T030, T031, and T032 can run in parallel for US3 tests.
- T037, T038, T039, and T044 can run in parallel once the relevant service seams exist.
- Different user stories can proceed in parallel after Phase 2 if contributors coordinate changes to `internal/telegram/handler.go`.

---

## Parallel Example: User Story 1

```bash
# After Phase 2 primitives exist, these tests can be implemented together:
Task: "T013 [P] [US1] Add Telegram handler tests for /setup in internal/telegram/handler_test.go"
Task: "T014 [P] [US1] Add service tests for idempotent setup in internal/groups/service_test.go"
Task: "T015 [P] [US1] Add fallback group command acceptance test in internal/telegram/handler_test.go"
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 only.
3. Validate that a new group can run `/setup`, use `/status`, and private chat setup is rejected.
4. Run targeted tests for `internal/groups` and `internal/telegram`.

### Incremental Delivery

1. Add US1 group registration and fallback compatibility.
2. Add US2 rejection behavior for unknown and disabled groups.
3. Add US3 `/groups` and `/disable_group`.
4. Add Phase 6 report, alert, cron fan-out, logging, documentation, and full regression testing.

### Validation Commands

```bash
go test ./internal/groups ./internal/telegram
go test ./internal/reports ./cmd/bot
go test ./...
```
