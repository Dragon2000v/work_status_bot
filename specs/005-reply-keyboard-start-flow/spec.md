# Feature Specification: Telegram Reply Keyboard, Languages, And Start Work Flow

**Feature Branch**: `005-reply-keyboard-start-flow`

**Created**: 2026-06-09

**Status**: Draft

**Input**: User description: "Add Telegram reply keyboard menu, native bot commands menu, per-user language preferences, and a button-driven start work flow to the existing Go Telegram work status bot. Replace the main menu inline buttons with Telegram ReplyKeyboardMarkup persistent buttons, use Ukrainian by default, support Ukrainian, English, and Russian, remember each user's selected language by Telegram user id, register native Telegram bot commands, support both slash commands and reply keyboard button text, delete user messages best-effort after commands, button messages, and manual flow input, and add a button-driven Start Work flow with localized date/time parsing in Europe/Kyiv while storing UTC timestamps. Keep existing webhook behavior, storage, multi-group setup, group authorization, structured logging, tests, reports, alerts, cron behavior, and banned runtime constraints."

## Clarifications

### Session 2026-06-09

- Q: Which Telegram UI should serve as the main menu? → A: ReplyKeyboardMarkup is the main menu mechanism; InlineKeyboardMarkup must not be used for the main menu anymore.
- Q: When should the native Telegram command menu be registered? → A: Register the native Telegram bot command menu with setMyCommands on startup.
- Q: How should localization work when Telegram group reply keyboards cannot be reliably personalized per user? → A: Use Ukrainian for group-wide keyboards and the selected user language for personal responses, prompts, and settings.
- Q: How should multi-step flow state be stored and cleaned up? → A: Store multi-step flow state in MongoDB and expire flow states to avoid stale state.
- Q: Which runtime mechanisms remain out of scope? → A: Do not introduce web UI, polling, Redis, queues, WebSockets, in-memory timers, or background infinite loops.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Use Persistent Reply Keyboard Menu (Priority: P1)

A group user can use large persistent Telegram menu buttons near the message input field
to run the bot's primary actions without typing or opening an inline button menu.

**Why this priority**: The reply keyboard is the requested primary interface and makes
routine group work-status actions easier to discover and repeat.

**Independent Test**: In an authorized group, open the bot and confirm the main menu
appears as persistent reply keyboard buttons. Press each main button and confirm it
starts or completes the same user-visible action as the matching command.

**Acceptance Scenarios**:

1. **Given** a user has no saved language preference or the bot is showing a group-wide
   keyboard, **When** the bot shows the main menu, **Then** the reply keyboard labels are
   Ukrainian and include Додати людину, Почати роботу, Статус, Зупинити роботу,
   Місячний звіт, Налаштування, and Допомога.
2. **Given** the main reply keyboard is visible, **When** a user presses Статус,
   **Then** the bot handles the resulting text message as the status action.
3. **Given** the main reply keyboard is visible, **When** a user presses Допомога,
   **Then** the bot returns localized help text that explains available button and
   slash-command actions.
4. **Given** the main reply keyboard is visible, **When** a user presses Додати людину,
   **Then** the bot starts a multi-step add-person flow and stores
   `add_person_enter_name` state for that user and chat.
5. **Given** the main reply keyboard is visible, **When** a user presses Зупинити роботу,
   **Then** the bot starts a multi-step stop-work flow and stores
   `stop_work_enter_person_or_reason` state for that user and chat.
6. **Given** the prior main menu used inline buttons, **When** the new main menu is
   shown, **Then** InlineKeyboardMarkup is not used for main-menu navigation.

---

### User Story 2 - Use Native Telegram Command Menu And Slash Fallback (Priority: P1)

A user can still run all existing slash-command flows and can discover those commands
through Telegram's native bot command menu.

**Why this priority**: Existing users and operational workflows rely on commands, while
the native command menu improves discoverability without replacing slash fallback.

**Independent Test**: Open Telegram's bot command menu and verify the expected commands
are listed. Run each command manually and confirm the command still performs the same
user-visible action as before.

**Acceptance Scenarios**:

