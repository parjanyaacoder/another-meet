# another-meet

**Manage Google Meet meetings from your terminal.** 🚀

`another-meet` is a command-line tool for creating, joining, scheduling, and inviting people to Google Meet meetings — without ever leaving your terminal. V1 focuses entirely on **Google Meet** integration via the Google Calendar API.

```
$ another-meet create -t "Sprint Planning" -d 1h -a "alice@co.com,bob@co.com"

✓ Meeting created!

  Title:  Sprint Planning
  Meet:   https://meet.google.com/abc-defg-hij
  Time:   10:30 AM — 11:30 AM IST
  Invited: alice@co.com, bob@co.com
```

---

## Installation

### Homebrew (macOS / Linux)

```bash
# Option 1: Direct install
brew install parjanyaacoder/another-meet/another-meet

# Option 2: Add tap first, then install
brew tap parjanyaacoder/another-meet
brew install another-meet
```

### Go Install

```bash
go install github.com/parjanyaacoder/another-meet@latest
```

### NPM / Node.js

```bash
npm install -g another-meet
# or
bun install -g another-meet
```

### Pip / Python

```bash
pip install another-meet
# or
uv pip install another-meet
```

### Binary Download

Download the latest release from the [Releases](https://github.com/parjanyaacoder/another-meet/releases) page.

```bash
# macOS / Linux
tar -xzf another-meet_*_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m).tar.gz
sudo mv another-meet /usr/local/bin/
```

---

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

---

## Commands

### `auth` — Manage authentication

```bash
# Login via browser (OAuth2 + PKCE)
another-meet auth login

# Login from SSH / headless environment
another-meet auth login --headless

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

### `schedule` — Schedule a future meeting

```bash
# Schedule for a specific date and time
another-meet schedule --title "Design Review" --at "2026-06-10 14:00" --duration 45m

# Schedule for tomorrow with attendees
another-meet schedule -t "Weekly Sync" --at "tomorrow 10:00" -d 30m -a "team@co.com"
```

### `join` — Join a meeting

```bash
# Join the next upcoming meeting
another-meet join

# Join a specific meeting by event ID
another-meet join --id <event-id>
```

### `invite` — Add attendees to a meeting

```bash
# Invite to a specific meeting
another-meet invite --id <event-id> -a "charlie@co.com"

# Invite to the next upcoming meeting
another-meet invite --next -a "charlie@co.com,dave@co.com"
```

### `version` — Print version info

```bash
another-meet version
```

---

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output structured JSON to stdout |
| `--no-color` | Disable colored output |
| `--config` | Custom config file path |
| `-v, --verbose` | Verbose output |

---

## Configuration

Configuration is stored at `~/.another-meet/config.yaml`:

```yaml
default_calendar: primary
timezone: Asia/Kolkata
default_duration: 30m
open_browser: false
```

Tokens are stored at `~/.another-meet/token.json` (permissions `0600`).

---

## Prerequisites

1. **Google Cloud Project** with the [Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com) enabled
2. **OAuth 2.0 credentials** — Desktop Application type, downloaded as JSON
3. Run `another-meet auth login` to authenticate

See the [Google Calendar API Quickstart](https://developers.google.com/calendar/api/quickstart/go) for setup details.

---

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m "feat: add my feature"`)
4. Push and open a Pull Request

```bash
go vet ./...
go test ./...
```

---

## License

[MIT License](LICENSE) © 2026 parjanyaacoder
