# Tasks: Telegram Reply Keyboard, Languages, And Start Work Flow

**Input**: Design documents from `specs/005-reply-keyboard-start-flow/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Tests are required by FR-038 and the user request. Test tasks are listed before implementation tasks in each story phase.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently after the foundational phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or depends only on completed foundation
- **[Story]**: User story label for story phases only
- Every task includes concrete file paths

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the existing Go service surface and create package/test files needed by later phases.

- [X] T001 Verify active feature context points to `specs/005-reply-keyboard-start-flow` in `.specify/feature.json`
- [X] T002 Verify active Spec Kit plan points to `specs/005-reply-keyboard-start-flow/plan.md` in `AGENTS.md`
- [X] T003 [P] Create package directory for date parsing in `internal/datetime/`
- [X] T004 [P] Create package directory for flow state in `internal/flows/`
- [X] T005 Review existing Telegram client, handler, keyboard, callback, stop, and delete tests before edits in `internal/telegram/client_test.go`, `internal/telegram/handler_test.go`, `internal/telegram/keyboard_test.go`, `internal/telegram/handler_callback_test.go`, `internal/telegram/handler_stop_test.go`, and `internal/telegram/handler_delete_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared Telegram API types, persistence models, i18n keys, parser service, and flow-state service required before user stories.

**CRITICAL**: No user story work should begin until this phase is complete.

### Tests

- [X] T006 [P] Add user flow state model validation tests for supported steps, UTC timestamps, and `expires_at = updated_at + 15 minutes` in `internal/flows/model_test.go`
- [X] T007 [P] Add user flow state repository tests for upsert, lookup by user/chat, completion delete, expired lookup, stale-state clear, and `expires_at` index expectations in `internal/flows/repository_test.go`
- [X] T008 [P] Add user flow service tests for start, advance, complete, cancel, expire, restart, and refresh of `updated_at` plus `expires_at` on update in `internal/flows/service_test.go`
- [X] T009 [P] Add date/time parser tests for `DD.MM HH:mm`, `DD.MM.YYYY HH:mm`, `YYYY-MM-DD HH:mm`, `HH:mm` Today/Yesterday, omitted current Kyiv year, invalid values, and future rejection in `internal/datetime/parser_test.go`
- [X] T010 [P] Add i18n tests for new menu, settings, add-person, stop-work, start-work, cancellation, expired-flow, and date/time validation keys in `internal/i18n/i18n_test.go`
- [X] T011 [P] Add works service tests for explicit UTC `started_at` support while keeping default now-based start behavior in `internal/works/service_test.go`

### Implementation

- [X] T012 Define `FlowState`, flow type constants, step constants, payload helpers, and 15-minute expiration helpers in `internal/flows/model.go`
- [X] T013 Implement MongoDB `user_flow_states` repository with unique `telegram_user_id` + `chat_id` lookup/upsert/delete and expired-state clear in `internal/flows/repository.go`
- [X] T014 Implement flow service for start, get active non-expired, advance with refreshed `updated_at`/`expires_at`, complete, cancel, expire, and restart behavior in `internal/flows/service.go`
- [X] T015 Add `user_flow_states` unique compound index and `expires_at` index without background cleanup loops in `internal/database/indexes.go`
- [X] T016 Implement Europe/Kyiv date/time parser with UTC output, future rejection, invalid-format errors, and current-year omitted-year rule in `internal/datetime/parser.go`
- [X] T017 Add Ukrainian, English, and Russian dictionary keys for reply keyboard labels, settings labels, add-person prompts, stop-work prompts, start-work prompts, cancellation, expired flow, and date/time validation examples in `internal/i18n/i18n.go`
- [X] T018 Add Telegram `ReplyKeyboardMarkup` and `KeyboardButton` models while keeping existing inline keyboard models in `internal/telegram/client.go`
- [X] T019 Generalize Telegram `SendMessageWithKeyboard` to accept reply or inline markup without breaking existing callback send paths in `internal/telegram/client.go`
- [X] T020 Add flow service dependency wiring fields and constructor parameters to the Telegram handler in `internal/telegram/handler.go`
- [X] T021 Wire `internal/flows` repository/service into application startup in `cmd/bot/main.go`
- [X] T022 Add `StartAt` or equivalent explicit UTC start method to works service without breaking existing `Start` behavior in `internal/works/service.go`