1. **Given** the bot starts, **When** startup registration completes, **Then** Telegram's
   native command menu is registered and lists /setup, /groups, /disable_group,
   /add_person, /start_work, /status, /stop_work, /report_month, and /help.
2. **Given** a user sends /status, **When** the bot processes the message, **Then** the
   status output is returned according to existing group authorization rules.
3. **Given** a user presses the localized status keyboard button, **When** the bot
   receives the resulting text message, **Then** it performs the same status action as
   /status.

---

### User Story 3 - Remember Per-User Language Preference (Priority: P2)

A user can choose Ukrainian, English, or Russian in Settings, and the bot remembers that
choice for future personal interactions with that Telegram user id.

**Why this priority**: Personal localization must persist without interrupting every
interaction, while public group output remains predictable.

**Independent Test**: One user changes language in Settings, another user keeps the
default language, the service restarts, and both users perform personal interactions.
Each user should continue seeing their own language where personal localization applies.

**Acceptance Scenarios**:

1. **Given** a user has never selected a language, **When** they interact with personal
   bot flows, **Then** the bot uses Ukrainian by default.
2. **Given** a user opens Settings and chooses to change language, **When** language
   options are shown, **Then** Ukrainian, English, and Russian are available.
3. **Given** a user selected English earlier, **When** they later use Help, Settings, or
   the Start Work flow, **Then** personal prompts, button labels, and validation errors
   use English.
4. **Given** users in the same group have different saved languages, **When** each user
   performs a personal interaction, **Then** each interaction uses the initiating user's
   saved language.
5. **Given** users have saved personal languages, **When** a public alert, incident
   message, status output, or monthly report is sent to the group, **Then** that public
   group output uses Ukrainian by default.
6. **Given** Telegram group reply keyboard localization is limited, **When** the bot
   displays a group-wide keyboard, **Then** the keyboard remains Ukrainian while personal
   responses and settings still use the initiating user's saved language when available.

---

### User Story 4 - Start Work From Buttons (Priority: P2)

A user can start work from the reply keyboard without typing the slash command, enter
the person and work title through prompts, and choose a start date/time mode.

**Why this priority**: Starting work is a primary workflow, and the preferred experience
should be button-driven rather than command-driven.

**Independent Test**: Press Почати роботу, complete the prompted flow with "start now",
today with time, yesterday with time, and manual date/time values, then verify the work
record appears in status with the expected person, title, and displayed Kyiv time.

**Acceptance Scenarios**:

1. **Given** a user presses Почати роботу, **When** the bot starts the flow, **Then** it
   asks for the person's first name and last name, or offers a simple existing-person
   selection if that is available in the current interaction.
2. **Given** the user provides a valid person, **When** the bot continues, **Then** it
   asks for the work title.
3. **Given** the user provides a work title, **When** the bot asks how to set start
   date/time, **Then** the user can choose Почати зараз, Сьогодні, Вчора, Ввести дату
   вручну, or Скасувати.
4. **Given** the user chooses Почати зараз, **When** the flow completes, **Then** the
   work start time is the current time of the interaction.
5. **Given** the user chooses Сьогодні or Вчора, **When** the bot asks for time and the
   user enters HH:mm, **Then** the selected local date and entered local time are used.
6. **Given** the user chooses Ввести дату вручну, **When** the user enters DD.MM HH:mm,
   DD.MM.YYYY HH:mm, or YYYY-MM-DD HH:mm, **Then** the bot parses the value as
   Europe/Kyiv local time and stores the resulting instant in UTC.
7. **Given** the user enters a manual date without a year, **When** the bot parses it,
   **Then** it uses the current year in Europe/Kyiv.
8. **Given** the user selects Скасувати at any start-work date/time choice point,
   **When** cancellation is confirmed, **Then** no work record is started.

---

### User Story 5 - Clean Up User Messages Without Hiding Bot Output (Priority: P3)

After handling user commands, keyboard button text, and manual flow input, the bot
attempts to delete the user's message while leaving all bot-generated outputs visible.

**Why this priority**: Group chats stay readable, but reports, alerts, and bot responses
must remain visible for the team.

