# Quickstart: Multi-Group Setup

## Prerequisites

- Go 1.22+
- Valid local `.env`
- MongoDB Atlas connection configured
- Telegram bot configured to receive webhook updates

## Static Validation

```sh
go test ./...
```

Expected result:
- Existing tests pass.
- New tests cover setup command in a group, private chat setup rejection, idempotent
  setup, missing or blank group title normalization to `Unknown group`, unknown group
  rejection, unknown group `/setup` allowed, fallback env group allowed, `/groups`
  allowed from enabled configured and fallback groups, `/groups` rejected from disabled,
  unknown, and private chats, multiple groups, disable group, callback query
  acknowledgement, cron fan-out, and current-chat alerts.

## Manual Validation

### 1. Setup In A New Group

Add the bot to a new Telegram group and run:

```text
/setup
```

Expected result:
- Bot confirms setup.
- Group is stored in `configured_groups` with `enabled=true`.
- Existing commands such as `/status` are accepted from this group without redeploy.

### 2. Private Chat Setup Rejected

Open a private chat with the bot and run:

```text
/setup
```

Expected result:
- Bot rejects setup with a clear group-only message.

### 3. Unknown Group Rejection

In another group that has not run setup, run:

```text
/status
```

Expected result:
- Bot rejects the command with setup-required guidance.
- `/help` and `/setup` remain available.

### 4. List And Disable Groups

From a configured group, run:

```text
/groups
/disable_group
/status
```

Expected result:
- `/groups` lists all stored configured groups, including disabled groups, with enabled
  status.
- `/disable_group` marks the current group disabled.
- `/groups` and `/status` are rejected after disable unless the group is allowed by
  fallback config.

### 5. Delivery Targets

Trigger `/stop_work` and `/report_month` from a configured group.

Expected result:
- Stop alert is sent to that same group.
- Manual report is sent or returned to that same group.

### 6. Cron Fan-Out

Configure two enabled groups and trigger:

```text
POST /cron/monthly-report
```

Expected result:
- Report send is attempted for each enabled configured group.
- Fallback `TELEGRAM_GROUP_CHAT_ID` also receives a send attempt if configured and not
  already one of the enabled configured groups.

### 7. Lightweight Timing Validation

Expected result for MVP-sized data:
- `/setup` can be completed manually in under 1 minute.
- Normal command handlers complete in under 5 seconds.
- `POST /cron/monthly-report` completes in under 5 seconds.
