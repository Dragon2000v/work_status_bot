# Contract: Telegram Interactions

## Webhook Inputs

The existing `POST /telegram/webhook` route continues to accept Telegram update JSON.

### Slash Command Message

```json
{
  "update_id": 1001,
  "message": {
    "message_id": 55,
    "text": "/status",
    "from": { "id": 123456 },
    "chat": { "id": -100123, "type": "group" }
  }
}
```

Contract:

- Parse as existing slash command.
- Apply existing group authorization.
- Use saved user language for personal responses; use Ukrainian for public group output.
- Attempt to delete `message.message_id` after processing.

### Reply Keyboard Message

```json
{
  "update_id": 1002,
  "message": {
    "message_id": 56,
    "text": "Статус",
    "from": { "id": 123456 },
    "chat": { "id": -100123, "type": "group" }
  }
}
```

Contract:

- Treat known labels as canonical actions.
- Ukrainian labels must always be recognized.
- Supported localized labels may also be recognized for users with saved language.
- Apply existing group authorization before executing protected actions.
- Attempt to delete the user button text message after processing.

### Manual Flow Input

```json
{
  "update_id": 1003,
  "message": {
    "message_id": 57,
    "text": "Іван Петренко",
    "from": { "id": 123456 },
    "chat": { "id": -100123, "type": "group" }
  }
}
```

Contract:

- Look up active non-expired `user_flow_states` by `telegram_user_id` and `chat_id`.
- Interpret text according to `flow_type` and `step`.
- Advance, complete, cancel, or expire the flow.
- Attempt to delete the user input message after processing.

### Callback Query Update

```json
{
  "update_id": 1004,
  "callback_query": {
    "id": "callback-id",
    "from": { "id": 123456 },
    "message": {
      "message_id": 77,
      "chat": { "id": -100123, "type": "group" }
    },
    "data": "settings:language:uk"
  }
}
```

Contract:

- Existing callback query behavior must not regress.
- Every callback query still receives an `answerCallbackQuery` attempt.
- Callback data is not used for main-menu actions after the reply keyboard replacement.

## Main Menu Mapping

| Ukrainian label | Canonical action |
|-----------------|------------------|
| `Додати людину` | start add-person flow |
| `Почати роботу` | start start-work flow |
| `Статус` | run `/status` |
| `Зупинити роботу` | start stop-work flow |
| `Місячний звіт` | run `/report_month` |
| `Налаштування` | start settings flow |
| `Допомога` | run `/help` |

## Settings Flow

Buttons:

- `Налаштування`
- `Мова`
- `Українська`
- `English`
- `Русский`
- `Назад`
- `Скасувати`

Contract:

- Language changes only when the user explicitly selects a language in Settings.
- Saved language is keyed by Telegram user id.
- Settings prompts use saved user language when available.
- Group-wide keyboard remains Ukrainian if per-user keyboard localization is limited.

## Add Person Flow

Steps:

1. User presses `Додати людину`.
2. Bot asks for first name and last name.
3. Bot stores `add_person_enter_name` flow state for the user/chat.
4. User enters `<first_name> <last_name>`.
5. Bot validates input, creates the person through existing people behavior, attempts to
   delete the user's input message, sends localized success, and clears flow state.

Validation:

- Missing first name or last name returns localized error with an example.
- Duplicate person behavior matches existing people service behavior.
- `Скасувати` cancels and clears flow state.
- Expired state is cleared before returning localized expired-flow message.

## Start Work Flow

Steps:

1. User presses `Почати роботу` or sends `/start_work` fallback.
2. Bot asks for first name and last name unless command args already provided.
3. Bot asks for work title.
4. Bot asks time mode: `Почати зараз`, `Сьогодні`, `Вчора`, `Ввести дату вручну`, `Скасувати`.
5. Bot parses date/time in Europe/Kyiv and stores UTC.

Accepted manual formats:

- `DD.MM HH:mm`
- `DD.MM.YYYY HH:mm`
- `YYYY-MM-DD HH:mm`
- `HH:mm` for Today/Yesterday mode

Validation:

- Omitted year uses current Europe/Kyiv year.
- Future values are invalid.
- Invalid values return localized examples.

## Stop Work Flow

Steps:

1. User presses `Зупинити роботу`.
2. Bot asks for first name and last name plus optional reason.
3. Bot stores `stop_work_enter_person_or_reason` flow state for the user/chat.
4. User enters stop input.
5. Bot validates input, stops active work through existing work behavior, creates the
   existing incident/event, attempts to send the public alert to the current group,
   attempts to delete the user's input message, sends localized result, and clears flow
   state.

Validation:

- Unknown person returns localized error.
- No active work returns localized error.
- Already stopped work returns localized error when applicable.
- `Скасувати` cancels and clears flow state.
- Expired state is cleared before returning localized expired-flow message.
- If alert sending fails after work is stopped and incident/event is saved, stop remains
  successful and alert failure is recorded according to existing behavior.

## Public Output Language

These outputs use Ukrainian by default:

- public group reports
- alerts
- incident messages
- status output
- monthly reports

Incident validation must separately cover event recording, current-group alert message,
Ukrainian public output, and monthly report inclusion.
