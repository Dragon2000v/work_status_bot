# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go [version or NEEDS CLARIFICATION]

**Primary Dependencies**: Telegram webhook handling, official MongoDB Go Driver

**Storage**: MongoDB Atlas with UTC timestamps

**Testing**: Go tests (`go test ./...`) or NEEDS CLARIFICATION

**Target Platform**: Server-hosted Telegram webhook service

**Project Type**: Go Telegram bot service

**Performance Goals**: [domain-specific webhook latency/throughput or NEEDS CLARIFICATION]

**Constraints**: No Telegram polling; no in-memory timers; no background infinite loops; no MVP web UI

**Scale/Scope**: [expected chats/users/status records or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Clean architecture boundaries are explicit: Telegram delivery, services, persistence.
- Go code plan stays idiomatic, readable, and explicit without unnecessary abstractions.
- Telegram integration uses webhook only; no polling, no background infinite loops.
- No in-memory timers; elapsed time derives from persisted UTC timestamps.
- MongoDB Atlas uses the official MongoDB Go Driver.
- Database timestamps are UTC; user-facing dates display in Europe/Kyiv timezone.
- MVP scope has no web UI.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Go Telegram bot service (DEFAULT)
cmd/bot/
internal/
├── config/
├── telegram/
├── service/
├── repository/
└── timefmt/

tests/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
