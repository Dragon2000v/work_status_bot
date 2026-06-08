# Feature Specification: Telegram Buttons And User Languages

**Feature Branch**: `003-telegram-buttons-languages`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "Add interactive Telegram buttons and per-user language preferences to the existing Go Telegram work status bot. The bot should use Ukrainian as the default language for all messages, buttons, errors, reports, alerts, and help text. The bot should support three languages: Ukrainian, English, and Russian. The bot must remember each user's selected language by Telegram user id. If a user selected a language before, the bot should use that language for future interactions with that user. The bot must not ask for language on every interaction. Language selection should only appear when the user opens Settings and chooses to change language. The bot should add an inline button interface with actions: Add person, Start work, Status, Stop work, Monthly report, Settings, Help. Users should be able to use inline buttons instead of typing commands. The bot should support Telegram callback_query updates. The bot must call answerCallbackQuery for every callback query. Text commands should remain available as fallback. When a user sends a text command, the bot should attempt to delete the user's command message after processing it. Command deletion is best-effort. If deletion fails because the bot has no admin permissions, the bot should log the failure and continue normally. The bot must not delete bot responses, reports, alerts, incident messages, or monthly reports. For user-specific interaction flows, such as Settings and language selection, the bot should use the language selected by the user who pressed the button. For public group reports and alerts, use Ukrainian by default unless a future group-level language setting is added. Do not implement group-level language settings in this feature. The implementation must keep existing webhook behavior, MongoDB storage, structured logging, tests, and fallback command functionality."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Use Bot With Inline Buttons (Priority: P1)

A group user can open the bot's inline button menu and perform the main work-status
actions without typing commands.

**Why this priority**: Buttons reduce command memorization and make the existing bot
usable for routine work in a group chat.

**Independent Test**: In the Telegram group, open the bot menu, press each main action
button, and confirm the bot starts or completes the same user-visible action that the
matching text command provides.

**Acceptance Scenarios**:

1. **Given** the bot is available in the configured group, **When** a user opens the
   button interface, **Then** the bot shows Ukrainian labels by default for Add person,
   Start work, Status, Stop work, Monthly report, Settings, and Help.
2. **Given** the inline button interface is visible, **When** a user presses Status,
   **Then** the bot returns the current status output without requiring a typed command.
3. **Given** the inline button interface is visible, **When** a user presses Help,
   **Then** the bot returns localized help text that includes both button and text
   command options.
4. **Given** a button action requires additional information, **When** a user presses
   that action, **Then** the bot starts a clear localized interaction flow to collect the
   required input.

---

### User Story 2 - Remember Each User's Language (Priority: P2)

A user can choose Ukrainian, English, or Russian in Settings, and the bot remembers that
choice for future user-specific interactions.

**Why this priority**: Language preference is personal and must not interrupt every bot
interaction with repeated language prompts.

**Independent Test**: A user changes language in Settings, performs several user-specific
button and command interactions, leaves and returns later, and confirms those
interactions use the selected language without asking again.

**Acceptance Scenarios**:

1. **Given** a user has never selected a language, **When** they interact with the bot,
   **Then** all user-specific messages, buttons, errors, and help text shown to that user
   use Ukrainian.
2. **Given** a user opens Settings, **When** they choose Change language, **Then** the bot
   offers Ukrainian, English, and Russian as language choices.
3. **Given** a user selects English, **When** they later press Settings, Help, or a
   user-specific action button, **Then** the bot uses English for that user's
   interaction.
4. **Given** one user selected Russian and another user selected English, **When** each
   presses the same user-specific button, **Then** each user sees the response in their
   own selected language.

---

### User Story 3 - Preserve Command Fallback And Message Cleanup (Priority: P3)

A user can still type existing commands, while the bot attempts to remove only the
user's command message after processing.

**Why this priority**: Existing users and automation may rely on commands, and command
cleanup keeps group chats readable without removing important bot outputs.

**Independent Test**: Run every existing command as text, confirm the command still works,
confirm the bot attempts to delete the user's command message, and confirm all bot
responses, reports, alerts, incidents, and monthly reports remain visible.

**Acceptance Scenarios**:

1. **Given** a user sends a supported text command, **When** the bot finishes processing
   it, **Then** the command behavior remains available and the bot attempts to delete only
   that user's command message.
2. **Given** command deletion fails because the bot lacks permissions, **When** the
   command response is complete, **Then** the bot logs the deletion failure and the user
   workflow still succeeds or fails according to the command result.
3. **Given** the bot sends a response, report, alert, incident message, or monthly report,
   **When** command cleanup occurs, **Then** those bot-generated messages remain visible.

---

### User Story 4 - Handle Button Callbacks Reliably (Priority: P4)

A user pressing an inline button receives prompt Telegram feedback and never leaves the
button press in a loading state.

**Why this priority**: Callback acknowledgement is required for a polished Telegram user
experience and prevents repeated presses caused by uncertainty.

**Independent Test**: Press every inline button, including invalid or expired button
payloads, and confirm each press is acknowledged while the bot returns the correct
localized result or error.

**Acceptance Scenarios**:

1. **Given** a user presses any supported inline button, **When** the bot receives the
   button update, **Then** the button press is acknowledged and the requested action is
   handled.
2. **Given** the bot receives an unsupported, malformed, or stale button action, **When**
   it handles the update, **Then** the button press is acknowledged and the user receives
   a localized error without disrupting other users.

