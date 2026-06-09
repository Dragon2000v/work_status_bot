# Feature Specification: Multi-Group Setup

**Feature Branch**: `004-multi-group-setup`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "Add multi-group setup support to the existing Go Telegram work status bot. Currently the bot uses TELEGRAM_GROUP_CHAT_ID from environment variables and works with one configured Telegram group. This should be improved. The bot should support a /setup command that allows a Telegram group to register itself as an allowed group. When the bot is added to a new group, a group admin should be able to run /setup. The bot should save the current Telegram chat id and group title in MongoDB. After setup, the bot should accept commands from that group without requiring TELEGRAM_GROUP_CHAT_ID to be changed or the app to be redeployed. The bot should support multiple groups. The bot should continue to support TELEGRAM_GROUP_CHAT_ID as a fallback for local development and backward compatibility. The bot should reject commands from groups that are not configured, except for /setup and /help. The /setup command should be idempotent. If the group is already configured, the bot should update the group title and respond that setup is already completed. The bot should store configured groups in MongoDB. Each configured group should have: Telegram chat id, group title, setup user id, setup username if available, enabled status, created at, updated at. The bot should add commands: /setup, /groups, /disable_group. For MVP: /setup registers the current group. /groups lists configured groups. /disable_group disables the current group. The bot should not require redeploy when adding a new group. The implementation must keep existing webhook behavior, MongoDB storage, structured logging, tests, Telegram commands, reports, alerts, and cron behavior."

## Clarifications

### Session 2026-06-09

- Q: How should group setup, authorization fallback, and report delivery behave for MVP? → A: `/setup` works only in group chats; private chat setup is rejected; admin verification may be skipped and documented as an MVP limitation; unknown groups may only use `/setup` and `/help`; `TELEGRAM_GROUP_CHAT_ID` remains an allowed fallback group; multiple configured groups share the same bot and database; command-driven reports and alerts are sent to the group where the command/action happened; cron monthly reports are sent to all enabled configured groups.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Register A New Telegram Group (Priority: P1)

A group user adds the bot to a Telegram group and runs `/setup` so the group becomes
authorized for normal bot commands without changing deployment configuration.

**Why this priority**: This is the core value of the feature. Without group self-setup,
every new group still requires environment changes and redeploys.

**Independent Test**: Add the bot to a previously unconfigured group, run `/setup` in
that group, then run an existing command such as `/status` and verify it is accepted
from that group. Also run `/setup` in a private chat and verify it is rejected.

**Acceptance Scenarios**:

1. **Given** the bot is present in an unconfigured Telegram group, **When** a user runs
   `/setup` in that group, **Then** the bot saves the current group details and confirms
   that setup is complete.
2. **Given** a group has completed setup, **When** a user in that group runs an existing
   supported command, **Then** the bot accepts and processes the command without any app
   redeploy or environment variable change.
3. **Given** a group has completed setup before, **When** a user runs `/setup` in that
   group again, **Then** the bot updates the stored group title and responds that setup
   was already completed.
4. **Given** a group is configured only through the existing fallback group setting,
   **When** a user runs an existing supported command from that group, **Then** the bot
   continues accepting the command for backward compatibility.
5. **Given** a user opens a private chat with the bot, **When** they run `/setup`, **Then**
   the bot rejects setup with a clear message that setup is available only in group chats.

---

### User Story 2 - Reject Unconfigured Groups Safely (Priority: P2)

The bot rejects operational commands from groups that have not been configured, while
still allowing `/setup` and `/help` so new groups can onboard.

**Why this priority**: Multi-group support must not accidentally let unknown groups use
work tracking commands.

**Independent Test**: Send supported operational commands from an unconfigured group and
verify they are rejected, then verify `/help` and `/setup` remain available.

**Acceptance Scenarios**:

1. **Given** a Telegram group is not configured and is not the fallback group, **When** a
   user runs `/status`, `/add_person`, `/start_work`, `/stop_work`, `/report_month`,
   `/groups`, or `/disable_group`, **Then** the bot rejects the command with a clear
   setup-required response.
2. **Given** a Telegram group is not configured, **When** a user runs `/help`, **Then**
   the bot responds with help text that includes `/setup`.
