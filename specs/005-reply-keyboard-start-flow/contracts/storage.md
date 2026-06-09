# Contract: MongoDB Storage

All persisted timestamps are UTC.

## Collection: `user_settings`

Purpose: Durable per-user language preference.

Document:

```json
{
  "_id": "ObjectId",
  "telegram_user_id": 123456,
  "language": "uk",
  "created_at": "2026-06-09T10:00:00Z",
  "updated_at": "2026-06-09T10:00:00Z"
}
```

Indexes:

- Unique: `telegram_user_id`

Rules:

- Missing setting means Ukrainian.
- Only Settings language selection mutates `language`.
- Supported values: `uk`, `en`, `ru`.

## Collection: `user_flow_states`

Purpose: Durable state for multi-step Telegram text flows.

Document:

```json
{
  "_id": "ObjectId",
  "telegram_user_id": 123456,
  "chat_id": -100123,
  "flow_type": "start_work",
  "step": "start_work_enter_title",
  "payload": {
    "first_name": "Іван",
    "last_name": "Петренко"
  },
  "created_at": "2026-06-09T10:00:00Z",
  "updated_at": "2026-06-09T10:01:00Z",
  "expires_at": "2026-06-09T10:16:00Z"
}
```

Indexes:

- Unique compound: `telegram_user_id`, `chat_id`
- Non-unique: `expires_at`
- Optional TTL on `expires_at` for eventual storage cleanup

Rules:

- Lookup by user/chat before treating a non-command text message as unsupported.
- `expires_at` is always `updated_at + 15 minutes`.
- Flow state updates refresh both `updated_at` and `expires_at`.
- Expired states are invalid even if still present in MongoDB.
- Expired states are deleted or cleared before returning a localized expired-flow
  message.
- Starting a new flow for the same user/chat replaces the active state.
- Completion, cancellation, or expiration clears active state.
- Correctness must not depend on a background cleanup loop.

## Existing Collections

Existing behavior remains:

- `people`: managed by add-person flow and existing `/add_person`.
- `work_records`: start-work stores UTC `started_at`; stop-work and elapsed time use stored timestamps.
- `configured_groups`: authorizes commands, button actions, and flow inputs.
- `incidents`, `monthly_reports`: incident/alert public output remains Ukrainian by
  default, alert failure recording remains unchanged, and monthly reports keep including
  incidents.
