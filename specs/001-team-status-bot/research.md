# Research: Telegram Team Work Status Tracker

## Decision: Go 1.22+ Backend

**Rationale**: Go fits a small HTTP webhook service with explicit error handling,
context-aware request flow, and straightforward deployment on free hosting.

**Alternatives considered**:
- Node.js: viable, but project constitution and user input require Go.
- Python: fast for scripting, but less aligned with the project direction.

## Decision: chi HTTP Router

**Rationale**: chi is small, idiomatic, and sufficient for `GET /health`,
`POST /telegram/webhook`, and `POST /cron/monthly-report`. It avoids heavier web
frameworks while keeping middleware and route definitions explicit.

**Alternatives considered**:
- `net/http` only: simplest, but chi gives cleaner routing with little overhead.
- Gin/Echo/Fiber: more framework surface than MVP needs.

## Decision: Direct Telegram Bot API HTTP Requests

**Rationale**: Direct HTTP calls keep dependencies small and make webhook handling
explicit. The bot only needs to send messages and receive webhook updates for MVP.

**Alternatives considered**:
- Telegram bot framework: convenient, but adds abstractions not needed for MVP.
- Telegram polling: rejected by constitution and user requirements.

## Decision: Telegram Webhook Updates Only

**Rationale**: Webhooks are request-driven and compatible with free hosting where the
server may sleep. The service does not need a continuously running receiver.

**Alternatives considered**:
- Long polling: prohibited.
- Background update loop: prohibited.

## Decision: MongoDB Atlas + Official MongoDB Go Driver

**Rationale**: Atlas provides managed durable storage. The official driver supports
contexts, indexes, atomic updates, and direct BSON control without extra persistence
layers.

**Alternatives considered**:
- Redis: rejected by user requirement and not durable enough for report history.
- SQL database: viable, but user requires MongoDB Atlas.
- Queue-backed writes: rejected by user requirement.

## Decision: External Scheduled HTTP Trigger For Monthly Reports

**Rationale**: `POST /cron/monthly-report` lets an external scheduler wake the service
and request report generation without in-process cron loops, timers, or background
workers. Manual `/report_month` and external triggers share duplicate prevention.

**Alternatives considered**:
- In-process cron loop: prohibited.
- Queue or worker system: rejected by user requirement and MVP scope.

## Decision: UTC Persistence, Europe/Kyiv Display

**Rationale**: UTC timestamps make elapsed-time calculations stable across sleeps,
restarts, and DST changes. Europe/Kyiv display satisfies user-facing timezone needs.
Elapsed time is always derived from stored `started_at` and optional `stopped_at`.

**Alternatives considered**:
- Store local time: rejected due DST and portability risk.
- Store elapsed counters: prohibited because elapsed time must derive from timestamps.

## Decision: One Active Work Record Per Person

**Rationale**: The spec assumes one active record per person while allowing many stopped
historical records. This keeps `/status`, `/stop_work`, and incident creation
unambiguous.

**Alternatives considered**:
- Multiple active records per person: possible, but command UX becomes ambiguous for
  `/stop_work` unless record IDs are exposed.

## Decision: Stored Monthly Report Idempotency

**Rationale**: A unique month key on reports prevents duplicate reports from manual and
external triggers. If a report already exists, return existing report content or a clear
duplicate notice.

**Alternatives considered**:
- Always regenerate: violates duplicate avoidance.
- Keep only latest report in memory: fails after sleep/restart.
