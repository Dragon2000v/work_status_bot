<!--
Sync Impact Report
Version change: unversioned template -> 1.0.0
Modified principles:
- PRINCIPLE_1_NAME -> I. Simple Clean Architecture
- PRINCIPLE_2_NAME -> II. Go Idioms and Explicit Code
- PRINCIPLE_3_NAME -> III. Webhook-Only Telegram Integration
- PRINCIPLE_4_NAME -> IV. MongoDB Atlas Persistence
- PRINCIPLE_5_NAME -> V. Timestamp-Based Time Handling
Added sections:
- MVP Scope Constraints
- Development Workflow
Removed sections:
- None
Templates requiring updates:
- updated .specify/templates/plan-template.md
- updated .specify/templates/spec-template.md
- updated .specify/templates/tasks-template.md
- reviewed .specify/extensions/agent-context/commands/speckit.agent-context.update.md
- reviewed .specify/extensions/git/commands/speckit.git.initialize.md
- reviewed .specify/extensions/git/commands/speckit.git.commit.md
- reviewed .specify/extensions/git/commands/speckit.git.feature.md
- reviewed .specify/extensions/git/commands/speckit.git.remote.md
- reviewed .specify/extensions/git/commands/speckit.git.validate.md
- reviewed AGENTS.md
Follow-up TODOs:
- None
-->
# Work Status Bot Constitution

## Core Principles

### I. Simple Clean Architecture
The project MUST use simple clean architecture with clear boundaries between Telegram
delivery, application services, and persistence. Services MUST be small and have clear
responsibilities. Interfaces are allowed at boundaries, but abstractions MUST exist only
when they reduce coupling or make tests clearer.

Rationale: The bot is small, but Telegram and database code still need separation so
features stay readable and testable without overengineering.

### II. Go Idioms and Explicit Code
Code MUST follow Go idioms: small packages, explicit error handling, context-aware
operations, clear names, and standard library solutions where practical. Control flow
MUST stay readable and direct. Clever generic abstractions, hidden side effects, and
framework-heavy designs are not acceptable for MVP.

Rationale: Simple Go code is easier to review, debug, and operate for a small bot.

### III. Webhook-Only Telegram Integration
Telegram integration MUST use webhooks. Telegram polling is prohibited. Background
infinite loops are prohibited. In-memory timers are prohibited. Request handling MUST be
driven by inbound webhook updates and explicit external events, not process-local
schedulers.

Rationale: Webhooks avoid polling infrastructure and keep runtime behavior stateless
enough for predictable deployments.

### IV. MongoDB Atlas Persistence
MongoDB Atlas is the database of record. The project MUST use the official MongoDB Go
Driver. Durable user, status, and event state MUST be persisted in MongoDB rather than
kept only in process memory. Repository code MUST use contexts and return explicit
errors.

Rationale: Atlas provides managed persistence, while the official driver avoids
unnecessary adapter layers and unsupported database clients.

### V. Timestamp-Based Time Handling
All elapsed time MUST be calculated from timestamps stored in MongoDB. The database MUST
store timestamps in UTC. User-facing date and time output MUST be displayed in the
Europe/Kyiv timezone. Code MUST NOT rely on in-memory elapsed counters, timers, or
process uptime to calculate work durations.

Rationale: Stored UTC timestamps keep calculations correct across restarts, deploys, and
timezone changes while still displaying local time for users.

## MVP Scope Constraints

The MVP MUST remain a Telegram bot service. It MUST NOT include a web UI. HTTP routes
are allowed only for Telegram webhook handling, health checks, and deployment support.
Features MUST prefer explicit command and message flows inside Telegram over additional
interfaces.

The approved core stack is Go, Telegram webhook integration, MongoDB Atlas, and the
official MongoDB Go Driver. Any additional dependency MUST have a concrete operational
or correctness reason.

## Development Workflow

Plans, specs, and tasks MUST pass the Constitution Check before implementation. Reviews
MUST verify webhook-only Telegram behavior, MongoDB Atlas persistence, UTC storage,
Europe/Kyiv display formatting, timestamp-derived elapsed time, and absence of MVP web
UI work.

Tests SHOULD focus on service logic, repository behavior, time calculations, Telegram
update handling, and user-visible date formatting. When tests are deferred, the reason
MUST be explicit in the plan or tasks.

## Governance

This constitution supersedes conflicting project guidance. Amendments MUST document the
reason for change, update dependent templates when principles affect planning or tasks,
and include a semantic version bump.

Versioning policy:
- MAJOR for removing or redefining a core principle in a backward-incompatible way.
- MINOR for adding a new principle or materially expanding governance.
- PATCH for wording, clarification, or typo-only changes.

Every feature plan MUST include a Constitution Check. Any exception MUST be documented
in Complexity Tracking with the simpler alternative considered and rejected.

**Version**: 1.0.0 | **Ratified**: 2026-06-08 | **Last Amended**: 2026-06-08