**Independent Test**: Run command, reply-keyboard, and manual flow messages in a chat
where the bot can delete messages and in a chat where it cannot. Confirm only user
messages are targeted, failures are logged, and workflows continue normally.

**Acceptance Scenarios**:

1. **Given** a user sends a slash command, **When** the bot processes the command, **Then**
   it attempts to delete only that user's command message.
2. **Given** a user presses a reply keyboard button, **When** the bot handles the
   resulting text message, **Then** it attempts to delete only that user's button text
   message.
3. **Given** a user sends manual input inside an active flow, **When** the bot processes
   the input, **Then** it attempts to delete only that user's input message.
4. **Given** deletion fails because the bot lacks permissions or the message cannot be
   deleted, **When** the workflow continues, **Then** the failure is logged and normal
   command or flow handling is not blocked.
5. **Given** the bot sends a response, alert, report, incident message, monthly report,
   or status output, **When** cleanup is attempted, **Then** those bot-generated messages
   are not deleted.

### Edge Cases

- A user sends a slash command with a bot username suffix, such as /status@botname.
- A user presses a localized reply keyboard button after changing their language.
- A user manually types text that exactly matches a button label.
- A user enters an unknown button label or unsupported command.
- A user starts the Add Person flow and sends a name without both first name and last
  name.
- A user starts the Add Person flow for a duplicate person.
- A user starts the Stop Work flow for an unknown person.
- A user starts the Stop Work flow for a person with no active work.
- A user starts the Start Work flow and sends unrelated text at the person, title, time,
  or manual-date step.
- A user enters a date/time in the future.
- A user enters an impossible local date or invalid time, such as 31.02 09:00 or 25:99.
- A user enters DD.MM HH:mm near a year boundary.
- The service restarts after a language preference has been saved.
- The service restarts while a user was in a multi-step interaction flow.
- A stored multi-step flow state expires before the user's next text message arrives.
- The bot receives updates from an unauthorized group.
- Command registration fails or is unavailable during startup or setup.
- Message deletion fails because of permissions, age limits, missing messages, or
  Telegram-side errors.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST use Ukrainian as the default language for personal
  interactions and public group output when no more specific saved personal language
  applies.
- **FR-002**: The system MUST support Ukrainian, English, and Russian for personal
  interactions.
- **FR-003**: The system MUST remember each user's selected language by Telegram user id.
- **FR-004**: The system MUST reuse a saved user language for future personal
  interactions with that user and MUST NOT ask for language on every interaction.
- **FR-005**: Language selection MUST appear only when a user opens Settings and chooses
  to change language.
- **FR-006**: Public group alerts, incident messages, status output, and monthly reports
  MUST use Ukrainian by default.
- **FR-007**: If Telegram reply keyboard localization per user is limited in group
  chats, group-wide keyboards MUST use Ukrainian while personal responses, prompts, and
  settings messages use the user's selected language when available.
- **FR-008**: The main menu MUST be presented as a Telegram ReplyKeyboardMarkup with
  large persistent buttons near the message input field.
- **FR-009**: The default Ukrainian main reply keyboard MUST include Додати людину,
  Почати роботу, Статус, Зупинити роботу, Місячний звіт, Налаштування, and Допомога.
- **FR-010**: The main menu MUST replace the prior inline-button main menu for primary
  navigation and MUST NOT use InlineKeyboardMarkup for main-menu actions.
- **FR-011**: The system MUST continue supporting slash commands as fallback actions.
- **FR-012**: The system MUST handle reply keyboard button text messages as actions
  equivalent to their matching slash commands.
- **FR-013**: The system MUST register the native Telegram bot command menu on startup
  with /setup, /groups, /disable_group, /add_person, /start_work, /status, /stop_work,
  /report_month, and /help.
- **FR-014**: The system MUST preserve existing setup, group listing, group disabling,
  add-person, start-work, status, stop-work, monthly report, and help command behavior
  unless this feature explicitly changes the interaction surface.
- **FR-015**: The system MUST apply existing multi-group setup and group authorization
  rules to command, reply-keyboard, and flow interactions.
