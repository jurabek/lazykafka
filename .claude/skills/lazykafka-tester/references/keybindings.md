# Lazykafka Keybindings Reference

## Global Navigation

| Key | Action |
|-----|--------|
| `1` | Jump to Brokers panel |
| `2` | Jump to Topics panel |
| `3` | Jump to Consumer Groups panel |
| `4` | Jump to Schema Registry panel |
| `←`/`h` | Previous panel |
| `→`/`l` | Next panel |
| `q` | Quit application |
| `Ctrl+C` | Force quit |

## List Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `g` `g` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+d` | Page down |
| `Ctrl+u` | Page up |

## Brokers Panel (1)

| Key | Action |
|-----|--------|
| `n` | Add new broker |
| `Enter` | Connect to selected broker |

## Topics Panel (2)

| Key | Action |
|-----|--------|
| `n` | Create new topic |
| `p` | Produce message to selected topic |
| `d` | Delete selected topic |
| `e` | Edit topic config |
| `Enter` | View topic details |

## Topic Detail Tabs

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `1` | Messages tab |
| `2` | Partitions tab |
| `3` | Config tab |

## Consumer Groups Panel (3)

| Key | Action |
|-----|--------|
| `Enter` | View consumer group details |

## Popup Forms

| Key | Action |
|-----|--------|
| `Enter` | Confirm/Next field |
| `Esc` | Cancel/Close |
| `↑`/`↓` | Select option in lists |

## Add Broker Wizard Steps

1. **Name**: Broker display name
2. **Bootstrap Servers**: `host:port` (comma-separated)
3. **Auth Type**: None / SASL
4. **SASL Mechanism** (if SASL): PLAIN / SCRAM-SHA-256 / SCRAM-SHA-512 / OAUTHBEARER
5. **Username** (if SASL)
6. **Password** (if SASL)

## Add Topic Wizard Steps

1. **Topic Name**
2. **Partitions** (number)
3. **Replication Factor** (number)
4. **Cleanup Policy**: delete / compact / compact,delete
5. **Min In-Sync Replicas**
6. **Retention (ms)**
