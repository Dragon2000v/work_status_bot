# Data Model: Telegram Buttons And User Languages

## UserSetting

Durable per-user Telegram preference stored in MongoDB collection `user_settings`.

**Fields**:
- `telegram_user_id` (int64): Telegram user id. Required. Unique.
- `language` (string): One of `uk`, `en`, `ru`. Required.
- `created_at` (datetime): UTC timestamp when the setting was first created. Required.
- `updated_at` (datetime): UTC timestamp when the setting was last changed. Required.

**Validation rules**:
- `telegram_user_id` must be non-zero.
- `language` must be one of the supported language codes.
- Empty, unknown, or unsupported language values resolve to Ukrainian at read time.
- `created_at` and `updated_at` are stored in UTC.

**Indexes**:
- Unique index on `telegram_user_id`.

**State transitions**:
- Missing setting -> fallback language `uk`.
- Missing setting -> created setting when user explicitly selects a language in Settings.
- Existing setting -> updated setting when user explicitly selects a different language.
- Existing setting -> unchanged when user selects the same language again.

## Language

Supported user-facing language code.

**Values**:
- `uk`: Ukrainian. Default for all fallback and public group output.
- `en`: English.
- `ru`: Russian.

**Validation rules**:
- Translation lookup must always return text; if the requested language or key is missing,
  fallback to Ukrainian for that key.

## TranslationKey

Stable identifier for localized text in `internal/i18n`.

**Fields**:
- `key`: Stable code-level identifier, for example `menu.add_person`,
  `settings.language`, `error.malformed_command`.
- `translations`: Text values for `uk`, `en`, and `ru`.

**Validation rules**:
- Every key used by Telegram messages or buttons must have a Ukrainian value.
- Tests should cover representative keys for all supported languages and fallback paths.

## InlineMenu

Telegram inline button layout rendered for a specific user language.

**Fields**:
- `message_text`: Localized menu or prompt text.
- `rows`: Ordered rows of buttons.
- `button_text`: Localized button label.
- `callback_data`: Stable callback action value independent of display language.

**Validation rules**:
- Main menu must include Add person, Start work, Status, Stop work, Monthly report,
  Settings, and Help.
- Settings menu must include language change entry.
- Language menu must include Ukrainian, English, and Russian choices.
- Callback data must not contain localized text.

## CallbackAction

Stable action routed from Telegram callback query data.

**Values**:
- `menu:add_person`
- `menu:start_work`
- `menu:status`
- `menu:stop_work`
- `menu:report_month`
- `menu:settings`
- `menu:help`
- `settings:language`
- `settings:language:uk`
- `settings:language:en`
- `settings:language:ru`

**Validation rules**:
- Unsupported, malformed, or stale callback data must be acknowledged and produce a
  localized user-facing error.
- Supported actions must use the language of the user who pressed the button, except
  public reports and alerts, which use Ukrainian.

## CommandDeletionAttempt

Request-scoped observation for best-effort deletion of a user command message.

**Fields**:
- `chat_id`: Telegram chat id containing the command.
- `message_id`: Telegram message id of the user's command.
- `telegram_user_id`: Telegram user id of the command sender when available.
- `command`: Parsed command name when available.
- `outcome`: `attempt`, `success`, or `failure`.
- `error_category`: Safe high-level failure detail when deletion fails.

**Validation rules**:
- Only user command messages are deletion candidates.
- Bot responses, reports, alerts, incident messages, and monthly reports are never
  deletion candidates.
- Deletion failure does not change the command result.