- **FR-016**: After processing a user slash command, reply keyboard button message, or
  manual input inside a flow, the system MUST attempt to delete that user's message from
  the chat.
- **FR-017**: User message deletion MUST be best-effort: deletion failures MUST be logged
  and MUST NOT prevent normal response, alert, report, or flow behavior.
- **FR-018**: The system MUST NOT delete bot responses, alerts, reports, incident
  messages, status output, or monthly reports.
- **FR-019**: The Start Work flow MUST be available from the reply keyboard and from the
  existing slash command fallback.
- **FR-020**: The Start Work flow MUST ask for the person's first name and last name, or
  allow selecting an existing person if that can be presented simply in the same flow.
- **FR-021**: The Start Work flow MUST ask for the work title after the person is known.
- **FR-022**: The Start Work flow MUST let the user choose Почати зараз, Сьогодні,
  Вчора, Ввести дату вручну, or Скасувати when selecting the start date/time.
- **FR-023**: For Сьогодні and Вчора, the system MUST ask for a time in HH:mm format.
- **FR-024**: For manual date/time entry, the system MUST accept DD.MM HH:mm,
  DD.MM.YYYY HH:mm, and YYYY-MM-DD HH:mm.
- **FR-025**: Date/time values entered by users MUST be interpreted in Europe/Kyiv local
  time.
- **FR-026**: If a date/time value omits the year, the system MUST use the current year
  in Europe/Kyiv.
- **FR-027**: The system MUST store accepted start timestamps in UTC.
- **FR-028**: The system MUST reject future start date/time values.
- **FR-029**: If date/time input is invalid, impossible, unsupported, or in the future,
  the system MUST respond with a localized error that includes accepted examples.
- **FR-030**: Multi-step flow state MUST be persisted durably so the system can identify
  what the user's next text message means after a prompt.
- **FR-031**: Multi-step flow states MUST expire to avoid stale state, and expired flows
  MUST require the user to restart the relevant action.
- **FR-032**: The system MUST calculate elapsed time from stored timestamps.
- **FR-033**: The system MUST display user-facing dates and times in Europe/Kyiv.
- **FR-034**: The system MUST keep existing webhook behavior.
- **FR-035**: The system MUST persist durable state in MongoDB Atlas.
- **FR-036**: The system MUST keep structured logging for command registration, button
  routing, language changes, flow progress, date/time parse failures, user message
  deletion attempts, and deletion failures.
- **FR-037**: The system MUST keep existing reports, alerts, cron behavior, and incident
  behavior unless explicitly changed by this feature.
- **FR-038**: The system MUST keep existing automated coverage and add coverage for
  reply keyboard actions, native command menu registration, language persistence,
  localized routing, Start Work date/time choices, future-date rejection, and
  best-effort message deletion.
- **FR-039**: The system MUST NOT add a web UI for MVP scope.
- **FR-040**: The system MUST NOT use Telegram polling, Redis, queues, WebSockets,
  in-memory timers, or background infinite loops.
- **FR-041**: Pressing Додати людину MUST start a multi-step add-person flow that asks
  for first name and last name and stores `add_person_enter_name` state.
- **FR-042**: The Add Person flow MUST validate `<first_name> <last_name>`, create the
  person through existing people behavior, send a localized success message, attempt to
  delete the user's manual input message, and clear flow state after success.
- **FR-043**: The Add Person flow MUST return localized validation errors with examples
  for missing first name or last name, preserve existing duplicate-person behavior, and
  support Скасувати cancellation.
- **FR-044**: Pressing Зупинити роботу MUST start a multi-step stop-work flow that asks
  for first name and last name plus optional reason and stores
  `stop_work_enter_person_or_reason` state.
- **FR-045**: The Stop Work flow MUST validate user input, stop active work through
  existing work behavior, create the existing incident/event, attempt to send the public
  alert to the current group, send a localized result message, attempt to delete the
  user's manual input message, and clear flow state after success.
- **FR-046**: If Stop Work alert sending fails after work is stopped and the
  incident/event is saved, the stop MUST remain successful and alert failure MUST be
  recorded according to existing behavior.
