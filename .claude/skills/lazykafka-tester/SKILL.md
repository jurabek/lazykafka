---
name: lazykafka-tester
description: "Test lazykafka TUI application in tmux. Use when: running the app for testing, sending key commands to the TUI, adding brokers/topics via wizard forms, capturing screenshots, or automating TUI interactions. Triggers on: test lazykafka, run lazykafka in tmux, test the TUI app, add broker in lazykafka."
---

# Lazykafka Tester

Test the lazykafka TUI application using tmux for automation.

## Prerequisites

- Docker compose running: `docker-compose up -d` (Kafka broker)
- Existing tmux session

## Quick Start

```bash
# List sessions
tmux list-sessions

# Run app in session
tmux send-keys -t 0 'go run cmd/lazykafka/main.go' Enter

# Capture screen
tmux capture-pane -t 0 -p | tail -30

# Quit app
tmux send-keys -t 0 q
```

## Tmux Helper Script

Use `scripts/tmux_tui.py` for common operations:

```bash
# Run app
python scripts/tmux_tui.py run /path/to/lazykafka

# Add broker (name, bootstrap_servers)
python scripts/tmux_tui.py add-broker local localhost:9092

# Add topic (name, partitions)
python scripts/tmux_tui.py add-topic test-topic 3

# Navigate to panel (1-4)
python scripts/tmux_tui.py panel 2

# Capture screen
python scripts/tmux_tui.py capture 0 50

# Quit app
python scripts/tmux_tui.py quit
```

## Direct Tmux Commands

### Navigation

```bash
# Send key to session
tmux send-keys -t 0 '<key>'

# Panel navigation (1=Brokers, 2=Topics, 3=Consumer Groups, 4=Schema Registry)
tmux send-keys -t 0 '1'
tmux send-keys -t 0 '2'

# List navigation
tmux send-keys -t 0 'j'    # down
tmux send-keys -t 0 'k'    # up
tmux send-keys -t 0 'g'    # first 'g'
tmux send-keys -t 0 'g'    # second 'g' = jump to top
tmux send-keys -t 0 'G'    # jump to bottom
```

### Add Broker Flow

```bash
# 1. Go to brokers panel
tmux send-keys -t 0 '1'

# 2. Open add broker wizard
tmux send-keys -t 0 'n'

# 3. Fill name
tmux send-keys -t 0 'my-broker'
tmux send-keys -t 0 Enter

# 4. Fill bootstrap servers
tmux send-keys -t 0 'localhost:9092'
tmux send-keys -t 0 Enter

# 5. Select auth (None is default)
tmux send-keys -t 0 Enter
```

### Add Topic Flow

```bash
# 1. Go to topics panel
tmux send-keys -t 0 '2'

# 2. Open add topic wizard
tmux send-keys -t 0 'n'

# 3. Fill topic name
tmux send-keys -t 0 'test-topic'
tmux send-keys -t 0 Enter

# 4. Fill partitions
tmux send-keys -t 0 '3'
tmux send-keys -t 0 Enter

# 5. Fill replication factor
tmux send-keys -t 0 '1'
tmux send-keys -t 0 Enter

# 6. Cleanup policy (arrow keys to select)
tmux send-keys -t 0 Enter

# Continue through remaining fields...
```

## Keybindings Reference

See [references/keybindings.md](references/keybindings.md) for complete keybindings.

### Essential Keys

| Key | Action |
|-----|--------|
| `1-4` | Jump to panel |
| `n` | New (broker/topic) |
| `q` | Quit |
| `Enter` | Confirm/Select |
| `Esc` | Cancel/Close |
| `↑↓` | Navigate options |

## Testing Workflow

1. Ensure Kafka running: `docker-compose up -d`
2. Start app in tmux session
3. Add broker with connection details
4. Navigate panels with number keys
5. Test operations (add topic, produce message)
6. Verify UI with `capture-pane`
7. Quit with `q`
