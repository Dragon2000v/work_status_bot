# Contract: Structured Logging Events

All events use existing structured logging conventions: safe fields only, no Telegram bot
token, no secret-bearing URLs, no full webhook payload dumps, and no raw request bodies.

## Callback Received

Event: `telegram.callback_received`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `callback_action`
- `outcome`: `received`

## Callback Handled

Event: `telegram.callback_result`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `callback_action`
- `outcome`: `success`, `invalid`, `unsupported`, or `failure`
- `error`: safe error category when failed

## Callback Answer

Event: `telegram.callback_answer`

Fields:
- `operation`: `telegram_send`
- `user_id`
- `callback_action`
- `outcome`: `success` or `failure`
- `error`: safe error category when failed

## User Language Changed

Event: `telegram.user_language_changed`

Fields:
- `operation`: `telegram`
- `user_id`
- `language`
- `outcome`: `success` or `failure`
- `error`: safe error category when failed

## Command Message Delete

Event: `telegram.command_delete`

Fields:
- `operation`: `telegram_send`
- `chat_id`
- `user_id`
- `message_id`
- `command`
- `outcome`: `attempt`, `success`, or `failure`
- `error`: safe error category when failed

## Public Output Language

Reports, alerts, incident messages, and monthly reports should either omit a language
field or log `language=uk` when a language field is useful for diagnostics.
