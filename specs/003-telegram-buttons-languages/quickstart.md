# Quickstart: Telegram Buttons And User Languages

## Prerequisites

- Go 1.22+
- Valid local `.env` for the existing Telegram webhook bot
- MongoDB Atlas connection configured
- Telegram bot configured to use webhook mode

## Static Validation

```sh
go test ./...
```

Expected result:
- All existing tests pass.
- New tests cover default Ukrainian language, saved language lookup, changing language,
  callback routing, menu rendering, localized button labels, localized command responses,
  best-effort command deletion, and fallback commands.

## Manual Webhook Validation

Run the bot locally or in the existing deployment environment using the project's normal
startup command.

### 1. Default Ukrainian Menu

Send `/help` or the command that opens the main menu in the configured Telegram group.

Expected result:
- The bot displays Ukrainian text by default.
- Inline buttons include Add person, Start work, Status, Stop work, Monthly report,
  Settings, and Help in Ukrainian.
- The user's command message is deleted if the bot has permission.
- If deletion permission is missing, the command still works and logs a delete failure.

### 2. Callback Acknowledgement

Press each main menu button.

Expected result:
- Each button press is acknowledged.
- Logs include callback received, callback handled, and callback answer success/failure.
- Unsupported or stale callback data is acknowledged and returns a localized error.

### 3. Change User Language

Open Settings, choose language, then select English.

Expected result:
- The language menu shows Ukrainian, English, and Russian.
- The user setting is saved in `user_settings`.
- Future user-specific responses for that Telegram user use English.
- Another user without a saved language still sees Ukrainian.

### 4. Russian User-Specific Flow

With a second Telegram user, open Settings and select Russian.

Expected result:
- That user's Settings, Help, validation prompts, and button labels use Russian.
- The first user's language remains unchanged.

### 5. Public Group Output Remains Ukrainian

Trigger a status report, stop alert, incident output, or monthly report from a user with
English or Russian selected.

Expected result:
- Public group alerts and reports are Ukrainian.
- No group-level language setting appears.

### 6. Fallback Commands Still Work

Run existing text commands:

```text
/add_person Ivan Petrenko
/start_work Ivan Petrenko "API fix"
/status
/stop_work Ivan Petrenko Done
/report_month
```

Expected result:
- Commands preserve existing business behavior.
- Command responses are localized for user-specific interactions.
- Public alerts and monthly reports remain Ukrainian.
- Only user command messages are deletion candidates; bot responses, reports, alerts, and
  monthly reports remain visible.