**Checkpoint**: Foundation ready. Flow state, datetime parsing, i18n keys, Telegram API models, and explicit start timestamp support are available.

---

## Phase 3: User Story 1 - Use Persistent Reply Keyboard Menu (Priority: P1) MVP

**Goal**: Users see and use a persistent Telegram reply keyboard main menu, including full Add Person and Stop Work button flows.

**Independent Test**: In an authorized group, the bot shows a persistent Ukrainian reply keyboard, `Статус` runs status, `Додати людину` completes add-person flow, `Зупинити роботу` completes stop-work flow with incident/alert behavior, and no main-menu inline keyboard is emitted.

### Tests

- [X] T023 [P] [US1] Add reply keyboard rendering tests for Ukrainian labels, `resize_keyboard=true`, `one_time_keyboard=false`, and `is_persistent=true` in `internal/telegram/keyboard_test.go`
- [X] T024 [P] [US1] Add test proving `MainMenuKeyboard` no longer emits `InlineKeyboardMarkup` for main-menu actions in `internal/telegram/keyboard_test.go`
- [X] T025 [P] [US1] Add keyboard text mapping tests for Ukrainian labels to canonical actions in `internal/telegram/commands_test.go`
- [X] T026 [P] [US1] Add webhook handler test for `Статус` reply keyboard text returning status in `internal/telegram/handler_test.go`
- [X] T027 [P] [US1] Add handler regression test that unknown button text outside a flow returns a localized unsupported action in `internal/telegram/handler_errors_test.go`
- [X] T028 [P] [US1] Add Add Person button test proving `Додати людину` starts `add_person_enter_name` flow state in `internal/telegram/handler_add_person_test.go`
- [X] T029 [P] [US1] Add Add Person valid-name test proving person creation, localized success, input deletion attempt, and flow clear in `internal/telegram/handler_add_person_test.go`
- [X] T030 [P] [US1] Add Add Person invalid-name, duplicate-person, cancel, expired-flow, and slash `/add_person` regression tests in `internal/telegram/handler_add_person_test.go`
- [X] T031 [P] [US1] Add Stop Work button test proving `Зупинити роботу` starts `stop_work_enter_person_or_reason` flow state in `internal/telegram/handler_stop_flow_test.go`
- [X] T032 [P] [US1] Add Stop Work valid-input test proving active work stop, incident/event recording, current-group alert send attempt, localized result, input deletion attempt, and flow clear in `internal/telegram/handler_stop_flow_test.go`
- [X] T033 [P] [US1] Add Stop Work alert-failure test proving stop remains successful, incident/event is saved, alert failure is recorded, and localized warning behavior remains unchanged in `internal/telegram/handler_stop_flow_test.go`
- [X] T034 [P] [US1] Add Stop Work unknown-person, no-active-work, already-stopped where applicable, cancel, expired-flow, and slash `/stop_work` regression tests in `internal/telegram/handler_stop_flow_test.go`

### Implementation

