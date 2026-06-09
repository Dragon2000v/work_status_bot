# Data Model: Telegram Reply Keyboard, Languages, And Start Work Flow

## UserSetting

Collection: `user_settings`

Fields:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `_id` | ObjectId | yes | MongoDB document id |
| `telegram_user_id` | int64 | yes | Unique Telegram user id |
| `language` | string | yes | `uk`, `en`, or `ru`; defaults to `uk` when missing |
| `created_at` | datetime UTC | yes | Creation timestamp |
| `updated_at` | datetime UTC | yes | Last language update timestamp |

Indexes:

- Unique index on `telegram_user_id`.

Validation:

- `telegram_user_id` must be non-zero when changing language.
- `language` must be one of `uk`, `en`, `ru`.
- Lookup returns Ukrainian when no setting exists.

## UserFlowState

Collection: `user_flow_states`

Fields:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `_id` | ObjectId | yes | MongoDB document id |
| `telegram_user_id` | int64 | yes | User whose next text input is expected |
| `chat_id` | int64 | yes | Chat where the flow is active |
| `flow_type` | string | yes | `add_person`, `start_work`, `stop_work`, or `settings` |
| `step` | string | yes | Current expected input/action |
| `payload` | object | yes | Small structured state for collected values |
| `created_at` | datetime UTC | yes | First flow creation timestamp |
| `updated_at` | datetime UTC | yes | Last state transition timestamp |
| `expires_at` | datetime UTC | yes | `updated_at + 15 minutes` stale-state cutoff |

Indexes:

- Unique compound index on `telegram_user_id`, `chat_id`.
- Index on `expires_at`; MongoDB TTL cleanup may be used for storage cleanup, but
  correctness must still check expiration during message handling.

Supported steps:

- `add_person_enter_name`
- `start_work_enter_person`
- `start_work_enter_title`
- `start_work_select_time_mode`
- `start_work_enter_time`
- `start_work_enter_manual_datetime`
- `stop_work_enter_person_or_reason`
- `settings_language_select`

State transitions:

```text
none
  -> add_person_enter_name
  -> completed | cancelled | expired

none
  -> start_work_enter_person
  -> start_work_enter_title
  -> start_work_select_time_mode
  -> start_work_enter_time | start_work_enter_manual_datetime
  -> completed | cancelled | expired

none
  -> stop_work_enter_person_or_reason
  -> completed | cancelled | expired

none
  -> settings_language_select
  -> completed | cancelled | expired
```

Payload examples:

| Flow | Payload fields |
|------|----------------|
| `add_person` | `first_name`, `last_name` after parsing input |
| `start_work` | `first_name`, `last_name`, `title`, `time_mode`, `started_at_utc` |
| `stop_work` | `first_name`, `last_name`, optional `reason` |
| `settings` | optional pending language selection metadata |

Rules:

- Starting a new flow for the same user/chat replaces the previous active state.
- Flow states expire 15 minutes after `updated_at`.
- New flow state has `expires_at = updated_at + 15 minutes`.
- Updating a flow state refreshes both `updated_at` and `expires_at`.
- Expired state is treated as inactive and must be cleared or superseded before a
  localized expired-flow message is returned.
- All timestamps are UTC.
- No process-local memory is required to continue a flow after restart.

## ReplyKeyboardAction

Not stored durably.

Fields:

| Field | Type | Notes |
|-------|------|-------|
| `label` | string | Incoming text from Telegram message |
| `language` | string | User language used for matching when available |
| `command` | string | Canonical action mapped from label |

Canonical Ukrainian mappings:

| Label | Canonical action |
|-------|------------------|
| `Додати людину` | `/add_person` flow |
| `Почати роботу` | `/start_work` flow |
| `Статус` | `/status` |
| `Зупинити роботу` | `/stop_work` flow |
| `Місячний звіт` | `/report_month` |
| `Налаштування` | settings flow |
| `Допомога` | `/help` |

Rules:

- Slash command parsing remains available for fallback.
- Button labels in supported languages may map to the same canonical action.
- Unknown labels outside an active flow return a localized unsupported-action message.

## ParsedStartTimestamp

Not stored as a separate collection.

Fields:

| Field | Type | Notes |
|-------|------|-------|
| `input` | string | User-entered value |
| `mode` | string | `now`, `today`, `yesterday`, or `manual` |
| `local_time` | datetime Europe/Kyiv | Parsed local value |
| `utc_time` | datetime UTC | Value persisted into work record |

Validation:

- Supported manual formats: `DD.MM HH:mm`, `DD.MM.YYYY HH:mm`, `YYYY-MM-DD HH:mm`.
- Supported Today/Yesterday format: `HH:mm`.
- Omitted year uses the current Europe/Kyiv year.
- Future values are rejected.
- Invalid values return localized examples.

## Existing Entities Touched

- `people`: add-person flow creates existing person records through current service.
- `work_records`: start-work flow creates active work records with parsed UTC
  `started_at`; elapsed time still derives from stored timestamps.
- `configured_groups`: existing authorization applies before command, keyboard, and flow
  actions.
- `incidents` and `monthly_reports`: existing report/alert behavior must not change.
