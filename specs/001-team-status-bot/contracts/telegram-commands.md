# Telegram Command Contract

Commands are accepted from the configured Telegram group only. MVP has no authentication
or role checks.

## /help

Shows supported commands and concise usage.

**Expected Output**:
- `/add_person`
- `/start_work`
- `/status`
- `/stop_work`
- `/report_month`

## /add_person

Adds a tracked person.

**Format**:

```text
/add_person <first_name> <last_name>
```

**Success**:
- Confirms person was added.

**Errors**:
- Missing first or last name.
- Duplicate person.

## /start_work

Starts a work record for an existing person.

**Format**:

```text
/start_work <first_name> <last_name> "<title>"
```

**Success**:
- Confirms active work started.
- Shows person, title, and start date/time in Europe/Kyiv timezone.

**Errors**:
- Person not found.
- Missing title.
- Person already has active work.
- Duplicate active work for the same person and title.

## /status

Lists people and current work status.

**Format**:

```text
/status
```

**Success**:
- Lists every person.
- For active work, shows title, status, start date/time, and elapsed time calculated from
  stored timestamps.
- For no active work, shows inactive status.

## /stop_work

Stops a person's active work record and creates an incident.

**Format**:

```text
/stop_work <first_name> <last_name> [reason]
```

**Success**:
- Stops active work.
- Stores stop date/time and optional reason.
- Creates a stopped incident.
- Attempts to send an alert to the configured Telegram group immediately.
- If alert delivery fails, returns a visible warning and records an alert failure event or
  equivalent alert failure state.

**Errors**:
- Person not found.
- No active work.
- Work record already stopped.
- Malformed command format or missing arguments.

## /report_month

Generates or returns a monthly report.

**Format**:

```text
/report_month [YYYY-MM]
```

If month is omitted, current Europe/Kyiv month is used.

**Success**:
- Lists all tracked people.
- Lists every work record active at any moment during the selected month, including
  records started before the month and still active during it, records started during the
  month, and records stopped during the month.
- Includes elapsed time calculated from stored timestamps.
- Includes incidents for the month.
- Stores one report per month.

**Duplicate**:
- If a report for the month already exists, returns existing report content or a clear
  duplicate notice.

**Errors**:
- Invalid month format.
- Malformed command format.

## Unsupported In MVP

- No separate reset command is supported in MVP.
- Any interruption is handled by `/stop_work`.