- [X] T035 [US1] Replace main menu inline keyboard builder with `ReplyKeyboardMarkup` builder in `internal/telegram/keyboard.go`
- [X] T036 [US1] Add settings, add-person, stop-work, start-work, and flow-choice reply keyboard builders in `internal/telegram/keyboard.go`
- [X] T037 [US1] Implement reply keyboard label-to-action routing for Ukrainian and supported localized labels in `internal/telegram/commands.go`
- [X] T038 [US1] Route reply keyboard messages through canonical command/action handling before malformed command rejection in `internal/telegram/handler.go`
- [X] T039 [US1] Send the main reply keyboard with help/menu responses using Ukrainian group-wide labels when group keyboard personalization is unreliable in `internal/telegram/handler.go`
- [X] T040 [US1] Remove main-menu callback action emission while preserving callback constants for non-main-menu regression behavior in `internal/telegram/keyboard.go` and `internal/telegram/callbacks.go`
- [X] T041 [US1] Implement Add Person flow start, `add_person_enter_name` state storage, valid-name parsing, existing people service creation, localized success, input cleanup request, cancellation, expiry clear, and flow clear in `internal/telegram/handler.go`
- [X] T042 [US1] Implement Add Person localized validation errors with examples and preserve existing duplicate-person behavior in `internal/telegram/messages.go` and `internal/telegram/handler.go`
- [X] T043 [US1] Implement Stop Work flow start, `stop_work_enter_person_or_reason` state storage, stop input parsing, existing works service stop, incident/event recording, current-group alert attempt, localized result, input cleanup request, cancellation, expiry clear, and flow clear in `internal/telegram/handler.go`
- [X] T044 [US1] Preserve Stop Work alert failure behavior by recording alert failure while keeping successful stop and incident/event save in `internal/telegram/handler.go` and `internal/reports/service.go`
- [X] T045 [US1] Implement Stop Work localized errors for unknown person, no active work, and already-stopped where applicable in `internal/telegram/messages.go` and `internal/telegram/handler.go`
- [X] T046 [US1] Add structured logs for `telegram.reply_keyboard_action`, `telegram.flow_started`, `telegram.flow_step_advanced`, `telegram.flow_completed`, `telegram.flow_cancelled`, and `telegram.flow_expired` in `internal/telegram/handler.go`

**Checkpoint**: User Story 1 is independently functional and provides the MVP visible menu plus Add Person and Stop Work button flows.

---

## Phase 4: User Story 2 - Use Native Telegram Command Menu And Slash Fallback (Priority: P1)

**Goal**: Telegram native command menu is registered at startup and all slash commands keep working even if command menu registration fails.

**Independent Test**: Startup attempts `setMyCommands`; success/failure is logged without secrets; failure does not stop startup; `/setup`, `/groups`, `/disable_group`, `/add_person`, `/start_work`, `/status`, `/stop_work`, `/report_month`, and `/help` still work through slash commands.

### Tests

- [X] T047 [P] [US2] Add Telegram client `setMyCommands` payload, success, and API failure tests in `internal/telegram/client_test.go`
- [X] T048 [P] [US2] Add startup registration attempt/success/failure logging tests without bot token, webhook secret, or cron secret leakage in `cmd/bot/main_test.go`
- [X] T049 [P] [US2] Add startup test proving `setMyCommands` failure does not stop router/startup wiring and the bot still handles commands and reply keyboard actions in `cmd/bot/main_test.go`
- [X] T050 [P] [US2] Add slash command regression tests for `/status`, `/help`, `/groups`, and `/disable_group` in `internal/telegram/handler_test.go` and `internal/telegram/handler_groups_test.go`
- [X] T051 [P] [US2] Add `/start_work <first> <last> "<title>"` regression test proving existing direct command behavior still works in `internal/telegram/handler_test.go`
- [X] T052 [P] [US2] Add callback query regression test proving existing callback updates still call `answerCallbackQuery` in `internal/telegram/handler_callback_test.go`

### Implementation

- [X] T053 [US2] Add Telegram `BotCommand` model and `SetMyCommands` client method in `internal/telegram/client.go`
- [X] T054 [US2] Define native command list and localized descriptions for `/setup`, `/groups`, `/disable_group`, `/add_person`, `/start_work`, `/status`, `/stop_work`, `/report_month`, and `/help` in `internal/telegram/commands.go`
- [X] T055 [US2] Register native Telegram commands on startup after Telegram client creation and before serving traffic in `cmd/bot/main.go`
- [X] T056 [US2] Log command menu registration attempt, success, and failure with `telegram.command_menu_registration` and no secrets in `cmd/bot/main.go` and `internal/telegram/client.go`
- [X] T057 [US2] Continue application startup when `setMyCommands` fails while preserving slash command and reply keyboard handling in `cmd/bot/main.go`
- [X] T058 [US2] Normalize slash commands with bot username suffix such as `/status@botname` in `internal/telegram/commands.go`
- [X] T059 [US2] Preserve direct slash command behavior for commands with full arguments while allowing bare flow-start commands to begin flows in `internal/telegram/handler.go`
- [X] T060 [US2] Preserve existing callback query handling and `answerCallbackQuery` behavior for non-main-menu callbacks in `internal/telegram/handler.go`