### Edge Cases

- A user presses a button before selecting a language.
- A user changes language while another user in the same group uses a different language.
- A user presses a stale or unsupported inline button.
- A button action needs additional input, but the user sends unrelated text.
- A command deletion attempt fails because the bot has no admin permissions.
- A command deletion attempt fails because the message is already gone or cannot be deleted.
- The bot receives a text command from a user who has a saved language preference.
- The bot sends public group reports or alerts after users have selected different languages.
- The service restarts after user language preferences have been saved.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST use Ukrainian as the default language for messages,
  buttons, errors, reports, alerts, and help text.
- **FR-002**: The system MUST support Ukrainian, English, and Russian for user-specific
  bot interactions.
- **FR-003**: The system MUST remember each user's selected language by Telegram user id.
- **FR-004**: If a user has a saved language selection, the system MUST use that language
  for future user-specific interactions with that user.
- **FR-005**: The system MUST NOT ask users to choose a language during every
  interaction.
- **FR-006**: The system MUST show language selection only when a user opens Settings and
  chooses to change language.
- **FR-007**: Settings and language selection flows MUST use the language selected by the
  user who pressed the button, falling back to Ukrainian if that user has no saved
  preference.
- **FR-008**: Public group reports and alerts MUST use Ukrainian by default.
- **FR-009**: The system MUST NOT add group-level language settings in this feature.
- **FR-010**: The system MUST provide an inline button interface for Add person, Start
  work, Status, Stop work, Monthly report, Settings, and Help.
- **FR-011**: Users MUST be able to perform supported actions through inline buttons
  instead of typing commands.
- **FR-012**: Button-triggered actions MUST preserve the same business outcomes and
  validation rules as their corresponding text command flows.
- **FR-013**: The system MUST support Telegram callback query updates for inline button
  actions.
- **FR-014**: The system MUST acknowledge every callback query, including successful,
  invalid, unsupported, and failed button actions.
- **FR-015**: Text commands MUST remain available as fallback for all existing supported
  command functionality.
- **FR-016**: When a user sends a text command, the system MUST attempt to delete that
  user's command message after processing.
- **FR-017**: Command message deletion MUST be best-effort: deletion failure MUST be
  logged and MUST NOT prevent normal command completion.
- **FR-018**: The system MUST NOT delete bot responses, reports, alerts, incident
  messages, or monthly reports.
- **FR-019**: The system MUST localize all user-facing button labels, settings text,
  command help, validation errors, and user-specific action prompts for each supported
  language.
- **FR-020**: The system MUST keep existing Telegram webhook behavior.
- **FR-021**: The system MUST persist durable state in MongoDB Atlas.
- **FR-022**: The system MUST keep structured logging for button handling, language
  preference changes, command deletion attempts, and deletion failures.
- **FR-023**: The system MUST preserve existing tests and add coverage for language
  preferences, button actions, callback acknowledgements, command fallback, and command
  deletion behavior.
- **FR-024**: The system MUST calculate elapsed time from stored timestamps.
- **FR-025**: The system MUST store timestamps in UTC and display dates in Europe/Kyiv
  timezone.
- **FR-026**: The system MUST NOT add a web UI for MVP scope.
- **FR-027**: The system MUST NOT use Telegram polling, in-memory timers, or background
  infinite loops.

### Key Entities *(include if feature involves data)*

- **User Language Preference**: A durable preference linking a Telegram user id to one of
  the supported languages and the date/time it was selected.
- **Button Action**: A user-facing action started from an inline button, such as status,
  help, settings, or a work-status operation.
- **Interaction Flow**: A short user-specific conversation started by a button or command
  when additional input is needed.
- **Command Cleanup Attempt**: A best-effort attempt to remove a user's typed command
  message, including whether it succeeded or failed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can reach each primary action through inline buttons in no more than
  2 button presses from the main menu.
- **SC-002**: 100% of tested inline button presses are acknowledged, including invalid
  and unsupported button actions.
- **SC-003**: A user's selected language is reused for future user-specific interactions
  in 100% of tested interactions after service restart.
- **SC-004**: New users who have not selected a language see Ukrainian user-facing text
  in 100% of tested messages, buttons, errors, reports, alerts, and help text.
- **SC-005**: Public group reports and alerts remain Ukrainian in 100% of tested cases,
  regardless of individual user language preferences.
- **SC-006**: 100% of existing text command fallback flows continue to complete with the
  same user-visible business outcome as before this feature.
- **SC-007**: For supported text commands, the bot attempts command-message cleanup in
  100% of tested command flows and logs deletion failures without failing the command.
- **SC-008**: Automated tests cover successful language changes, default language
  fallback, per-user language differences, button action routing, callback
  acknowledgements, and best-effort command deletion.

## Assumptions

- Ukrainian is the product default for both user-specific and public group output unless
  a saved user preference applies to a user-specific interaction.
- User-specific interactions include Settings, language selection, Help opened by a
  user, validation prompts, and action prompts caused by that user's button press or text
  command.
- Public group output includes reports, alerts, incident messages, and monthly reports
  intended for the whole group rather than a single user's settings flow.
- Existing text commands and button actions share the same underlying business rules and
  produce equivalent results, even if their interaction steps differ.
- Command cleanup applies only to user-sent command messages, not normal non-command text.
- Language preferences must survive service restarts.
- Future group-level language settings are intentionally out of scope for this feature.
