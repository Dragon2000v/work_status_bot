# Contract: Telegram API Methods

All Telegram API calls use the existing Telegram client boundary. Logs must never include
the bot token, webhook secret, cron secret, raw webhook payload, or secret-bearing URLs.

## `setMyCommands`

Called during application startup.

Payload:

```json
{
  "commands": [
    { "command": "setup", "description": "..." },
    { "command": "groups", "description": "..." },
    { "command": "disable_group", "description": "..." },
    { "command": "add_person", "description": "..." },
    { "command": "start_work", "description": "..." },
    { "command": "status", "description": "..." },
    { "command": "stop_work", "description": "..." },
    { "command": "report_month", "description": "..." },
    { "command": "help", "description": "..." }
  ]
}
```

Contract:

- Registration is attempted on startup before serving normal traffic.
- Attempt, success, and failure are logged.
- Failure does not stop application startup.
- When registration fails, the bot still handles slash commands and reply keyboard
  actions.
- Logs must not include the bot token, webhook secret, cron secret, raw webhook payload,
  or secret-bearing URLs.

## `sendMessage` with `ReplyKeyboardMarkup`

Used for main menu, settings, language selection, and flow choice prompts.

Payload:

```json
{
  "chat_id": -100123,
  "text": "Оберіть дію:",
  "reply_markup": {
    "keyboard": [
      [{ "text": "Додати людину" }, { "text": "Почати роботу" }],
      [{ "text": "Статус" }, { "text": "Зупинити роботу" }],
      [{ "text": "Місячний звіт" }],
      [{ "text": "Налаштування" }, { "text": "Допомога" }]
    ],
    "resize_keyboard": true,
    "one_time_keyboard": false,
    "is_persistent": true
  }
}
```

Contract:

- `resize_keyboard=true`.
- `one_time_keyboard=false`.
- `is_persistent=true` when supported by the Telegram model in use.
- Ukrainian is used for group-wide keyboards when per-user keyboard localization is not
  reliable in group chats.

## `deleteMessage`

Used for best-effort cleanup of user messages only.

Payload:

```json
{
  "chat_id": -100123,
  "message_id": 55
}
```

Contract:

- Attempt after slash command messages.
- Attempt after reply keyboard button text messages.
- Attempt after manual input inside active flows.
- Never target bot responses, alerts, reports, incident messages, status output, or
  monthly reports.
- Failure is logged and does not change command/flow outcome.

## Existing Methods

`setWebhook`, `sendMessage` without keyboard, `answerCallbackQuery`, and existing
callback-query handling remain supported. Callback behavior must not regress if callback
updates are still received.
