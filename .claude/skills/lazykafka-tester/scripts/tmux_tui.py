#!/usr/bin/env python3
"""Helper script for interacting with lazykafka TUI in tmux."""

import subprocess
import sys
import time
from pathlib import Path


def find_tmux():
    """Find tmux binary path."""
    for path in ["/opt/homebrew/bin/tmux", "/usr/local/bin/tmux", "tmux"]:
        try:
            result = subprocess.run(
                [path, "-V"],
                capture_output=True,
                text=True,
                shell=False if "/" in path else True,
            )
            if result.returncode == 0:
                return path
        except FileNotFoundError:
            continue
    raise RuntimeError("tmux not found")


TMUX = find_tmux()


def run_tmux(args: list[str], session: str = "0") -> str:
    """Run tmux command and return output."""
    cmd = [TMUX] + args + ["-t", session] if "-t" not in " ".join(args) else [TMUX] + args
    if "-t" not in " ".join(args) and args[0] not in ["list-sessions", "ls", "new-session", "has-session"]:
        cmd = [TMUX] + args[:1] + ["-t", session] + args[1:]
    else:
        cmd = [TMUX] + args

    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.stdout.strip()


def send_keys(keys: str, session: str = "0", enter: bool = False):
    """Send keys to tmux session."""
    args = ["send-keys", "-t", session, keys]
    if enter:
        args = ["send-keys", "-t", session, keys, "Enter"]
    subprocess.run([TMUX] + args)


def capture_pane(session: str = "0", lines: int = 50) -> str:
    """Capture tmux pane content."""
    result = subprocess.run(
        [TMUX, "capture-pane", "-t", session, "-p", f"-S", f"-{lines}"],
        capture_output=True,
        text=True,
    )
    return result.stdout


def run_app(project_path: str, session: str = "0"):
    """Run lazykafka in tmux session."""
    send_keys(f"cd {project_path} && go run cmd/lazykafka/main.go", session, enter=True)


def quit_app(session: str = "0"):
    """Quit lazykafka application."""
    send_keys("q", session)


def add_broker(name: str, bootstrap_servers: str, session: str = "0", auth: str = "none"):
    """Add a broker through the wizard."""
    # Navigate to brokers panel
    send_keys("1", session)
    time.sleep(0.2)

    # Open add broker popup
    send_keys("n", session)
    time.sleep(0.3)

    # Enter name
    send_keys(name, session)
    send_keys("Enter", session)
    time.sleep(0.2)

    # Enter bootstrap servers
    send_keys(bootstrap_servers, session)
    send_keys("Enter", session)
    time.sleep(0.2)

    # Select auth type (None is default)
    if auth == "none":
        send_keys("Enter", session)
    # TODO: Handle SASL auth


def add_topic(name: str, partitions: int, session: str = "0"):
    """Add a topic through the wizard."""
    # Navigate to topics panel
    send_keys("2", session)
    time.sleep(0.2)

    # Open add topic popup
    send_keys("n", session)
    time.sleep(0.3)

    # Enter topic name
    send_keys(name, session)
    send_keys("Enter", session)
    time.sleep(0.2)

    # Enter partitions
    send_keys(str(partitions), session)
    send_keys("Enter", session)
    time.sleep(0.2)


def navigate_panel(panel: int, session: str = "0"):
    """Navigate to specific panel (1-4)."""
    send_keys(str(panel), session)


def list_sessions() -> list[str]:
    """List all tmux sessions."""
    result = subprocess.run([TMUX, "list-sessions", "-F", "#{session_name}"], capture_output=True, text=True)
    return result.stdout.strip().split("\n") if result.stdout.strip() else []


def main():
    if len(sys.argv) < 2:
        print("Usage: tmux_tui.py <command> [args]")
        print("Commands:")
        print("  run <project_path> [session]  - Run lazykafka")
        print("  quit [session]                - Quit app")
        print("  capture [session] [lines]     - Capture pane")
        print("  add-broker <name> <servers> [session]  - Add broker")
        print("  add-topic <name> <partitions> [session] - Add topic")
        print("  panel <1-4> [session]         - Navigate to panel")
        print("  send <keys> [session]         - Send keys")
        print("  sessions                      - List sessions")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "run":
        run_app(sys.argv[2], sys.argv[3] if len(sys.argv) > 3 else "0")
    elif cmd == "quit":
        quit_app(sys.argv[2] if len(sys.argv) > 2 else "0")
    elif cmd == "capture":
        lines = int(sys.argv[3]) if len(sys.argv) > 3 else 50
        print(capture_pane(sys.argv[2] if len(sys.argv) > 2 else "0", lines))
    elif cmd == "add-broker":
        add_broker(sys.argv[2], sys.argv[3], sys.argv[4] if len(sys.argv) > 4 else "0")
    elif cmd == "add-topic":
        add_topic(sys.argv[2], int(sys.argv[3]), sys.argv[4] if len(sys.argv) > 4 else "0")
    elif cmd == "panel":
        navigate_panel(int(sys.argv[2]), sys.argv[3] if len(sys.argv) > 3 else "0")
    elif cmd == "send":
        send_keys(sys.argv[2], sys.argv[3] if len(sys.argv) > 3 else "0")
    elif cmd == "sessions":
        for s in list_sessions():
            print(s)


if __name__ == "__main__":
    main()