- **FR-047**: The Stop Work flow MUST return localized errors for unknown person, no
  active work, and already-stopped work where applicable, and MUST support Скасувати
  cancellation.
- **FR-048**: Flow states MUST expire 15 minutes after `updated_at`; every flow state read
  MUST treat expired state as invalid, clear the stale state, and return a localized
  expired-flow message.
- **FR-049**: Flow state updates MUST refresh both `updated_at` and `expires_at`.
- **FR-050**: If native Telegram command menu registration fails on startup, application
  startup MUST continue, the failure MUST be logged without secrets, and the bot MUST
  still handle slash commands and reply keyboard actions.
- **FR-051**: Public group incident/alert messages MUST use Ukrainian by default and
  monthly reports MUST continue including incidents correctly.

### Key Entities *(include if feature involves data)*

- **User Language Preference**: A durable personal setting that links a Telegram user id
  to Ukrainian, English, or Russian, with selection metadata.
- **Reply Keyboard Action**: A text message produced by pressing a persistent Telegram
  reply keyboard button and mapped to a bot action.
- **Native Bot Command**: A slash command exposed through Telegram's command menu and
  mapped to existing bot behavior.
- **Start Work Flow State**: A durable, expiring record of a user's current step in the
  Start Work interaction, including collected person, title, selected date/time mode,
  pending input, and expiration information.
- **Parsed Start Timestamp**: A user-entered or selected local date/time converted to a
  UTC instant for storage.
- **User Message Cleanup Attempt**: A best-effort attempt to delete a user-sent command,
  keyboard text message, or flow input, including outcome and failure reason when known.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of primary actions are available from the persistent reply keyboard
  and from slash-command fallback in tested authorized groups.
- **SC-002**: A new user can find and run status, help, and start-work actions from
  visible Telegram controls without typing a slash command.
- **SC-003**: The native Telegram command menu lists all required commands in 100% of
  verified startup checks.
- **SC-004**: A user's saved language is reused for personal interactions in 100% of
  tested interactions after service restart.
- **SC-005**: Public group alerts, incident messages, status output, and monthly reports
  remain Ukrainian in 100% of tested cases regardless of personal language preferences.
- **SC-006**: Users can complete the button-driven Start Work flow with "start now" in
  under 1 minute during manual acceptance testing.
- **SC-007**: Valid custom date/time examples in all accepted formats are parsed
  correctly in Europe/Kyiv and stored as the correct UTC instant in 100% of test cases.
- **SC-008**: Future date/time values and invalid date/time formats are rejected in 100%
  of tested cases with localized examples.
- **SC-009**: For command, reply-keyboard, and manual flow messages, the bot attempts
  user-message cleanup in 100% of tested applicable cases and never deletes bot outputs.
- **SC-010**: Existing webhook, group authorization, report, alert, cron, and command
  fallback behavior passes the existing regression test suite after this feature.
- **SC-011**: Expired multi-step flow states are rejected in 100% of tested stale-flow
  cases and prompt the user to restart the action.
- **SC-012**: Users can complete Add Person and Stop Work button flows without slash
  commands in 100% of acceptance test runs.
- **SC-013**: Stop Work records incident/event data and attempts the current-group alert
  in 100% of tested successful stop-flow runs, including alert-failure recording when
  delivery fails.

## Assumptions

- Ukrainian remains the default for public group communication even when individual users
  choose English or Russian for personal flows.
- Personal interactions include Settings, Help requested by a user, Start Work prompts,
  validation errors, and confirmations directed at the initiating user.
- Public group output includes alerts, incident messages, status output, and monthly
  reports intended for the group rather than one user's settings flow.
- Reply keyboard labels are localized for the initiating user's personal menu where that
  is reliable; if Telegram group reply keyboard localization is limited, group-wide
  keyboards use Ukrainian and personal responses/settings use the selected user language.
- Existing-person selection in Start Work is optional only when it can stay simple; manual
  first-name and last-name entry is always acceptable.
- Multi-step flow state is durable but expires 15 minutes after `updated_at`; after
  expiration, the bot clears the stale state and requires the user to restart that flow
  while saved language preferences and completed work records remain durable.
