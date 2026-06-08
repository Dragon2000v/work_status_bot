# Research: Multi-Group Setup

## Decision: Add `configured_groups` Collection

**Rationale**: Allowed groups must be durable and not tied to deployment environment.
Use a dedicated collection keyed by Telegram chat id with setup metadata and enabled
state.

**Alternatives considered**:
- Reuse `telegram_group_config`: existing naming suggests a single configured group and
  would blur migration semantics.
- Store group list in environment variables: requires redeploy and fails the core goal.

## Decision: Keep `TELEGRAM_GROUP_CHAT_ID` As Fallback

**Rationale**: The current deployment and local development flows rely on the env group.
Treating it as allowed preserves backward compatibility while stored configured groups
become the normal multi-group path.

**Alternatives considered**:
- Remove env fallback immediately: creates migration risk and breaks local workflows.
- Auto-copy env fallback into MongoDB on startup: introduces implicit writes and makes
  local configuration harder to reason about.

## Decision: Authorization Service In `internal/groups`

**Rationale**: Authorization needs shared rules: enabled stored group, fallback chat,
unknown-group exceptions for `/setup` and `/help`, and disabled group rejection. Keeping
this as a small service keeps Telegram handler code explicit but avoids duplicating rules.

**Alternatives considered**:
- HTTP middleware: too early in the flow because it does not know parsed commands.
- Direct checks inside every command branch: simple initially but high duplication and
  easy to miss new commands.

## Decision: Skip Admin Verification In MVP Unless Cheap

**Rationale**: The clarified MVP allows admin verification to be skipped if documented.
Skipping avoids adding Telegram admin API calls and failure modes to the first
multi-group increment.

**Alternatives considered**:
- Always require admin verification: stronger control, but requires extra Telegram API
  integration and tests.
- Never support admin verification: too rigid for future hardening; document MVP
  limitation instead.

## Decision: Command-Driven Delivery Uses Current Chat

**Rationale**: In a multi-group bot, stop alerts and manual monthly reports must return
to the group where the user acted. A single environment group is no longer a correct
delivery target for command flows.

**Alternatives considered**:
- Keep sending to env group: breaks multi-group isolation.
- Broadcast command-driven alerts to all groups: noisy and leaks group activity.

## Decision: Cron Reports Fan Out To Enabled Groups

**Rationale**: Cron has no Telegram source chat, so it must deliver reports to all enabled
configured groups. Include fallback chat when present and not already duplicated so
backward-compatible deployments still receive reports.

**Alternatives considered**:
- Cron sends only to fallback chat: misses configured groups.
- Cron sends only to stored groups: can break legacy local/dev deployments.
