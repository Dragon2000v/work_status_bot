# Contract: User Settings Storage

## Collection

`user_settings`

## Document Shape

```json
{
  "telegram_user_id": 123456,
  "language": "uk",
  "created_at": "2026-06-08T12:00:00Z",
  "updated_at": "2026-06-08T12:00:00Z"
}
```

## Language Codes

| Code | Language |
|------|----------|
| `uk` | Ukrainian |
| `en` | English |
| `ru` | Russian |

## Lookup Rules

1. If a user has a saved valid language, use it.
2. Otherwise use Ukrainian.
3. Only change language when the user explicitly selects a language in Settings.

## Write Rules

- Upsert by `telegram_user_id`.
- Create new document with `created_at` and `updated_at` set to current UTC time.
- Update existing document by changing `language` and `updated_at`; preserve
  `created_at`.
- Reject unsupported language codes before writing.

## Indexes

- Unique index: `{ "telegram_user_id": 1 }`

## Out Of Scope

- Group-level language settings.
- Automatic language detection from Telegram profile.
- Storing translation text in MongoDB.
