# Tasks: Telegram Buttons And User Languages

**Input**: Design documents from `/specs/003-telegram-buttons-languages/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Focused Go test tasks are included because the feature specification requires
coverage for default Ukrainian language, saved language lookup, language changes,
callback routing, menu rendering, localized button labels, localized command responses,
command deletion, and fallback commands.

**Organization**: Tasks are grouped by user story so each story can be implemented and
validated independently after the foundational phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency
  on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm active feature context and add file structure needed by the plan.

- [X] T001 Verify `.specify/feature.json` points to `specs/003-telegram-buttons-languages` in .specify/feature.json
- [X] T002 [P] Create placeholder package files for i18n helpers in internal/i18n/i18n.go and internal/i18n/i18n_test.go
- [X] T003 [P] Create placeholder package files for user settings in internal/users/model.go, internal/users/repository.go, internal/users/service.go, and internal/users/service_test.go
- [X] T004 [P] Create placeholder Telegram helper files in internal/telegram/callbacks.go and internal/telegram/keyboard.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared language, persistence, Telegram API, and logging primitives required
before user story work.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T005 [P] Add i18n tests for default Ukrainian fallback, unsupported language fallback, missing key fallback, and uk/en/ru lookup in internal/i18n/i18n_test.go
- [X] T006 Implement language constants, translation keys, and in-code uk/en/ru dictionaries for shared menu/help/error labels in internal/i18n/i18n.go
- [X] T007 [P] Add user settings service tests for missing user fallback to Ukrainian, saved language lookup, invalid language rejection, and same-language no-op in internal/users/service_test.go
- [X] T008 Implement UserSetting model with telegram_user_id, language, created_at, and updated_at fields in internal/users/model.go
- [X] T009 Implement user settings repository interface and MongoDB repository methods GetByTelegramUserID and UpsertLanguage in internal/users/repository.go
- [X] T010 Implement user settings service with LookupLanguage and SetLanguage using UTC timestamps in internal/users/service.go
- [X] T011 Add unique `telegram_user_id` index for MongoDB collection `user_settings` in internal/database/indexes.go
- [X] T012 [P] Add Telegram client tests for sendMessage with inline keyboard payload, answerCallbackQuery payload, deleteMessage payload, and editMessageText payload in internal/telegram/client_test.go
- [X] T013 Extend Telegram client with InlineKeyboardMarkup structs and SendMessageWithKeyboard, AnswerCallbackQuery, DeleteMessage, and EditMessageText methods in internal/telegram/client.go
- [X] T014 [P] Add Telegram update decoding tests for message_id and callback_query fields in internal/telegram/handler_test.go
- [X] T015 Extend Telegram Update model with message_id and callback_query fields in internal/telegram/handler.go
- [X] T016 Update cmd/bot/main.go to construct the users repository/service and pass it into the Telegram handler in cmd/bot/main.go

**Checkpoint**: Shared i18n, user settings persistence, Telegram Bot API methods, and update shapes exist and compile.

---

## Phase 3: User Story 1 - Use Bot With Inline Buttons (Priority: P1) MVP

**Goal**: A group user can use inline buttons for main bot actions instead of typing
commands.

**Independent Test**: Open the main menu in the configured Telegram group, press each
main action button, and confirm the bot starts or completes the same user-visible action
as the corresponding text command with Ukrainian labels by default.

### Tests for User Story 1

- [X] T017 [P] [US1] Add keyboard rendering tests for Ukrainian main menu button labels and stable callback_data in internal/telegram/keyboard_test.go
- [X] T018 [P] [US1] Add callback routing tests for menu:status and menu:help actions in internal/telegram/handler_callback_test.go
- [X] T019 [P] [US1] Add tests that Add person, Start work, Stop work, and Monthly report buttons return localized usage guidance or trigger the existing command-equivalent flow in internal/telegram/handler_callback_test.go
- [X] T020 [P] [US1] Add localized help/menu response tests that include both button and text command options in internal/telegram/messages_test.go

### Implementation for User Story 1

- [X] T021 [US1] Define callback action constants for menu:add_person, menu:start_work, menu:status, menu:stop_work, menu:report_month, menu:settings, and menu:help in internal/telegram/callbacks.go
- [X] T022 [US1] Implement MainMenuKeyboard using InlineKeyboardMarkup and localized labels in internal/telegram/keyboard.go
- [X] T023 [US1] Add localized help text, menu text, and button-action usage prompts in internal/telegram/messages.go
- [X] T024 [US1] Extend the Telegram handler interface to send messages with inline keyboards in internal/telegram/handler.go
- [X] T025 [US1] Implement callback routing for main menu actions using existing status, help, and report services where available in internal/telegram/handler.go
- [X] T026 [US1] Update `/help` handling to send a localized main menu with InlineKeyboardMarkup in internal/telegram/handler.go

**Checkpoint**: User Story 1 is functional: the main menu renders, buttons route, and existing command-equivalent actions remain available through buttons.

---

## Phase 4: User Story 2 - Remember Each User's Language (Priority: P2)

**Goal**: A user can select Ukrainian, English, or Russian in Settings and future
user-specific interactions use that saved language without repeated prompts.

**Independent Test**: Change language in Settings, perform Settings, Help, and button
interactions later, and confirm that user sees the saved language while another user
without a saved language still sees Ukrainian.

### Tests for User Story 2

- [X] T027 [P] [US2] Add repository tests for user_settings upsert, created_at preservation, updated_at change, and telegram_user_id lookup in internal/users/repository_test.go
- [X] T028 [P] [US2] Add settings menu rendering tests for "🌐 Мова / Language / Язык", Українська, English, and Русский in internal/telegram/keyboard_test.go
- [X] T029 [P] [US2] Add handler tests for settings:language:uk, settings:language:en, and settings:language:ru callbacks changing only the pressing user's language in internal/telegram/handler_language_test.go
- [X] T030 [P] [US2] Add handler tests proving two Telegram users in the same group receive different user-specific languages in internal/telegram/handler_language_test.go
- [X] T031 [P] [US2] Add tests proving public alerts and monthly reports use Ukrainian even when the pressing user selected English or Russian in internal/telegram/handler_language_test.go

### Implementation for User Story 2

- [X] T032 [US2] Implement SettingsKeyboard and LanguageKeyboard with localized labels and stable settings callback_data in internal/telegram/keyboard.go
- [X] T033 [US2] Add settings and language translation keys for uk/en/ru in internal/i18n/i18n.go
- [X] T034 [US2] Inject language lookup service into the Telegram handler constructors in internal/telegram/handler.go
- [X] T035 [US2] Implement per-user language resolution for user-specific message and callback flows in internal/telegram/handler.go
- [X] T036 [US2] Implement settings:language and settings:language:<code> callback handling with durable SetLanguage calls in internal/telegram/handler.go
- [X] T037 [US2] Log successful and failed language changes with event telegram.user_language_changed in internal/telegram/handler.go
- [X] T038 [US2] Keep StopAlertMessage and ReportMessage public output Ukrainian by using Ukrainian translations regardless of user preference in internal/telegram/messages.go

**Checkpoint**: User Story 2 is functional: language settings persist, are user-specific, and public group output remains Ukrainian.

---

## Phase 5: User Story 3 - Preserve Command Fallback And Message Cleanup (Priority: P3)

**Goal**: Existing text commands still work, and the bot attempts to delete only the
user's command message after processing.

**Independent Test**: Run every existing text command, confirm the command result is
unchanged, confirm deleteMessage is attempted for the user's command, and confirm bot
responses, reports, alerts, incidents, and monthly reports are not deleted.

### Tests for User Story 3

- [X] T039 [P] [US3] Add handler tests proving existing /add_person, /start_work, /status, /stop_work, and /report_month commands still return expected outcomes in internal/telegram/handler_command_fallback_test.go
- [X] T040 [P] [US3] Add handler tests proving deleteMessage is attempted after supported text command processing in internal/telegram/handler_delete_test.go
- [X] T041 [P] [US3] Add handler tests proving deleteMessage failure is logged and does not fail the command response in internal/telegram/handler_delete_test.go
- [X] T042 [P] [US3] Add handler tests proving bot responses, reports, alerts, incident messages, and monthly reports are never passed to DeleteMessage in internal/telegram/handler_delete_test.go
- [X] T043 [P] [US3] Add localized command response tests for saved user language lookup on text commands in internal/telegram/handler_command_fallback_test.go

### Implementation for User Story 3

- [X] T044 [US3] Extend Telegram handler client dependency with DeleteMessage support in internal/telegram/handler.go
- [X] T045 [US3] Track message_id, chat_id, user_id, and command name during text command webhook handling in internal/telegram/handler.go
- [X] T046 [US3] Attempt best-effort DeleteMessage after processing user text commands in internal/telegram/handler.go
- [X] T047 [US3] Log telegram.command_delete attempt, success, and failure outcomes in internal/telegram/handler.go
- [X] T048 [US3] Localize user-specific command responses and validation errors using the sender's saved language in internal/telegram/messages.go and internal/telegram/handler.go

**Checkpoint**: User Story 3 is functional: fallback commands remain usable and command cleanup is best-effort only.

---

## Phase 6: User Story 4 - Handle Button Callbacks Reliably (Priority: P4)

**Goal**: Every callback query is acknowledged and invalid or stale callback data receives
a localized error without affecting other users.

**Independent Test**: Press every supported button and send malformed callback data, then
confirm AnswerCallbackQuery is attempted each time and callback logs show received,
handled, and answer outcomes.

### Tests for User Story 4

- [X] T049 [P] [US4] Add tests that every supported callback action triggers AnswerCallbackQuery in internal/telegram/handler_callback_test.go
- [X] T050 [P] [US4] Add tests that unsupported, malformed, and stale callback data are acknowledged and return localized errors in internal/telegram/handler_callback_test.go
- [X] T051 [P] [US4] Add callback structured logging tests for telegram.callback_received, telegram.callback_result, and telegram.callback_answer in internal/telegram/handler_logging_test.go
- [X] T052 [P] [US4] Add tests that callback answer failure is logged but does not prevent a user-visible fallback response when possible in internal/telegram/handler_callback_test.go

### Implementation for User Story 4

- [X] T053 [US4] Extend Telegram handler client dependency with AnswerCallbackQuery support in internal/telegram/handler.go
- [X] T054 [US4] Ensure every callback query path attempts AnswerCallbackQuery exactly once in internal/telegram/handler.go
- [X] T055 [US4] Add unsupported and malformed callback handling with localized errors in internal/telegram/callbacks.go and internal/telegram/handler.go
- [X] T056 [US4] Log callback received, callback handled, and callback answer success/failure with safe fields in internal/telegram/handler.go

**Checkpoint**: User Story 4 is functional: callback UX is reliable and observable.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validate full feature behavior, documentation, and project constraints.

- [X] T057 [P] Update README command/help examples to mention inline buttons, Settings language selection, and Ukrainian default behavior in README.md
- [X] T058 [P] Update README operational notes for command deletion permission and best-effort behavior in README.md
- [X] T059 Run gofmt on all changed Go files in cmd/ and internal/
- [X] T060 Run go test ./... and fix any failing tests in cmd/ and internal/
- [X] T061 Validate specs/003-telegram-buttons-languages/quickstart.md scenarios against implemented behavior in Telegram or handler tests
- [X] T062 Verify no implementation adds web UI files, Telegram polling, Redis, queues, WebSockets, background infinite loops, or in-memory timers in cmd/, internal/, go.mod, and README.md
- [X] T063 Verify public alerts, incident messages, monthly reports, and report content remain Ukrainian in internal/telegram/messages.go and internal/reports/service.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational and is the MVP increment.
- **User Story 2 (Phase 4)**: Depends on Foundational; integrates naturally after US1 because Settings appears in the menu.
- **User Story 3 (Phase 5)**: Depends on Foundational; can run after language lookup is available for localized command responses.
- **User Story 4 (Phase 6)**: Depends on Foundational; can run alongside US1/US2 callback work if file conflicts are coordinated.
- **Polish (Phase 7)**: Depends on selected user stories.

### User Story Dependencies

- **US1**: Independent after Foundational for default Ukrainian menu/buttons.
- **US2**: Independent after Foundational, but most useful after US1 menu rendering exists.
- **US3**: Independent after Foundational for command fallback and cleanup; uses US2 language lookup for localized command responses.
- **US4**: Independent after Foundational for callback acknowledgement and invalid callback behavior; overlaps handler files with US1/US2.

### Within Each User Story

- Tests before implementation.
- Models and data contracts before services.
- Services before Telegram handler integration.
- Telegram client methods before handler calls.
- Story complete before moving to the next priority unless coordinating parallel work.

---

## Parallel Opportunities

- Setup placeholders T002-T004 can run in parallel.
- Foundational tests T005, T007, T012, and T014 can run in parallel.
- US1 tests T017-T020 can run in parallel.
- US2 tests T027-T031 can run in parallel.
- US3 tests T039-T043 can run in parallel.
- US4 tests T049-T052 can run in parallel.
- README updates T057-T058 can run in parallel.

## Parallel Example: User Story 1

```bash
Task: "T017 [P] [US1] Add keyboard rendering tests for Ukrainian main menu button labels and stable callback_data in internal/telegram/keyboard_test.go"
Task: "T018 [P] [US1] Add callback routing tests for menu:status and menu:help actions in internal/telegram/handler_callback_test.go"
Task: "T020 [P] [US1] Add localized help/menu response tests that include both button and text command options in internal/telegram/messages_test.go"
```

## Parallel Example: User Story 2

```bash
Task: "T027 [P] [US2] Add repository tests for user_settings upsert, created_at preservation, updated_at change, and telegram_user_id lookup in internal/users/repository_test.go"
Task: "T028 [P] [US2] Add settings menu rendering tests for language selection labels in internal/telegram/keyboard_test.go"
Task: "T030 [P] [US2] Add handler tests proving two Telegram users in the same group receive different user-specific languages in internal/telegram/handler_language_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T039 [P] [US3] Add handler tests proving existing commands still return expected outcomes in internal/telegram/handler_command_fallback_test.go"
Task: "T040 [P] [US3] Add handler tests proving deleteMessage is attempted after supported text command processing in internal/telegram/handler_delete_test.go"
Task: "T043 [P] [US3] Add localized command response tests for saved user language lookup on text commands in internal/telegram/handler_command_fallback_test.go"
```

## Parallel Example: User Story 4

```bash
Task: "T049 [P] [US4] Add tests that every supported callback action triggers AnswerCallbackQuery in internal/telegram/handler_callback_test.go"
Task: "T050 [P] [US4] Add tests that unsupported, malformed, and stale callback data are acknowledged and return localized errors in internal/telegram/handler_callback_test.go"
Task: "T051 [P] [US4] Add callback structured logging tests for telegram.callback_received, telegram.callback_result, and telegram.callback_answer in internal/telegram/handler_logging_test.go"
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete User Story 1 for default Ukrainian inline buttons and command-equivalent actions.
3. Stop and validate the main menu, Status, Help, and action prompts independently.

### Incremental Delivery

1. Add User Story 2 for per-user language persistence and Settings.
2. Add User Story 3 for command fallback cleanup and localized text command responses.
3. Add User Story 4 for complete callback acknowledgement, invalid callback handling, and callback logs.
4. Run quickstart validation and `go test ./...`.

### Notes

- Keep translation dictionaries in code; do not add translation files or third-party i18n dependencies.
- Do not ask for language automatically; only Settings may change language.
- Do not delete bot-generated messages of any type.
- Do not add group-level language settings in this feature.
- Do not add web UI, polling, Redis, queues, WebSockets, background infinite loops, or in-memory timers.
