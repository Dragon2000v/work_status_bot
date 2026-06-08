# Feature Specification: Telegram Team Work Status Tracker

**Feature Branch**: `001-team-status-bot`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "Build a Telegram-only team work status tracker. The system should work as a Telegram bot used inside a group chat. People can be added with first name and last name. Each person can have multiple work records. A work record has title, start date/time, optional stop date/time, status, and optional stop reason. The bot should support commands: /help, /add_person, /start_work, /status, /stop_work, /report_month. The system must not have a web UI in the MVP. The system must not rely on continuously running timers. It should store start and stop timestamps and calculate elapsed time only when a Telegram command is received or when a report is generated. When a work timer is stopped or reset, the bot must immediately send an alert message to the configured Telegram group. The monthly report should list all people, active work records, elapsed time, and incidents for the month. The system should avoid duplicate monthly reports for the same month. The project should be designed for free hosting where the server may sleep."

## Clarifications

### Session 2026-06-08

- Q: What MVP scope and operating assumptions apply before planning? -> A: No web UI,
  no authentication, people managed through Telegram commands, one person can have many
  work records, stopped records stay in history, stopping creates an incident, manual
  reports use `/report_month`, automatic monthly reports are triggered by an external
  scheduled HTTP request, Telegram webhook only.
- Q: How should interruption, alert failure, report overlap, cron security, and hard
  technical constraints work for MVP? -> A: MVP has no separate reset command; `/stop_work`
  is the only interruption action. A stop succeeds after database stop and incident save,
  then Telegram alert delivery is attempted. Alert failure is recorded and shown as a
  command warning. Monthly reports include records active at any moment in the selected
  month. Cron report requests require `X-Cron-Secret`. MVP must not use Redis, queues,
  WebSockets, in-memory timers, background infinite loops, Telegram polling, or web UI.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Track Work In Group Chat (Priority: P1)

A team member uses the Telegram group bot to add a person, start a work record, view
current status, and stop the work record when needed.

**Why this priority**: This is the core daily workflow. Without it, the system cannot
track team work status.

**Independent Test**: In a Telegram group, add a person, start work for that person,
check status, then stop the work record and verify the status changes.

**Acceptance Scenarios**:

1. **Given** the group has no person named "Ivan Petrenko", **When** a user runs
   `/add_person Ivan Petrenko`, **Then** the bot confirms that Ivan Petrenko was added.
2. **Given** Ivan Petrenko exists, **When** a user runs `/start_work Ivan Petrenko "API fix"`,
   **Then** the bot creates an active work record with the current start date/time.
3. **Given** Ivan Petrenko has active work, **When** a user runs `/status`,
   **Then** the bot lists Ivan Petrenko's active work, status, start date/time, and
   elapsed time calculated from the stored start date/time.
4. **Given** Ivan Petrenko has active work, **When** a user runs
   `/stop_work Ivan Petrenko "Blocked by dependency"`, **Then** the bot stops the active
   work record, stores the stop reason, and shows the updated stopped status.

---

### User Story 2 - Alert On Stop (Priority: P2)

Team members receive an immediate group alert attempt when an active work timer is
stopped.

**Why this priority**: Stop events are operationally important and must be visible to the
team without waiting for reports.

**Independent Test**: Stop an active work record and verify that the configured Telegram
group receives one alert message with the person, work title, event type, and reason when
available. If alert delivery fails, verify that the command response contains a visible
warning and an alert failure event is recorded.

**Acceptance Scenarios**:

1. **Given** a person has an active work record, **When** the work record is stopped,
   **Then** the bot saves the stopped record, creates an incident, and attempts to send a
   Telegram alert immediately.
2. **Given** Telegram alert delivery fails after a stop, **When** the command response is
   returned, **Then** the response includes a visible warning and an alert failure event
   is recorded.
3. **Given** a stop reason is provided, **When** the alert is sent, **Then** the alert
   includes the stop reason.

---

### User Story 3 - Generate Monthly Report (Priority: P3)

A team lead requests a monthly report in Telegram to review all people, active work,
elapsed time, and incidents for the selected month.

**Why this priority**: Reporting gives the team a monthly summary after core tracking is
available.