3. **Given** a Telegram group is not configured, **When** a user runs `/setup`,
   **Then** the bot allows setup instead of rejecting the command.

---

### User Story 3 - View And Disable Configured Groups (Priority: P3)

An authorized group user can list configured groups and disable the current group when it
should no longer use the bot.

**Why this priority**: Operators need visibility into configured groups and a simple way
to disable a group without redeploying.

**Independent Test**: Configure two groups, disable one group, run `/groups` from an
enabled configured group or fallback group, verify both stored groups appear with
title/chat id/enabled status, then verify `/groups` and other operational commands from
the disabled group are rejected afterward.

**Acceptance Scenarios**:

1. **Given** multiple groups are stored, including disabled groups, **When** a user runs
   `/groups` from an enabled configured group or the fallback group, **Then** the bot lists
   all stored configured groups with their title, chat id, and enabled/disabled status for
   operator visibility.
2. **Given** the current group is configured and enabled, **When** a user runs
   `/disable_group`, **Then** the bot marks the current group disabled and confirms the
   change.
3. **Given** a group has been disabled, **When** a user runs an operational command from
   that group, **Then** the bot rejects the command until setup enables the group again.

### Edge Cases

- `/setup` is run in a private chat instead of a Telegram group.
- `/setup` or `/disable_group` is run by a user who is not a group admin when admin
  verification is implemented.
- Telegram group title changes after initial setup.
- Telegram update has no group title or an empty group title.
- A configured group is disabled and later runs `/setup` again.
- The fallback group setting is present but no stored group exists yet.
- The fallback group and a stored group share the same chat id.
- `/groups` is run when there are no stored groups but the fallback group exists.
- `/groups` is run from a disabled configured group.
- `/groups` is run from a private chat.
- `/disable_group` is run in an unconfigured group.
- Telegram user information does not include a username.
- Existing reports and alerts are triggered for a configured group after setup.
- Cron monthly report generation runs when multiple groups are enabled.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support `/setup` in Telegram groups to register the current
  group as an allowed group.
- **FR-001a**: The system MUST reject `/setup` from private chats with a clear message
  that setup is available only inside group chats.
- **FR-002**: `/setup` MUST save the current Telegram chat id and group title as durable
  configured group data.
- **FR-003**: Configured group data MUST include Telegram chat id, group title, setup user
  id, setup username when available, enabled status, created at, and updated at.
- **FR-004**: `/setup` MUST be idempotent for the current group.
- **FR-005**: If `/setup` is run for an already configured group, the system MUST update
  the stored group title, preserve the existing created date, update the updated date,
  ensure the group is enabled, and respond that setup was already completed.
- **FR-006**: After successful setup, the system MUST accept existing supported commands
  from that group without redeploying or changing environment configuration.
- **FR-007**: The system MUST support multiple configured groups at the same time.
- **FR-007a**: Multiple configured groups MUST use the same bot service and the same
  database without requiring separate deployments per group.
- **FR-008**: The existing fallback group setting MUST remain supported for local
  development and backward compatibility.
- **FR-009**: A group MUST be considered allowed if it is enabled in configured group data
  or matches the fallback group setting.
- **FR-010**: The system MUST reject operational commands from unconfigured or disabled
  groups, except `/setup` and `/help`.
- **FR-011**: `/help` MUST remain available in unconfigured groups and MUST mention
  `/setup`.
- **FR-012**: `/setup` MUST be available in unconfigured groups so groups can onboard.
- **FR-013**: Admin verification for `/setup` and `/disable_group` MAY be skipped in MVP;
  if skipped, the limitation MUST be documented in user-facing or operator-facing
  documentation.
- **FR-014**: If admin verification is implemented, the system MUST respond with a clear
  authorization message when a non-admin user attempts an admin-only group command.
- **FR-015**: The system MUST allow `/groups` only from enabled configured groups or the
  fallback `TELEGRAM_GROUP_CHAT_ID` group.
- **FR-015a**: The system MUST reject `/groups` from unknown groups, disabled configured
  groups, and private chats.
- **FR-015b**: When `/groups` is allowed, the system MUST list all stored configured groups,
  including disabled groups, for operator visibility.
