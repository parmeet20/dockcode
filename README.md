# DockCode

AI-powered Docker management right inside your terminal.

DockCode is a premium, modern Terminal User Interface (TUI) that integrates AI agents
with your local Docker daemon. Monitor, debug, inspect, and manage containers, images,
volumes, and networks using natural language — no GUI required.

---

## Quick Install

Make sure Go 1.21+ is installed, then run:

```
go install github.com/parmeet20/dockcode@latest
```

**That's it.** On first launch, DockCode automatically and permanently adds
`%USERPROFILE%\go\bin` (Windows) or `~/go/bin` (macOS/Linux) to your system
`PATH` — so typing `dockcode` in any new terminal works from that point forward.

> **Windows users (CMD or PowerShell)**
> After running `go install`, open a **new** CMD or PowerShell window and type:
> ```
> dockcode
> ```
> The first-run PATH setup writes directly to the Windows Registry
> (`HKEY_CURRENT_USER\Environment`) and broadcasts a system-wide change
> notification, so no logout or reboot is needed.

> **macOS / Linux users**
> The first run appends `export PATH="$PATH:~/go/bin"` to your
> `.zshrc` / `.bashrc` / `.profile`. Open a new terminal tab and type `dockcode`.

---

## Features

- **Natural Language AI** — Ask the agent to inspect logs, stop containers, pull images, prune resources, and more.
- **Dynamic Sidebar** — Live view of containers, images, volumes, and networks, refreshed in the background.
- **Session Browser** — Full session history with search. Press `/sessions` to open, `Enter` to switch, `q` to go back.
- **Rich Slash Commands** — `/newchat`, `/sessions`, `/logs`, `/stop`, `/rm`, `/theme`, and many more.
- **ASCII Fallbacks** — Works in classic CMD/PowerShell (braille spinners replaced with `-\|/`, emojis replaced with clean ASCII symbols).
- **Copy / Paste** — Native terminal copy-paste works in every supported terminal.

---

## Slash Commands

| Command | Description |
|---|---|
| `/help` | Show all available commands |
| `/exit` | Gracefully exit DockCode |
| `/clear` | Clear chat and reset session memory |
| `/newchat` | Start a new chat session |
| `/sessions` | Open session browser (switch / search past sessions) |
| `/session rename <title>` | Rename current session |
| `/session delete` | Delete current session |
| `/session export` | Export chat to a Markdown file |
| `/session tag <tag>` | Tag current session |
| `/settoken <token>` | Set a new API token |
| `/seturl <url>` | Set a new API base URL |
| `/model <name>` | Switch the active LLM model |
| `/models` | List available models |
| `/config` | Show current configuration |
| `/theme` | Toggle dark / light theme |
| `/containers` | Focus containers panel |
| `/images` | Focus images panel |
| `/volumes` | Focus volumes panel |
| `/networks` | Focus networks panel |
| `/logs <name>` | Stream container logs |
| `/stop <name>` | Stop a running container |
| `/rm <name>` | Remove a container |

---

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `Tab` | Cycle sidebar panels |
| `↑ / ↓` | Scroll chat or navigate autocomplete |
| `/` | Open command autocomplete |
| `Enter` | Send message / confirm command |
| `q` | Go back from session browser |
| `Esc` | Dismiss autocomplete or overlays |
| `Ctrl+C` | Quit DockCode |

---

## Configuration

Config is stored at `~/.dockcode/config.json`. It is created automatically
during the onboarding wizard on first launch.

Sessions are stored under `~/.dockcode/sessions/`.