**Independent Test**: Create several records for a month, including active and stopped
records with incidents, then run `/report_month` and verify the report includes the
expected people, work records, elapsed time, and incidents exactly once for the month.

**Acceptance Scenarios**:

1. **Given** work records exist for the current month, **When** a user runs
   `/report_month`, **Then** the bot sends a report listing all people, active work
   records, elapsed time, and incidents for that month.
2. **Given** a monthly report has already been generated for a month, **When** a user
   runs `/report_month` for the same month again, **Then** the bot avoids creating a
   duplicate monthly report and returns the existing report or a clear duplicate notice.
3. **Given** the server was sleeping before the report command, **When** the command is
   received, **Then** elapsed time is still calculated from stored timestamps and the
   report remains accurate.
4. **Given** an external scheduler triggers the monthly report for a month, **When** a
   report already exists for that month, **Then** the bot avoids creating a duplicate
   report and sends or records a clear duplicate result.
5. **Given** work records overlap the selected month, **When** a report is generated,
   **Then** the report includes every record active at any moment during that month,
   including records started before the month and still active during it, records started
   during the month, and records stopped during the month.

### Edge Cases

- A user tries to add a person with the same first and last name as an existing person.
- A user starts work for a person who does not exist.
- A user starts another work record for a person who already has active work.
- A user stops work for a person with no active work.
- A user requests `/status` when no people or no active work records exist.
- A user requests `/report_month` for a month with no work records.
- The bot receives malformed commands or missing command arguments.
- The server sleeps between start and stop commands.
- A stop alert cannot be delivered to the configured group.
- A work record crosses month boundaries.
- An external monthly report trigger arrives for the same month after a manual report.
- An external monthly report trigger arrives while the hosting server is waking from sleep.
- A monthly report trigger omits or sends an invalid cron secret.
- A user attempts an unsupported interruption flow other than `/stop_work`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST operate as a Telegram bot used in a group chat.
- **FR-002**: The system MUST support `/help` and list available commands with concise
  usage examples.
- **FR-003**: The system MUST support `/add_person` with first name and last name.
- **FR-004**: The system MUST prevent duplicate people with the same first and last name
  within the tracked team.
- **FR-005**: The system MUST allow each person to have multiple work records.
- **FR-006**: Each work record MUST include a title, start date/time, status, optional
  stop date/time, and optional stop reason.
- **FR-007**: The system MUST support `/start_work` to create an active work record for
  an existing person.
- **FR-008**: The system MUST support `/status` to list all people and their current
  work status, including active elapsed time where applicable.
- **FR-009**: The system MUST support `/stop_work` to stop an active work record and
  record an optional stop reason.
- **FR-010**: When a work timer is stopped, the system MUST attempt to send a Telegram
  alert immediately. If alert delivery fails, the stop action remains successful when
  the work record was stopped in durable storage and the stopped incident was saved; the
  alert failure MUST be recorded and the command response MUST show a warning.
- **FR-011**: The MVP MUST NOT support a separate reset command; `/stop_work` is the only
  supported interruption action.
- **FR-012**: Stop alerts MUST include the person, work title, event type, event
  date/time, and reason when provided.
- **FR-013**: The system MUST support `/report_month` for the current month by default.
- **FR-014**: The system MUST allow `/report_month` to target a specific month using a
  clear month input.
- **FR-015**: The monthly report MUST list all people tracked by the bot.
- **FR-016**: The monthly report MUST list every work record that was active at any time
  during the selected month. This includes records started before the month and still
  active during it, records started during the month, and records stopped during the
  month.
- **FR-017**: The monthly report MUST include elapsed time for work records, calculated
  from stored timestamps at report generation time.
- **FR-018**: The monthly report MUST include incidents for the month, including stopped
  work events and alert failure events with reasons when available.
- **FR-019**: The system MUST avoid duplicate monthly reports for the same month.
- **FR-020**: The system MUST calculate elapsed time only when a Telegram command is
  received or when a report is generated.
- **FR-021**: The system MUST NOT rely on continuously running timers.
- **FR-022**: The system MUST remain accurate when the hosting server sleeps between
  commands.
- **FR-023**: The system MUST NOT include a web UI in the MVP.
- **FR-024**: The system MUST store timestamps in UTC and display dates in Europe/Kyiv
  timezone.
