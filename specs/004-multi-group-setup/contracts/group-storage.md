# Contract: Group Storage

## Collection

`configured_groups`

## Document Shape

```json
{
  "_id": "object-id",
  "telegram_chat_id": -1001234567890,
  "title": "Work Team",
  "setup_user_id": 123456,
  "setup_username": "admin_user",
  "enabled": true,
  "created_at": "2026-06-09T10:00:00Z",
  "updated_at": "2026-06-09T10:00:00Z"
}
```

## Indexes

- Unique index: `{ "telegram_chat_id": 1 }`
- Query index: `{ "enabled": 1 }`

## Write Rules

- `/setup` upserts by `telegram_chat_id`.
- New group sets `created_at` and `updated_at` to current UTC time.
- Existing group preserves `created_at`, updates `title`, setup user fields,
  `enabled=true`, and `updated_at`.
- `/disable_group` sets `enabled=false` and updates `updated_at`; it never deletes the
  document.

## Read Rules

- Authorization lookup reads by `telegram_chat_id`.
- `/groups` reads all stored groups so disabled groups remain visible to operators.
- Cron report delivery reads enabled groups only.
- Fallback env group is not automatically persisted.
