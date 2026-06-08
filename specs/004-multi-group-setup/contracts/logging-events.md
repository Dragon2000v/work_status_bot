# Contract: Structured Logging Events

Logs must not include bot token, webhook secret, cron secret, raw Telegram update payloads,
or full request bodies.

## Group Setup

Event: `telegram.group_setup`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `outcome`: `attempt`, `created`, `updated`, `re_enabled`, `rejected`, or `failure`
- `reason`: safe rejection or failure category when applicable

## Group Authorization Rejected

Event: `telegram.group_authorization`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `command`
- `outcome`: `rejected`
- `reason`: `unknown_group`, `disabled_group`, or `private_chat_setup`

## Group List

Event: `telegram.group_list`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `outcome`: `success` or `failure`
- `group_count` when successful

## Group Disable

Event: `telegram.group_disable`

Fields:
- `operation`: `telegram`
- `chat_id`
- `user_id`
- `outcome`: `disabled`, `already_disabled`, `rejected`, or `failure`

## Cron Per-Group Report Send

Event: `cron.monthly_report_send`

Fields:
- `operation`: `cron`
- `chat_id`
- `month`
- `outcome`: `success` or `failed`
- `duration_ms`
- `error`: safe category when failed

## Alert Send Target

Event: `telegram.alert_send`

Fields:
- `operation`: `telegram_send`
- `chat_id`: group where stop action happened
- `outcome`: `success` or `failure`
- `error`: safe category when failed
