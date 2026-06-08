# Feature Specification: Structured Application Logging

**Feature Branch**: `002-structured-logging`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "Add structured application logging to the existing Go Telegram bot. The system should log application startup, HTTP requests, MongoDB connection status, Telegram webhook updates, command handling, cron report triggers, and Telegram alert send results. Logs must help debug local development and production issues. The system must not log secrets, including Telegram bot token, MongoDB password, webhook secret, cron secret, or full MongoDB URI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Diagnose Request Flow (Priority: P1)

An operator can inspect logs for a request and understand which route was called, whether it succeeded, how long it took, and which major application action happened during the request.

**Why this priority**: Request flow visibility is the fastest way to debug local setup, webhook delivery, and production incidents.

**Independent Test**: Can be tested by sending health, webhook, and cron requests, then confirming logs contain structured request events with outcome and timing without exposing secrets.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** any HTTP route receives a request, **Then** a log entry records method, path, status outcome, and duration.
2. **Given** a request fails, **When** the failure is handled, **Then** logs identify the route and error category without exposing request secrets.

---

### User Story 2 - Trace Bot Operations (Priority: P2)

An operator can follow Telegram webhook updates and command handling through logs to see which command was processed and whether the bot sent responses or alerts successfully.

**Why this priority**: Most user-visible behavior happens through Telegram commands and alerts, so operators need enough context to diagnose command issues.

**Independent Test**: Can be tested by sending supported and malformed webhook updates, then confirming logs show chat validation result, command name, handling result, and alert send result.

**Acceptance Scenarios**:

1. **Given** a Telegram webhook update arrives from the configured group, **When** it contains a supported command, **Then** logs record the update, command name, and success or failure result.
2. **Given** a Telegram alert send fails, **When** the command completes with a visible warning, **Then** logs record the alert failure result without exposing the bot token or full request URL.

---

### User Story 3 - Confirm Startup And Data Connectivity (Priority: P3)

An operator can confirm from logs whether startup reached configuration loading, database connection, index setup, and route serving.

**Why this priority**: Startup and database visibility helps diagnose deployment and local development setup before users interact with the bot.

**Independent Test**: Can be tested by starting the service with valid and invalid local configuration, then confirming logs show startup progress and safe connection status details.

**Acceptance Scenarios**:

1. **Given** valid configuration, **When** the service starts, **Then** logs show startup progress, database connection success, index setup result, and serving address.
2. **Given** database connection fails, **When** the service starts, **Then** logs show a database connection failure category without logging the MongoDB password or full connection URI.

### Edge Cases

- Logs must not include Telegram bot token, MongoDB password, webhook secret, cron secret, or full MongoDB URI, including in failure messages.
- Logs for rejected Telegram chats must avoid exposing unnecessary message content while still recording that chat validation failed.
- Logs for malformed webhook payloads must identify invalid input handling without dumping full payload bodies.
- Logs for cron report triggers must indicate generated, duplicate, invalid, unauthorized, or failed outcomes without logging the cron secret.
- Logs must remain useful when multiple requests occur close together by including enough structured context to correlate each event.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST produce structured log entries with consistent field names for event type, severity, timestamp, route or operation, outcome, and duration where applicable.
- **FR-002**: System MUST log application startup progress, including configuration loading result, database connection status, index setup result, and service listen address.
- **FR-003**: System MUST log each HTTP request with method, route path, status code or outcome category, and elapsed handling time.
- **FR-004**: System MUST log MongoDB connection attempts and outcomes without exposing credentials or the full connection string.
- **FR-005**: System MUST log Telegram webhook update handling with configured-chat validation result, command name when available, and handling outcome.
- **FR-006**: System MUST log cron monthly report triggers with source, selected month when available, generated or duplicate result, authorization failure, invalid month, and handling failure outcomes.
- **FR-007**: System MUST log Telegram message or alert send attempts with success or failure outcome while hiding token-bearing URLs and secrets.
- **FR-008**: System MUST avoid logging secret values, including Telegram bot token, MongoDB password, webhook secret, cron secret, full MongoDB URI, and raw authorization headers.
- **FR-009**: System MUST avoid dumping full request bodies, full Telegram update payloads, and full environment configuration into logs.
- **FR-010**: System MUST keep logs usable for both local development and production by making entries human-readable and machine-parseable.
- **FR-011**: System MUST use Telegram webhook handling for bot interactions.
- **FR-012**: System MUST persist durable state in MongoDB Atlas.
- **FR-013**: System MUST calculate elapsed time from stored timestamps.
- **FR-014**: System MUST store timestamps in UTC and display dates in Europe/Kyiv timezone.
- **FR-015**: System MUST NOT add a web UI for MVP scope.

### Key Entities *(include if feature involves data)*

- **Log Event**: A single structured application observation with event type, time, severity, operation context, outcome, duration, and safe identifiers.
- **Safe Identifier**: Non-secret context that helps debugging, such as route path, command name, month key, status code, and high-level error category.
- **Secret Value**: Sensitive configuration or request data that must never appear in logs, including tokens, passwords, secrets, full connection strings, and secret-bearing headers or URLs.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For each health, webhook, and cron request, operators can find a corresponding structured request log entry in 100% of tested requests.
- **SC-002**: For supported Telegram commands, operators can identify command name and success or failure outcome from logs in 100% of tested command flows.
- **SC-003**: For startup with valid configuration, operators can confirm service readiness and database readiness from logs in under 2 minutes.
- **SC-004**: For alert delivery success and failure test cases, logs show the send outcome in 100% of tested cases.
- **SC-005**: Automated or manual secret checks find zero occurrences of configured secret values, full MongoDB URIs, Telegram bot tokens, webhook secrets, or cron secrets in generated logs.

## Assumptions

- Logging is written to the process standard output or standard error so local terminals and hosting platforms can collect it.
- Existing behavior for Telegram commands, MongoDB persistence, cron triggering, timestamps, and reports remains unchanged.
- Logs may include safe operational identifiers, but not personal message bodies beyond the command name needed to diagnose command routing.
- Existing request and command tests can be extended to validate logging behavior and secret redaction.
