# Contract: Structured Logging Events

All events use existing structured logging conventions. Logs must use safe fields only:
no Telegram bot token, no webhook secret, no cron secret, no raw request bodies, and no
secret-bearing URLs.

## Reply Keyboard Action Received

Event: `telegram.reply_keyboard_action`

Fields:

- `operation`: `telegram`
- `chat_id`
- `user_id`
- `label_key` or canonical action
- `outcome`: `received`, `unsupported`, `failure`

## Flow Lifecycle

Events:

- `telegram.flow_started`
- `telegram.flow_step_advanced`
- `telegram.flow_completed`
- `telegram.flow_cancelled`
- `telegram.flow_expired`

Fields:

- `operation`: `telegram`
- `chat_id`
- `user_id`
- `flow_type`
- `step`
- `outcome`
- `error`: safe category when failed

## User Language Changed

Event: `telegram.user_language_changed`

Fields:

- `operation`: `telegram`
- `user_id`
- `language`
- `outcome`: `success` or `failure`
- `error`: safe category when failed

## Bot Command Menu Registration

Event: `telegram.command_menu_registration`

Fields:

- `operation`: `startup`
- `command_count`
- `outcome`: `attempt`, `success`, or `failure`
- `error`: safe category when failed

## Message Delete

Event: `telegram.message_delete`

Fields:

- `operation`: `telegram_send`
- `chat_id`
- `user_id`
- `message_id`
- `message_kind`: `slash_command`, `reply_keyboard`, or `flow_input`
- `outcome`: `attempt`, `success`, or `failure`
- `error`: safe category when failed

Rules:

- Deletion failure is warn-level or error-level according to existing logging policy, but
  it must not fail the user workflow.
- Bot outputs are never deletion targets.

## Date/Time Parse

Event: `telegram.datetime_parse`

Fields:

- `operation`: `telegram`
- `chat_id`
- `user_id`
- `flow_type`
- `step`
- `format`: matched format name when successful
- `outcome`: `success`, `invalid`, `future`, or `failure`
- `error`: safe category when failed

## Callback Regression Events

Existing callback events remain valid where callbacks still exist:

- `telegram.callback_received`
- `telegram.callback_result`
- `telegram.callback_answer`