- **FR-025**: The system MUST use Telegram webhook handling for bot interactions.
- **FR-026**: The system MUST persist durable state in MongoDB Atlas.
- **FR-027**: The system MUST calculate elapsed time from stored timestamps.
- **FR-028**: The MVP MUST NOT require authentication or user roles for Telegram commands.
- **FR-029**: People MUST be managed through Telegram commands only in the MVP.
- **FR-030**: Stopping a work record MUST create an incident and retain the stopped work
  record in history.
- **FR-031**: The system MUST support automatic monthly report generation when triggered
  by an external scheduled HTTP request.
- **FR-032**: The system MUST NOT include an internal scheduler for automatic monthly
  reports.
- **FR-033**: Manual and externally triggered monthly reports MUST share the same
  duplicate prevention rule for each month.
- **FR-034**: A stop action MUST be considered successful after the work record is stopped
  in durable storage and a stopped incident is saved.
- **FR-035**: After a successful stop action, the system MUST attempt to send a Telegram
  alert to the configured group.
- **FR-036**: If Telegram alert delivery fails, the system MUST record an `alert_failed`
  event or equivalent alert failure state.
- **FR-037**: If Telegram alert delivery fails, the bot MUST return a visible warning in
  the `/stop_work` command response.
- **FR-038**: A monthly report MUST include every work record that was active at any
  moment during the selected month, including records started before the month and still
  active during it, records started during the month, and records stopped during the month.
- **FR-039**: `POST /cron/monthly-report` MUST require an `X-Cron-Secret` header matching
  the configured cron secret.
- **FR-040**: `POST /cron/monthly-report` MUST return `401 Unauthorized` when
  `X-Cron-Secret` is missing or invalid.
- **FR-041**: The MVP MUST NOT use Redis, queues, WebSockets, in-memory timers,
  background infinite loops, Telegram polling, or a web UI.

### Key Entities *(include if feature involves data)*

- **Person**: A tracked team member, identified by first name and last name. A person can
  have multiple work records.
- **Work Record**: A record of work for a person. Includes title, start date/time, status,
  optional stop date/time, and optional stop reason.
- **Incident**: A notable work event for reporting and alerts, such as stopped work or
  alert failure, including event date/time and reason when available.
- **Monthly Report**: A generated report for a specific month. Tracks the month, report
  content, and whether that month has already had a report generated.
- **Telegram Group Configuration**: The configured group where bot commands are received
  and stop alerts are sent.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can add a person, start work, check status, and stop work from the
  Telegram group in under 2 minutes.
- **SC-002**: Status output reflects active elapsed time within 1 minute of the expected
  value when calculated at command time.
- **SC-003**: Every successful stop action attempts Telegram alert delivery and produces
  either a visible group alert or a visible command warning with a recorded alert failure.
- **SC-004**: Monthly reports include all tracked people and all work records active at
  any moment during the selected month, plus relevant incidents for the selected month in
  test data.
- **SC-005**: Re-running the monthly report for the same month does not create duplicate
  report records.
- **SC-006**: After the service has been asleep for at least 30 minutes, status and report
  elapsed times still match calculations from stored start and stop timestamps.
- **SC-007**: 90% of common command mistakes return a clear correction message instead of
  failing silently.

## Assumptions

- The Telegram group used for commands is also the configured alert group for MVP.
- All MVP Telegram commands are accepted from the configured Telegram group without
  authentication or role checks.
- A duplicate person means the same first name and last name within the same tracked team.
- A person can have only one active work record at a time, while retaining any number of
  stopped historical work records.
- Stopping a timer creates an incident.
- MVP does not implement a separate reset command. Any interruption is handled by
  `/stop_work`.
- `/report_month` without arguments targets the current month in Europe/Kyiv timezone.
- Specific month input uses `YYYY-MM` format.
- If a monthly report already exists, the bot returns the existing report or a duplicate
  notice rather than generating another stored report.
- Automatic monthly reports are initiated by an external scheduler calling the system;
  the bot does not schedule reports with internal timers.
- External monthly report calls include `X-Cron-Secret` matching the configured cron
  secret.
- Failed alert delivery is recorded as an incident or error visible to later operators.