- **FR-016**: `/groups` MUST show each stored configured group's title, Telegram chat id,
  and enabled/disabled status.
- **FR-017**: The system MUST support `/disable_group` to disable the current configured
  group.
- **FR-018**: After `/disable_group` succeeds, operational commands from that group MUST
  be rejected unless the group is set up again or is allowed by the fallback group
  setting.
- **FR-019**: `/setup` in a disabled group MUST re-enable that group and update the stored
  title and updated date.
- **FR-020**: The system MUST keep existing webhook behavior.
- **FR-021**: The system MUST persist durable state in MongoDB Atlas.
- **FR-022**: The system MUST keep structured logging for setup attempts, setup success,
  idempotent setup updates, group listing, group disable actions, and rejected group
  commands.
- **FR-023**: The system MUST preserve existing Telegram commands, reports, alerts, and
  cron behavior for allowed groups.
- **FR-023a**: Command-driven monthly reports and alerts MUST be sent to the group where
  the command or action happened.
- **FR-023b**: Cron monthly report generation MUST send reports to all enabled configured
  groups.
- **FR-024**: The system MUST calculate elapsed time from stored timestamps.
- **FR-025**: The system MUST store timestamps in UTC and display dates in Europe/Kyiv
  timezone.
- **FR-026**: The system MUST NOT add a web UI for MVP scope.
- **FR-027**: The system MUST NOT use Telegram polling, in-memory timers, or background
  infinite loops.

### Key Entities *(include if feature involves data)*

- **Configured Group**: A Telegram group allowed to use the bot. Contains Telegram chat
  id, group title, setup user id, optional setup username, enabled status, created at,
  and updated at.
- **Fallback Group Setting**: Existing single group configuration used for local
  development and backward compatibility. It allows one group even if no stored group
  exists.
- **Group Setup Attempt**: A user action to register or re-enable the current Telegram
  group through `/setup`.
- **Group Authorization Decision**: The result of checking whether a command from a group
  should be accepted, rejected, or allowed only for setup/help.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A group user can complete setup for a new Telegram group in under 1 minute
  without changing environment variables or redeploying the app.
- **SC-002**: 100% of tested configured enabled groups can run existing supported
  commands after setup.
- **SC-003**: 100% of tested unconfigured groups are blocked from operational commands
  while still receiving `/help` and `/setup` responses.
- **SC-004**: Re-running `/setup` for an already configured group updates the title and
  returns an already-configured response in 100% of tested cases.
- **SC-005**: From an allowed group, `/groups` lists all stored configured groups with
  title, chat id, and enabled/disabled status in 100% of tested cases; from unknown,
  disabled, or private chats it is rejected in 100% of tested cases.
- **SC-006**: After `/disable_group`, operational commands from that group are rejected
  in 100% of tested cases unless the group is re-enabled through setup or matches the
  fallback group setting.
- **SC-007**: Existing webhook, report, alert, cron, timestamp, and fallback group
  behavior continue passing automated tests after multi-group setup is added.
- **SC-008**: Cron monthly report generation sends one report attempt to every enabled
  configured group in 100% of tested multi-group cron cases.
- **SC-009**: Command-driven alerts and reports are delivered to the same group where the
  triggering command or action occurred in 100% of tested multi-group cases.

## Assumptions

- Telegram groups and supergroups are both treated as group chats for setup purposes.
- If Telegram provides no group title or an empty title, setup normalizes the stored
  title to `Unknown group`.
- Admin verification for `/setup` and `/disable_group` is optional for MVP and may be
  documented as a limitation if not implemented.
- `/groups` is available from enabled configured groups and the fallback group only.
- The fallback group setting remains an allowed group even if it is not present in stored
  configured group data.
- Disabling a stored group does not delete its history.
- `/setup` in a disabled group re-enables that group.
- Group-level language settings are out of scope for this feature.
- Existing per-user language preferences, button behavior, reports, alerts, and cron
  behavior remain unchanged except for using the new allowed-group decision.
- Unknown groups may only use `/setup` and `/help`; all other commands receive a clear
  setup-required rejection.
- Command-driven reports and alerts use the current Telegram group as their delivery
  target.
- Cron monthly reports target all enabled configured groups, while the fallback group is
  primarily for local development and backward compatibility.
