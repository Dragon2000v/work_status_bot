# Research: Telegram Buttons And User Languages

## Decision: Use Simple In-Code Translation Dictionaries

**Rationale**: The feature supports only three languages and the project favors simple Go
code. A small `internal/i18n` package with explicit language constants, translation keys,
and nested maps keeps lookup testable and avoids external dependencies.

**Alternatives considered**:
- External translation files: useful for large localization programs, but unnecessary for
  this MVP and adds file loading/error paths.
- Third-party i18n package: more features than needed and conflicts with the project's
  preference for simple explicit code.

## Decision: Ukrainian Is The Fallback Language

**Rationale**: The feature requires Ukrainian as the default for all messages and public
group output. Language lookup should normalize unsupported or empty language codes to
`uk`.

**Alternatives considered**:
- Ask every user for language first: explicitly prohibited by the feature.
- Infer language from Telegram profile: not required, less predictable, and can conflict
  with explicit saved preferences.

## Decision: Persist User Settings In `user_settings`

**Rationale**: Language choice must survive restarts and be keyed by Telegram user id.
Use a dedicated MongoDB collection with `telegram_user_id`, `language`, `created_at`, and
`updated_at`, plus a unique index on `telegram_user_id`.

**Alternatives considered**:
- Add fields to person records: Telegram users are not the same as tracked people.
- In-memory map: loses data on restart and violates durable state requirements.

## Decision: Callback Queries Are Webhook Update Variants

**Rationale**: Telegram sends button presses as `callback_query` updates through the same
webhook route. Extending the existing update struct and handler keeps webhook-only
behavior and avoids new routes.

**Alternatives considered**:
- Polling for callback updates: prohibited.
- Separate callback endpoint: unnecessary because Telegram delivers update types to the
  configured webhook.

## Decision: Acknowledge Every Callback Query

**Rationale**: Telegram clients keep buttons in a loading state until callback queries
are answered. The handler should attempt acknowledgement for success, validation errors,
unsupported actions, and internal failures, and should log acknowledgement success or
failure.

**Alternatives considered**:
- Acknowledge only successful callbacks: causes poor UX on invalid or failed actions.
- Acknowledge after all business work only: can delay feedback; implementation should
  still ensure every callback path attempts acknowledgement.

## Decision: Use InlineKeyboardMarkup For Menus

**Rationale**: Telegram `InlineKeyboardMarkup` provides button callbacks inside messages,
matching the requested button interface. Keep keyboard construction in a small helper so
button labels can be localized and callback data stays stable.

**Alternatives considered**:
- ReplyKeyboardMarkup: visible as a chat keyboard and less suitable for action-specific
  callback routing.
- Text-only menus: does not satisfy the button requirement.

## Decision: Best-Effort Command Deletion After Processing

**Rationale**: The command result should not depend on delete permissions. The handler
should process the command first, attempt to delete only the user's command message, log
attempt/success/failure, and continue when deletion fails.

**Alternatives considered**:
- Delete before processing: risks hiding commands that fail and makes debugging harder.
- Treat deletion failure as command failure: explicitly prohibited.

## Decision: Public Group Reports And Alerts Stay Ukrainian

**Rationale**: Individual language preferences apply to user-specific interactions. Public
reports and alerts are group-wide, and group language settings are out of scope.

**Alternatives considered**:
- Use the language of the user who triggered a report: would make public output
  inconsistent for the group.
- Add group language settings now: explicitly out of scope.