**Checkpoint**: Native command menu and slash fallback are independently functional.

---

## Phase 5: User Story 3 - Remember Per-User Language Preference (Priority: P2)

**Goal**: Users can change language only in Settings, and personal interactions use saved user language while public group output remains Ukrainian.

**Independent Test**: One user changes to English, another remains Ukrainian, service restarts, personal prompts use saved language, and public status/alerts/reports/incident messages remain Ukrainian.

### Tests

- [X] T061 [P] [US3] Add default Ukrainian language lookup tests in `internal/users/service_test.go`
- [X] T062 [P] [US3] Add saved user language lookup and unsupported language normalization tests in `internal/users/service_test.go`
- [X] T063 [P] [US3] Add user settings repository upsert timestamp and unique user id tests in `internal/users/repository_test.go`
- [X] T064 [P] [US3] Add settings flow handler tests for `Налаштування`, `Мова`, `Українська`, `English`, `Русский`, `Назад`, and `Скасувати` in `internal/telegram/handler_language_test.go`
- [X] T065 [P] [US3] Add localized personal prompt and button label tests for Ukrainian, English, and Russian in `internal/telegram/messages_test.go` and `internal/telegram/keyboard_test.go`
- [X] T066 [P] [US3] Add public output language regression tests proving group status, alerts, incident/alert messages, and monthly reports use Ukrainian by default in `internal/telegram/handler_report_test.go`, `internal/telegram/handler_stop_test.go`, and `internal/reports/service_report_test.go`

### Implementation

- [X] T067 [US3] Extend `internal/i18n/i18n.go` with all Settings and language selection labels for Ukrainian, English, and Russian
- [X] T068 [US3] Ensure `internal/users/service.go` returns Ukrainian for missing user settings and only mutates language through explicit Settings selection
- [X] T069 [US3] Implement Settings flow state transitions for `settings_language_select`, language save, Back, and Cancel in `internal/telegram/handler.go`
- [X] T070 [US3] Use saved user language for personal responses, Settings messages, Help requested by a user, and flow prompts in `internal/telegram/handler.go` and `internal/telegram/messages.go`
- [X] T071 [US3] Force public group status output, alerts, incident/alert messages, and monthly reports to Ukrainian by default in `internal/telegram/handler.go`, `internal/telegram/messages.go`, and `internal/reports/service.go`
- [X] T072 [US3] Add structured log `telegram.user_language_changed` for success and failure in `internal/telegram/handler.go`

**Checkpoint**: Language preference behavior is independently functional.

---

## Phase 6: User Story 4 - Start Work From Buttons (Priority: P2)

**Goal**: Users can start work through buttons and manual date/time choices with UTC storage.

**Independent Test**: Press `Почати роботу`, enter person/title, choose now/today/yesterday/manual date, verify active work record has expected UTC start and Kyiv display; invalid/future inputs are rejected.

### Tests

- [X] T073 [P] [US4] Add Start Work flow creation and person prompt tests for reply keyboard button and bare `/start_work` in `internal/telegram/handler_start_work_test.go`
- [X] T074 [P] [US4] Add Start Work flow title prompt and time mode selection tests in `internal/telegram/handler_start_work_test.go`
- [X] T075 [P] [US4] Add Start Work cancellation, expired-flow clear, and localized expired-flow message tests in `internal/telegram/handler_start_work_test.go`
- [X] T076 [P] [US4] Add parser tests for `DD.MM HH:mm`, `DD.MM.YYYY HH:mm`, `YYYY-MM-DD HH:mm`, and `HH:mm` Today/Yesterday in `internal/datetime/parser_test.go`
- [X] T077 [P] [US4] Add parser tests for omitted year using current Kyiv year, future rejection, invalid format examples, and impossible dates in `internal/datetime/parser_test.go`
- [X] T078 [P] [US4] Add integration-style handler test proving Start Work flow saves UTC timestamp and status displays Europe/Kyiv time in `internal/telegram/handler_start_work_test.go`

