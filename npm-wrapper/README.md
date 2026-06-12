# another-meet

**Manage Google Meet meetings from your terminal.** 🚀

`another-meet` is a command-line tool for creating, joining, scheduling, and inviting people to Google Meet meetings — without ever leaving your terminal. V1 focuses entirely on **Google Meet** integration via the Google Calendar API.

## Installation (Node.js)

You can install `another-meet` globally using your preferred Node.js package manager:

```bash
# npm
npm install -g another-meet

# yarn
yarn global add another-meet

# pnpm
pnpm add -g another-meet

# bun
bun install -g another-meet
```

*(Note: This package automatically downloads the pre-compiled standalone Go binary for your operating system during installation).*

## Quick Start

```bash
# 1. Authenticate with Google
another-meet auth login

# 2. Create an instant meeting with a Google Meet link
another-meet create

# 3. List today's meetings
another-meet list

# 4. Join the next upcoming meeting
another-meet join
```

## Commands

### `auth` — Manage authentication

```bash
# Login via browser (OAuth2 + PKCE)
another-meet auth login

# Check who you're authenticated as
another-meet auth status

# Remove stored credentials
another-meet auth logout
```

### `create` — Create an instant meeting

```bash
# Create a quick 30-min meeting with a Meet link
another-meet create

# Custom title and duration
another-meet create --title "Sprint Planning" --duration 1h

# With attendees and auto-open in browser
another-meet create -t "Design Review" -d 45m -a "alice@co.com,bob@co.com" --open
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | Quick Meeting | Meeting title |
| `--duration` | `-d` | 30m | Duration (30m, 1h, 1h30m) |
| `--attendees` | `-a` | — | Comma-separated emails |
| `--description` | | — | Meeting description |
| `--open` | `-o` | false | Open Meet link in browser |
| `--calendar` | | primary | Calendar ID |
| `--no-meet` | | false | Skip Meet link creation |

### `list` — List upcoming meetings

```bash
# Today's meetings
another-meet list

# Specific date range
another-meet list --from "2026-06-10" --to "2026-06-12"

# Only meetings with Meet links
another-meet list --has-meet

# JSON output for scripting
another-meet list --json | jq '.[].meet_link'
```

| Flag | Default | Description |
|------|---------|-------------|
| `--from` | today | Start date (e.g. "today", "2026-06-10") |
| `--to` | today | End date |
| `--has-meet` | false | Only list events that have a Google Meet link |
| `--calendar` | primary | Calendar ID to fetch events from |

### `schedule` — Schedule a future meeting

```bash
# Schedule for a specific date and time
another-meet schedule --title "Design Review" --at "2026-06-10 14:00" --duration 45m

# Schedule for tomorrow with attendees
another-meet schedule -t "Weekly Sync" --at "tomorrow 10:00" -d 30m -a "team@co.com"
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | Quick Meeting | Meeting title |
| `--at` | | — | When the meeting starts (e.g. "tomorrow 10:00", "2026-06-10 14:00") |
| `--duration` | `-d` | 30m | Duration (30m, 1h, 1h30m) |
| `--attendees` | `-a` | — | Comma-separated emails |
| `--description` | | — | Meeting description |
| `--calendar` | | primary | Calendar ID |

### `join` — Join a meeting

```bash
# Join the next upcoming meeting
another-meet join

# Join a specific meeting by event ID
another-meet join --id <event-id>
```

| Flag | Description |
|------|-------------|
| `--id` | Specific event ID to join (if not provided, joins the next upcoming meeting) |

### `invite` — Add attendees to a meeting

```bash
# Invite to a specific meeting
another-meet invite --id <event-id> -a "charlie@co.com"

# Invite to the next upcoming meeting
another-meet invite --next -a "charlie@co.com,dave@co.com"
```

| Flag | Short | Description |
|------|-------|-------------|
| `--id` | | Specific event ID to invite to |
| `--next` | | Invite to the next upcoming meeting |
| `--attendees` | `-a` | Comma-separated emails to invite |

### For Full Documentation
For the complete list of command-line flags, configuration details, and contribution guidelines, please visit the [official GitHub repository](https://github.com/parjanyaacoder/another-meet).
