# Research: Telegram Reply Keyboard, Languages, And Start Work Flow

## Decision: Keep existing Go service boundaries and add two small packages

**Decision**: Use existing `internal/telegram`, `internal/users`, `internal/i18n`,
`internal/groups`, `internal/people`, `internal/works`, and `internal/reports`.
Add `internal/flows` for persisted multi-step state and `internal/datetime` for
Europe/Kyiv parsing.

**Rationale**: The repository already has clear Telegram, domain, and persistence
boundaries. Adding small focused packages avoids mixing flow state and date parsing into
the webhook handler while keeping the implementation simple.

**Alternatives considered**:
- Put all flow logic in `internal/telegram`: rejected because it would make the handler
  stateful and harder to test.
- Add a workflow framework: rejected as overengineering for short Telegram flows.

## Decision: ReplyKeyboardMarkup replaces only the main menu

**Decision**: Use Telegram `ReplyKeyboardMarkup` for main menu and Settings flow buttons.
Do not use `InlineKeyboardMarkup` for main-menu navigation. Preserve callback query
handling for any existing non-main-menu callback behavior and regression coverage.

**Rationale**: The requested UX is persistent large buttons near the input field. Existing
callback tests and callback support should not regress if callbacks remain elsewhere.

**Alternatives considered**:
- Remove callback handling entirely: rejected because it would risk regressions in
  existing webhook behavior.
- Keep inline main menu: rejected because it conflicts with the feature requirement.

## Decision: Register setMyCommands on startup through Telegram client

**Decision**: Add `SetMyCommands` support to the Telegram client and call it during
application startup after client construction and before serving HTTP requests. Log
attempt, success, and failure without token or secret values. If registration fails,
startup continues and the bot still handles slash commands and reply keyboard actions.

**Rationale**: Startup registration makes the native command menu deterministic and easy
to verify. Keeping it in the client boundary matches existing `SetWebhook` behavior.

**Alternatives considered**:
- Register commands lazily on first webhook: rejected because first user interaction
  should not carry setup work or hide registration failures.
- Manual operator registration: rejected because it is not testable as application
  behavior.

## Decision: Persist user settings in user_settings

**Decision**: Use `user_settings` keyed by `telegram_user_id`, with `language`,
`created_at`, and `updated_at`. Unknown or unsupported languages normalize to Ukrainian.

**Rationale**: The existing `internal/users` package and collection already match this
need. Per-user language must survive restarts and must not depend on chat state.

**Alternatives considered**:
- Store language in flow state: rejected because language is a durable preference.
- Store group language: rejected as out of scope.

## Decision: Persist flow state in user_flow_states

**Decision**: Use a MongoDB-backed flow state record keyed by user/chat with
`flow_type`, `step`, JSON-like `payload`, and UTC expiration timestamps. Flow states
expire 15 minutes after `updated_at`. Expiration is checked when handling inbound
messages and stale states are cleared or superseded before returning a localized
expired-flow message.

**Rationale**: The bot must know what a user's next text message means after restarts,
without in-memory timers or loops. Timestamp checks on inbound webhooks satisfy the
constitution and keep behavior simple.

**Alternatives considered**:
- In-memory maps: rejected because state is lost on restart and violates the request.
- Background cleanup loop: rejected because background infinite loops and timers are
  banned. MongoDB TTL index may be used for eventual cleanup, but correctness cannot
  depend on a local loop.

## Decision: Date/time parser is a dedicated service

**Decision**: Add `internal/datetime` to parse `DD.MM HH:mm`, `DD.MM.YYYY HH:mm`,
`YYYY-MM-DD HH:mm`, and `HH:mm` for Today/Yesterday. Parsing uses Europe/Kyiv; accepted
instants are returned in UTC; future values are rejected.

**Rationale**: Date parsing has enough validation rules to deserve isolated tests. The
existing `works.KyivLocation` behavior can be reused or centralized to keep display and
parse rules consistent.

**Alternatives considered**:
- Parse directly in Telegram handler: rejected because format and future-date tests would
  be harder to isolate.
- Accept free-form natural language dates: rejected as too broad for MVP.

## Decision: Best-effort delete applies to user input only

**Decision**: Attempt `deleteMessage` after slash commands, reply-keyboard action text,
and manual flow input. Never target bot responses, alerts, reports, incidents, status
output, or monthly reports. Deletion failure logs and does not alter command/flow result.

**Rationale**: This keeps group chats readable without making Telegram admin permissions
a functional dependency.

**Alternatives considered**:
- Delete only slash commands: rejected because the reply keyboard sends text messages
  that create the same group-chat noise.
- Treat deletion failure as an error: rejected because bots may not have admin rights.