### Implementation

- [X] T079 [US4] Add Start Work flow prompts and validation messages to `internal/i18n/i18n.go` and `internal/telegram/messages.go`
- [X] T080 [US4] Implement Start Work flow state transitions for person, title, time mode, time input, manual datetime input, completion, cancellation, and expiration in `internal/telegram/handler.go`
- [X] T081 [US4] Connect `internal/datetime` parser to Start Work flow with Europe/Kyiv parse, UTC storage, future rejection, and localized examples in `internal/telegram/handler.go`
- [X] T082 [US4] Treat omitted-year manual dates as current Europe/Kyiv year and reject if that resolved timestamp is in the future in `internal/datetime/parser.go`
- [X] T083 [US4] Support Today and Yesterday `HH:mm` choices with Europe/Kyiv date selection in `internal/datetime/parser.go` and `internal/telegram/handler.go`
- [X] T084 [US4] Keep `/start_work <first> <last> "<title>"` direct command path working while bare `/start_work` starts the button-driven flow in `internal/telegram/handler.go`
- [X] T085 [US4] Add structured log `telegram.datetime_parse` for parse success, invalid format, and future rejection in `internal/telegram/handler.go`

**Checkpoint**: Start Work button flow is independently functional.

---

## Phase 7: User Story 5 - Clean Up User Messages Without Hiding Bot Output (Priority: P3)

**Goal**: User slash command, reply keyboard, and manual flow input messages are deleted best-effort; bot outputs are never deleted.

**Independent Test**: In chats with and without delete permission, applicable user messages trigger delete attempts, failures are logged, workflows continue, and bot responses/reports/alerts remain visible.

### Tests

- [X] T086 [P] [US5] Add deleteMessage client payload and API failure tests in `internal/telegram/client_test.go`
- [X] T087 [P] [US5] Add handler tests for slash command delete attempt, success, and failure continuation in `internal/telegram/handler_delete_test.go`
- [X] T088 [P] [US5] Add handler tests for reply keyboard button text deletion in `internal/telegram/handler_delete_test.go`
- [X] T089 [P] [US5] Add handler tests for Add Person, Stop Work, Start Work, and Settings manual flow input deletion in `internal/telegram/handler_delete_test.go`
- [X] T090 [P] [US5] Add regression tests proving bot responses, status output, alerts, reports, incident messages, and monthly reports are never deletion targets in `internal/telegram/handler_delete_test.go`, `internal/telegram/handler_report_test.go`, and `internal/telegram/handler_stop_test.go`

### Implementation

- [X] T091 [US5] Generalize command deletion helper into user message cleanup by message kind in `internal/telegram/handler.go`
- [X] T092 [US5] Attempt cleanup after slash command processing, reply keyboard action processing, and manual flow input processing in `internal/telegram/handler.go`
- [X] T093 [US5] Ensure cleanup never receives bot response/report/alert/incident/monthly report message ids in `internal/telegram/handler.go` and `internal/reports/service.go`
- [X] T094 [US5] Add `telegram.message_delete` attempt, success, and failure structured logs with `message_kind` in `internal/telegram/handler.go`

**Checkpoint**: Message cleanup is independently functional and non-destructive to bot output.

---

## Phase 8: Polish & Cross-Cutting Regression

**Purpose**: End-to-end regression, banned-tech scan, and documentation alignment across all user stories.

