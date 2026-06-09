# Data Model: Multi-Group Setup

## ConfiguredGroup

Durable Telegram group allowed to use the bot. Stored in MongoDB collection
`configured_groups`.

**Fields**:
- `id`: Internal document identifier.
- `telegram_chat_id` (int64): Telegram group or supergroup chat id. Required. Unique.
- `title` (string): Latest known group title. Required for group setup; may be updated by
  later setup runs.
- `setup_user_id` (int64): Telegram user id that most recently ran setup. Required.
- `setup_username` (string): Telegram username of setup user when available. Optional.
- `enabled` (bool): Whether the group is allowed to run operational commands. Required.
- `created_at` (datetime): UTC timestamp when the group record was first created.
- `updated_at` (datetime): UTC timestamp when the group record was last changed.

**Validation rules**:
- `telegram_chat_id` must be non-zero.
- `title` is trimmed; if Telegram provides no title or an empty title, normalize it to
  `Unknown group` for display while preserving chat id as identity.
- `setup_user_id` must be non-zero when setup is run from a group.
- `created_at` and `updated_at` are stored in UTC.
- Unique index on `telegram_chat_id`.

**State transitions**:
- Missing group -> `/setup` creates group with `enabled=true`.
- Enabled group -> repeated `/setup` updates title/setup user fields/updated_at and keeps
  `enabled=true`.
- Disabled group -> `/setup` re-enables group and updates title/setup user fields.
- Enabled group -> `/disable_group` sets `enabled=false` and updates `updated_at`.
- Disabled group -> repeated `/disable_group` remains disabled and returns an idempotent
  response.

## GroupAuthorizationDecision

Result of checking whether a Telegram update should be handled.

**Fields**:
- `chat_id`: Telegram chat id being checked.
- `command`: Parsed command name.
- `allowed`: Whether the command may proceed.
- `reason`: `configured`, `fallback`, `setup_allowed`, `help_allowed`,
  `unknown_group`, `disabled_group`, or `private_chat_setup`.

**Rules**:
- Enabled configured group is allowed.
- Fallback `TELEGRAM_GROUP_CHAT_ID` is allowed when present.
- Unknown groups may run only `/setup` and `/help`.
- Private chats may not run `/setup`.
- Unknown or disabled groups reject all operational commands.

## GroupSetupRequest

Data extracted from a Telegram group message when `/setup` is run.

**Fields**:
- `telegram_chat_id`
- `title`
- `setup_user_id`
- `setup_username`
- `occurred_at`

**Rules**:
- Valid only for Telegram group or supergroup chats.
- Admin verification is optional for MVP. If skipped, documentation must state the
  limitation.

## ReportDeliveryTarget

Telegram group that should receive a report or alert.

**Fields**:
- `telegram_chat_id`
- `source`: `current_chat`, `configured_group`, or `fallback_group`
- `title`

**Rules**:
- Command-driven reports and alerts use `current_chat`.
- Cron monthly reports use all enabled configured groups and include fallback chat if
  configured and not duplicated.
