# Quickstart: Telegram Reply Keyboard, Languages, And Start Work Flow

## Prerequisites

- Go 1.22
- MongoDB Atlas connection string configured for the existing app
- Telegram bot token and webhook settings configured in `.env`
- A Telegram group where the bot is present

## Static Checks

Run:

```sh
go test ./...
```

Expected:

- All existing tests pass.
- New tests cover reply keyboard rendering, command registration, user language lookup,
  flow state, date parsing, deletion behavior, and regression paths.

## Manual Validation

### 1. Startup Command Menu

Start the service with normal webhook configuration.

Expected logs:

- `telegram.command_menu_registration` with `outcome=attempt`
- `telegram.command_menu_registration` with `outcome=success` or safe failure details
- no token, webhook secret, or cron secret in logs

Expected Telegram command menu:

- `/setup`
- `/groups`
- `/disable_group`
- `/add_person`
- `/start_work`
- `/status`
- `/stop_work`
- `/report_month`
- `/help`

### 2. Reply Keyboard Main Menu

Open help or the main menu in an authorized group.

Expected:

- Telegram shows persistent reply keyboard buttons.
- Ukrainian group-wide labels include:
  - `Додати людину`
  - `Почати роботу`
  - `Статус`
  - `Зупинити роботу`
  - `Місячний звіт`
  - `Налаштування`
  - `Допомога`
- Main menu does not use inline buttons.

### 3. Slash Fallback

Send:

```text
/status
```

Expected:

- Status behavior matches existing command behavior.
- User command message delete is attempted.
- Bot status output remains visible.

### 4. Keyboard Action Mapping

Press:

```text
Статус
```

Expected:

- The resulting text message is handled as status.
- Message delete is attempted for the user button text message.
- Status output remains visible.

### 5. Settings Language Flow

Press:

```text
Налаштування
Мова
English
```

Expected:

- `user_settings` stores `telegram_user_id`, `language=en`, `created_at`, `updated_at`.
- Future personal prompts/settings for that user use English.
- Public reports, alerts, incident messages, status output, and monthly reports remain
  Ukrainian by default.

### 6. Start Work Now

Press:

```text
Почати роботу
```

Then enter a valid person, title, and choose:

```text
Почати зараз
```

Expected:

- `user_flow_states` is created and advanced through each step.
- Flow completes and active state is cleared.
- Work record `started_at` is stored in UTC.

### 7. Start Work Manual Date/Time

Repeat Start Work with each input:

```text
09.06 09:30
09.06.2026 09:30
2026-06-09 09:30
```

For Today/Yesterday mode, enter:

```text
09:30
```

Expected:

- Values parse as Europe/Kyiv local time.
- Omitted year uses the current Europe/Kyiv year.
- Stored work timestamp is UTC.
- Displayed time is Europe/Kyiv.

### 8. Invalid And Future Date/Time

Enter invalid and future values:

```text
31.02 09:00
25:99
2999-01-01 09:00
```

Expected:

- Bot rejects the value.
- Error is localized and includes accepted examples.
- `telegram.datetime_parse` logs invalid/future outcome.

### 9. Flow Expiration

Create a flow state whose `expires_at` is in the past, then send the next expected text.

Expected:

- Flow is treated as expired.
- Bot asks user to restart the action.
- `telegram.flow_expired` is logged.

### 10. Regression Checks

Verify existing behavior:

- `/setup` still configures groups.
- Multi-group authorization still rejects unknown/disabled groups.
- `/report_month` and cron monthly report still work.
- Stop-work alert and incident behavior still work.
- Existing callback_query updates are acknowledged and handled or rejected gracefully.