- [X] T095 [P] Add multi-group authorization regression tests for slash commands, reply keyboard actions, and flow inputs in `internal/telegram/handler_auth_test.go` and `internal/telegram/handler_groups_test.go`
- [X] T096 [P] Add group setup regression tests for `/setup`, `/groups`, and `/disable_group` after command menu changes in `internal/telegram/handler_groups_test.go`
- [X] T097 [P] Add cron/report/alert regression tests covering monthly report, stop alert, incident/event recording, current-group incident/alert message, Ukrainian public incident/alert output, alert failure recording, and monthly report incident inclusion in `cmd/bot/main_test.go`, `internal/reports/service_test.go`, `internal/reports/service_report_test.go`, `internal/telegram/handler_stop_test.go`, and `internal/telegram/handler_report_test.go`
- [X] T098 [P] Add structured logging regression tests for reply keyboard, flow lifecycle, language change, command menu registration, message deletion, and datetime parse events in `internal/telegram/handler_logging_test.go` and `cmd/bot/main_test.go`
- [X] T099 [P] Update README usage examples for reply keyboard, native commands, Settings language, Add Person flow, Stop Work flow, and Start Work manual date/time formats in `README.md`
- [X] T100 Run banned runtime scan for polling, Redis, queues, WebSockets, in-memory timers, background infinite loops, and web UI references in `cmd/`, `internal/`, `go.mod`, and `README.md`
- [X] T101 Run quickstart validation scenarios from `specs/005-reply-keyboard-start-flow/quickstart.md`
- [X] T102 Run full test suite with `go test ./...` from repository root `.`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: no dependencies
- **Phase 2 Foundational**: depends on Phase 1 and blocks all user stories
- **Phase 3 US1**: depends on Phase 2; suggested MVP first
- **Phase 4 US2**: depends on Phase 2; can run alongside US1 after shared Telegram models exist
- **Phase 5 US3**: depends on Phase 2 and benefits from US1 keyboard routing
- **Phase 6 US4**: depends on Phase 2 and benefits from US1/US3 prompts and routing
- **Phase 7 US5**: depends on Phase 2 and benefits from US1/US4 message classification
- **Phase 8 Polish**: depends on selected story phases being complete

### User Story Dependencies

- **US1**: independent after foundation; MVP-visible reply keyboard plus Add Person and Stop Work button flows
- **US2**: independent after foundation; native commands and slash fallback
- **US3**: depends on foundation; integrates with US1 keyboard/settings surfaces
- **US4**: depends on foundation; integrates with US1 routing and US3 localization
- **US5**: depends on foundation; best after US1/US4 so all message kinds exist

### Parallel Opportunities

- T003 and T004 can run in parallel.
- T006-T011 can run in parallel because they target different packages.
- Test tasks within each user story can run in parallel.
- Implementation tasks in separate packages can run in parallel after their tests and dependencies are ready.
- Phase 8 regression tests T095-T099 can run in parallel.

## Parallel Example: User Story 1

```text
Task: "T023 [P] [US1] Add reply keyboard rendering tests in internal/telegram/keyboard_test.go"
Task: "T028 [P] [US1] Add Add Person flow-start test in internal/telegram/handler_add_person_test.go"
Task: "T031 [P] [US1] Add Stop Work flow-start test in internal/telegram/handler_stop_flow_test.go"
```

## Parallel Example: User Story 4

```text
Task: "T076 [P] [US4] Add parser tests in internal/datetime/parser_test.go"
Task: "T078 [P] [US4] Add handler UTC timestamp integration test in internal/telegram/handler_start_work_test.go"
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1) to replace the main menu and deliver Add Person plus Stop Work button flows.
3. Complete Phase 4 (US2) to preserve slash fallback and register native commands.
4. Run US1/US2 tests before adding language and start-work behavior.

### Incremental Delivery

1. Add US3 language/settings persistence.
2. Add US4 Start Work flow with date/time parser.
3. Add US5 cleanup once all message kinds exist.
4. Finish Phase 8 regressions and quickstart validation.

### Validation Rules

- Every story must pass its tests independently before moving on.
- Existing webhook, multi-group setup, group authorization, reports, alerts, incidents, cron, and callback_query behavior must not regress.
- Flow states expire 15 minutes after `updated_at`; expired states are cleared before localized expired-flow response.
- `setMyCommands` failure is logged without secrets and does not stop startup.
- No task may introduce web UI, polling, Redis, queues, WebSockets, in-memory timers, or background infinite loops.
