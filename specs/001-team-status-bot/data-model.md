# Data Model: Telegram Team Work Status Tracker

## Person

Represents a tracked team member.

**Fields**:
- `id`: stable unique identifier
- `first_name`: required, trimmed, non-empty
- `last_name`: required, trimmed, non-empty
- `normalized_full_name`: required, lowercased uniqueness key
- `created_at`: UTC timestamp
- `updated_at`: UTC timestamp

**Validation**:
- `first_name` and `last_name` are required.
- `normalized_full_name` must be unique for the tracked team.

**Relationships**:
- One `Person` has many `WorkRecord` documents.

## WorkRecord

Represents one work interval for a person.

**Fields**:
- `id`: stable unique identifier
- `person_id`: required reference to `Person`
- `title`: required, trimmed, non-empty
- `status`: required enum: `active`, `stopped`
- `started_at`: required UTC timestamp
- `stopped_at`: optional UTC timestamp
- `stop_reason`: optional text
- `created_at`: UTC timestamp
- `updated_at`: UTC timestamp

**Validation**:
- `title` is required.
- `started_at` is required and stored in UTC.
- `stopped_at` is set only for stopped records.
- `stopped_at` must not be earlier than `started_at`.
- Only one `active` work record may exist per person.

**Elapsed Time Rule**:
- If `status` is `active`, elapsed time is `now - started_at`.
- If `status` is `stopped`, elapsed time is `stopped_at - started_at`.
- Elapsed time is never stored as the source of truth.

**State Transitions**:
```text
active -> stopped
```

No transition reopens a stopped record in MVP. MVP has no separate reset command.

## Incident

Represents a stop event or alert failure that must be visible in reports.

**Fields**:
- `id`: stable unique identifier
- `person_id`: required reference to `Person`
- `work_record_id`: required reference to `WorkRecord`
- `type`: required enum: `stopped`, `alert_failed`
- `reason`: optional text
- `occurred_at`: required UTC timestamp
- `created_at`: UTC timestamp

**Validation**:
- Stopping a work record creates a `stopped` incident.
- Alert delivery failure creates an `alert_failed` incident or equivalent operator-visible
  error record.

## MonthlyReport

Represents a generated monthly report.

**Fields**:
- `id`: stable unique identifier
- `month`: required `YYYY-MM` key in Europe/Kyiv calendar context
- `trigger_source`: required enum: `telegram_command`, `external_http`
- `content`: required rendered report text
- `generated_at`: required UTC timestamp
- `created_at`: UTC timestamp

**Validation**:
- `month` must be unique.
- Manual and external triggers use the same uniqueness rule.
- Existing month returns existing report or duplicate notice; it does not create another
  report.
- Report membership includes every work record active at any moment during the selected
  month: records started before and overlapping the month, records started during the
  month, and records stopped during the month.

## TelegramGroupConfig

Represents the configured Telegram group for commands and alerts.

**Fields**:
- `id`: stable unique identifier
- `chat_id`: required Telegram chat identifier
- `created_at`: UTC timestamp
- `updated_at`: UTC timestamp

**Validation**:
- MVP has one configured Telegram group.
- Commands from other chats are rejected or ignored with no state change.

## Suggested MongoDB Collections

- `people`
- `work_records`
- `incidents`
- `monthly_reports`
- `telegram_group_config`

## Suggested Indexes

- `people.normalized_full_name`: unique
- `work_records.person_id`
- `work_records.person_id + status`: supports one active record per person
- `work_records.started_at`
- `work_records.stopped_at`
- `incidents.occurred_at`
- `monthly_reports.month`: unique
- `telegram_group_config.chat_id`: unique
