# Contract: Cron Report Delivery

## Trigger

Existing `POST /cron/monthly-report` remains the cron entrypoint and keeps its existing
secret validation and month validation behavior.

## Delivery Targets

Cron monthly report delivery targets:
- Every enabled group in `configured_groups`.
- The `TELEGRAM_GROUP_CHAT_ID` fallback group when configured and not already present in
  the enabled configured group list.

## Duplicate Handling

Monthly report generation still uses the existing duplicate prevention by month. Delivery
fan-out sends the generated or duplicate report content to each target group.

## Per-Group Result Logging

Each target send attempt logs:
- `event=cron.monthly_report_send`
- `chat_id`
- `month`
- `outcome=success|failed`
- `duration_ms` where available
- safe error category when failed

## Response

The HTTP response should still summarize the overall cron result. If at least one target
send fails, the response should indicate failure according to existing error handling
style while logs identify the affected chat id.
