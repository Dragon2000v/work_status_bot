# Contract: Telegram Interactions

## Webhook Update Inputs

The existing Telegram webhook route continues to accept Telegram update JSON.

### Message Command Update

Required fields for command handling:

```json
{
  "update_id": 1001,
  "message": {
    "message_id": 55,
    "text": "/status",
    "from": { "id": 123456 },
    "chat": { "id": -100123 }
  }
}
```

Contract:
- `message.text` is parsed as an existing text command.
- `message.message_id` is used only for best-effort deletion of the user command.
- `message.from.id` is used for language lookup on user-specific responses.
- After command processing, the bot attempts to delete only this user command message.

### Callback Query Update

Required fields for callback handling:

```json
{
  "update_id": 1002,
  "callback_query": {
    "id": "callback-id",
    "from": { "id": 123456 },
    "message": {
      "message_id": 77,
      "chat": { "id": -100123 }
    },
    "data": "menu:status"
  }
}
```

Contract:
- Every callback query must receive an `answerCallbackQuery` attempt.
- `callback_query.from.id` drives user language lookup.
- `callback_query.data` is routed by stable action values, never localized labels.
- Unsupported or malformed `data` is acknowledged and returns a localized error.

## Callback Actions

| Callback data | Result |
|---------------|--------|
| `menu:add_person` | Start localized Add person prompt or show usage guidance |
| `menu:start_work` | Start localized Start work prompt or show usage guidance |
| `menu:status` | Send localized status output for the pressing user |
| `menu:stop_work` | Start localized Stop work prompt or show usage guidance |
| `menu:report_month` | Generate or show public Ukrainian monthly report flow |
| `menu:settings` | Show localized settings menu |
| `menu:help` | Show localized help text |
| `settings:language` | Show language selection menu |
| `settings:language:uk` | Save Ukrainian for the pressing user |
| `settings:language:en` | Save English for the pressing user |
| `settings:language:ru` | Save Russian for the pressing user |

## Telegram Bot API Methods

### `sendMessage`

Used for existing bot responses and new inline menus.

Payload without keyboard:

```json
{
  "chat_id": -100123,
  "text": "..."
}
```

Payload with inline keyboard:

```json
{
  "chat_id": -100123,
  "text": "...",
  "reply_markup": {
    "inline_keyboard": [
      [
        { "text": "Статус", "callback_data": "menu:status" }
      ]
    ]
  }
}
```

### `answerCallbackQuery`

Used for every callback query.

```json
{
  "callback_query_id": "callback-id"
}
```

Optional localized `text` may be included for short callback feedback when useful.

### `deleteMessage`

Used only for best-effort deletion of user command messages.

```json
{
  "chat_id": -100123,
  "message_id": 55
}
```

Deletion failure is logged and does not fail the command.

### `editMessageText`

Optional. Use only when updating an existing menu message is clearer than sending a new
message.

```json
{
  "chat_id": -100123,
  "message_id": 77,
  "text": "...",
  "reply_markup": {
    "inline_keyboard": [
      [
        { "text": "Українська", "callback_data": "settings:language:uk" }
      ]
    ]
  }
}
```
