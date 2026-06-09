# Contract: Telegram Group Commands

## `/setup`

Scope:
- May be accepted only from Telegram group or supergroup chats.
- Must be allowed from unknown groups so a new group can onboard.
- Must reject private chats with a clear message.
- Admin verification may be skipped in MVP if documented.

Stored data:
- Current chat id
- Current group title
- Setup user id
- Setup username if available
- `enabled=true`
- UTC created/updated timestamps

Outcomes:
- New group: setup success response.
- Existing enabled group: update title/setup metadata and return already-completed
  response.
- Existing disabled group: re-enable and return setup success or re-enabled response.

## `/groups`

Scope:
- Accepted only from enabled configured groups or the fallback `TELEGRAM_GROUP_CHAT_ID`
  group.
- Rejected from unknown groups, disabled configured groups, and private chats.

Output:
- Lists all stored configured groups, including disabled groups, for operator visibility.
- Each row includes group title, Telegram chat id, and enabled/disabled status.
- If no stored groups exist but fallback group is configured, response should clearly
  mention the fallback group context.

## `/disable_group`

Scope:
- Accepted only from the current configured group or fallback env group when behavior is
  explicitly supported.
- Rejected from unknown groups.
- Admin verification may be skipped in MVP if documented.

Outcomes:
- Enabled stored group: set `enabled=false`, preserve history, confirm disabled.
- Already disabled stored group: keep disabled and return idempotent response.
- Fallback-only group with no stored group: return clear message that there is no stored
  group to disable.

## Existing Operational Commands

Commands:
- `/add_person`
- `/start_work`
- `/status`
- `/stop_work`
- `/report_month`

Authorization:
- Accepted from enabled configured groups.
- Accepted from fallback env group.
- Rejected from unknown or disabled groups with setup-required response.

Delivery:
- Stop alerts are sent to the group where `/stop_work` happened.
- Manual monthly reports are sent or returned to the group where `/report_month` happened.
