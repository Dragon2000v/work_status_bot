# HTTP API Contract

## GET /health

Health check endpoint for hosting platforms.

**Request**:
- Method: `GET`
- Body: none

**Success Response**:
- Status: `200 OK`
- Body:

```json
{
  "status": "ok"
}
```

## POST /telegram/webhook

Receives Telegram webhook updates. This is the only Telegram update delivery path.

**Request**:
- Method: `POST`
- Content-Type: `application/json`
- Body: Telegram update JSON from Telegram Bot API

**Behavior**:
- Parse the incoming update.
- Accept supported commands from the configured Telegram group.
- Return quickly after command handling.
- Do not poll Telegram.
- Do not start background loops or timers.

**Success Response**:
- Status: `200 OK`
- Body:

```json
{
  "ok": true
}
```

**Error Responses**:
- `400 Bad Request`: malformed update JSON
- `403 Forbidden`: update is from an unconfigured chat
- `500 Internal Server Error`: command processing failed

## POST /cron/monthly-report

External scheduled trigger for monthly report generation.

**Request**:
- Method: `POST`
- Content-Type: `application/json`
- Header: `X-Cron-Secret: <CRON_SECRET>`
- Body:

```json
{
  "month": "2026-06"
}
```

`month` is optional. If omitted, the current Europe/Kyiv month is used.

**Behavior**:
- Reject the request unless `X-Cron-Secret` matches the configured `CRON_SECRET`.
- Generate the same report as `/report_month`.
- Use the same duplicate prevention as manual reports.
- Send the report or duplicate notice to the configured Telegram group.
- Do not run an in-process scheduler.

**Success Response: Generated**
- Status: `200 OK`
- Body:

```json
{
  "status": "generated",
  "month": "2026-06"
}
```

**Success Response: Duplicate**
- Status: `200 OK`
- Body:

```json
{
  "status": "duplicate",
  "month": "2026-06"
}
```

**Error Responses**:
- `401 Unauthorized`: missing or invalid `X-Cron-Secret`
- `400 Bad Request`: invalid month format
- `500 Internal Server Error`: report generation or Telegram delivery failed
