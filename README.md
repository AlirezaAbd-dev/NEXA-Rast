# NEXA Rast (`nexa-rast`)

**Fix RTL (Persian, Arabic, Hebrew) rendering in Antigravity 2.0 chat — fast, native, and non-destructive.**

`nexa-rast` (*Rast* = راست / "right" in Persian) is an open-source **Antigravity 2.0** plugin and lightweight **Go** sidecar. It injects a CSS-first RTL engine into the Antigravity 2.0 chat over Chrome DevTools Protocol (CDP) without modifying `app.asar`, breaking app signatures, or slowing down token streaming.

---

## Features

- **Automatic Per-Paragraph Direction**: Detects RTL vs. LTR from the first strong character (`unicode-bidi: plaintext`, `text-align: start`, `dir="auto"`).
- **Strict Code Isolation**: Keeps code blocks (`pre`, `code`), inline code, and `.monaco-editor` strictly LTR (`direction: ltr`).
- **Composer Input Support**: Typing Persian, Arabic, or Hebrew in the chat input aligns right immediately.
- **Zero-Lag Streaming**: CSS-first layout paired with a `requestAnimationFrame`-batched `MutationObserver`.
- **Safe & Reversible**: Works on read-only Linux Snap installs, survives Antigravity auto-updates, and uninstalls cleanly.

---

## Platform Support

| OS | Plugin & Sidecar | Launcher Flag (`--remote-debugging-port=9222`) |
| :--- | :--- | :--- |
| **Linux** (Snap & `.deb`/tarball) | Automatic | **Automatic** (creates user `.desktop` override in `~/.local/share/applications/`) |
| **macOS** | Automatic | **One-time launch command / alias** (see below) |
| **Windows** | Automatic | **One-time shortcut edit** (see below) |

---

## Quick Start

### Option A: Prebuilt Releases (No Go Required)

Download the latest release for your OS from [GitHub Releases](https://github.com/AlirezaAbd-dev/NEXA-Rast/releases):

- **Ubuntu / Debian (`.deb`)**:
  ```bash
  sudo dpkg -i nexa-rast_*_amd64.deb
  nexa-rast install
  ```
- **macOS (Apple Silicon `arm64` / Intel `amd64`)**:
  ```bash
  tar -xzf nexa-rast_*_darwin_arm64.tar.gz
  chmod +x nexa-rast
  ./nexa-rast install
  open -a Antigravity --args --remote-debugging-port=9222
  ```
- **Windows (`.zip` with `nexa-rast.exe`)**:
  Extract the `.zip`, open PowerShell in that folder, and run `.\nexa-rast.exe install` (then add `--remote-debugging-port=9222` to your Antigravity shortcut).

### Option B: Install via Go (Any OS with Go 1.22+)

```bash
go install github.com/AlirezaAbd-dev/NEXA-Rast/cmd/nexa-rast@latest
nexa-rast install
```

*(Ensure `$(go env GOPATH)/bin` is in your `PATH`.)*

### Prerequisites

- **Antigravity 2.0**
- **Go 1.22+** (if building from source)

---

### 1. Linux

```bash
git clone https://github.com/AlirezaAbd-dev/NEXA-Rast.git
cd NEXA-Rast
./scripts/install.sh
```

Then **fully quit and reopen Antigravity 2.0** from your application launcher.

---

### 2. macOS

```bash
git clone https://github.com/AlirezaAbd-dev/NEXA-Rast.git
cd NEXA-Rast
./scripts/install.sh
```

Quit Antigravity 2.0 (`Cmd + Q`), then launch it with the debugging port enabled:

```bash
open -a Antigravity --args --remote-debugging-port=9222
```

> **Tip:** Add an alias to `~/.zshrc` so you can launch it easily anytime:
> ```bash
> alias agy-rtl='open -a Antigravity --args --remote-debugging-port=9222'
> ```

---

### 3. Windows (PowerShell)

```powershell
git clone https://github.com/AlirezaAbd-dev/NEXA-Rast.git
cd NEXA-Rast
go build -trimpath -ldflags="-s -w" -o nexa-rast.exe .\cmd\nexa-rast
.\nexa-rast.exe install
```

Then update your Antigravity desktop/Start Menu shortcut:
1. Right-click the **Antigravity** shortcut → **Properties**.
2. Append `--remote-debugging-port=9222` to the end of **Target** (after the closing quote), e.g.:
   ```text
   "C:\Users\<YOU>\AppData\Local\Programs\Antigravity\Antigravity.exe" --remote-debugging-port=9222
   ```
3. Fully quit Antigravity (check the system tray) and reopen it using that shortcut.

---

## Verify Installation

After restarting Antigravity 2.0, run:

```bash
./bin/nexa-rast status
```

You should see:

```text
Plugin installed: true (...)
Launcher configured: true (...)
CDP reachable (port 9222): true (targets: ...)
```

You will also see **NEXA Rast** listed under **Customizations → Plugins** inside Antigravity 2.0, where you can toggle it on or off.

---

## CLI Commands

```text
nexa-rast install [--port 9222]    Install plugin & configure launcher
nexa-rast uninstall [--port 9222]  Remove plugin & restore launcher
nexa-rast status [--port 9222]     Check plugin, launcher, and CDP status
nexa-rast sidecar [--port 9222]    Run CDP injector (managed by Antigravity)
nexa-rast version                  Print version
```

If port `9222` is already used on your machine, pass `--port <port>` to `install` (and update your launch flag to match).

---

## How It Works

```text
Antigravity 2.0 (--remote-debugging-port=9222)
   │
   ├── starts sidecar ──▶ nexa-rast sidecar
   │                         │
   │                         ├── polls http://127.0.0.1:9222/json
   │                         └── injects rtl.css + rtl.js via CDP WebSocket
   │
   └── Chat Canvas DOM ──▶ CSS unicode-bidi: plaintext + dir="auto" (code stays LTR)
```

1. **Plugin (`~/.gemini/config/plugins/nexa-rast/`)**: Registers the `rtl-injector` sidecar so Antigravity 2.0 starts and supervises it automatically.
2. **Go Sidecar (`internal/cdp`)**: Discovers active renderer pages on `127.0.0.1:9222`, injects the payload once per target, and idles with near-zero CPU.
3. **RTL Engine (`internal/assets`)**: Applies native CSS bidi rules and a minimal `MutationObserver` for streamed markdown and composer inputs.

---

## Troubleshooting

- **`CDP reachable (port 9222): false`**:
  Antigravity was still running in the background when you restarted it. Quit it completely (including any background/tray process) and launch it again.
- **Port conflict**:
  Run `nexa-rast install --port 9229` (or any free port) and restart Antigravity with `--remote-debugging-port=9229`.
- **Clean uninstall**:
  Run `nexa-rast uninstall` and restart Antigravity. All plugin files and `.desktop` overrides are removed cleanly.

---

## Contributing

Contributions are welcome! Please read [AGENTS.md](AGENTS.md) for the project's Go architecture, SOLID guidelines, and testing standards, then run:

```bash
gofmt -l .
go vet ./...
go test -race ./...
```

## License

MIT — see [LICENSE](LICENSE).
