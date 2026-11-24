# Project Rules — antigravity-rtl

These rules apply to all code, scripts, and assets in `antigravity-rtl`.

## 1. Folder Structure (Feature-Based Go Layout)

Organize code by **feature/domain** under `internal/`, not by technical layers (`models/`, `services/`, `utils/`). Keep the hierarchy shallow and idiomatic:

```text
antigravity-rtl/
├── cmd/antigravity-rtl/   # CLI entrypoint & command wiring only
├── internal/
│   ├── cdp/               # Feature: CDP target discovery, WS transport, injection (+ *_test.go)
│   ├── installer/         # Feature: Plugin install/uninstall & OS launcher setup (+ *_test.go)
│   └── assets/            # Feature: Embedded CSS/JS RTL payload (//go:embed)
├── plugin/                # Antigravity plugin bundle (plugin.json, sidecars/)
├── scripts/               # End-user install/release scripts
├── go.mod
├── LICENSE
└── README.md
```

- **Colocate by feature**: Keep each feature's types, consumer interfaces, implementation, and `*_test.go` files in its own `internal/<feature>/` package.
- **No `utils/` or `common/` buckets**: Never create generic `utils`, `helpers`, or `models` packages. Shared contracts belong with the consumer; small feature-specific helpers stay unexported inside their feature package.
- **Keep `cmd/` thin**: `cmd/antigravity-rtl/main.go` parses flags/subcommands, wires dependencies, and calls into `internal/` packages—no business logic in `main`.

## 2. Architecture & SOLID Principles (Go Style)

- **Single Responsibility (SRP)**: Keep packages, types, and functions focused on one job.
  - `cmd/antigravity-rtl` / `main.go`: CLI argument parsing and wiring only.
  - `internal/cdp`: Chrome DevTools Protocol target discovery, WebSocket transport, and script injection.
  - `internal/installer`: Plugin installation and OS-specific launcher configuration.
  - `internal/assets`: Embedded CSS/JS payload (`//go:embed`).
- **Open/Closed (OCP)**: Model OS-specific behavior (Linux `.desktop`, macOS, Windows) and transport layers behind small interfaces so new platforms or injectors can be added without modifying core orchestration.
- **Liskov Substitution (LSP)**: Implementations must honor interface contracts, respect `context.Context` cancellation, and return documented error types without unexpected panics.
- **Interface Segregation (ISP)**: Define small, consumer-side interfaces (1–3 methods). Follow *"Accept interfaces, return structs."*
- **Dependency Inversion (DIP)**: High-level workflows (installer, sidecar loop) must depend on interfaces (e.g., filesystem/launcher/CDP client abstractions) rather than concrete OS or network calls, keeping business logic 100% unit-testable without a live Electron process.

## 3. Clean Code & Idiomatic Go

- Follow *Effective Go* and *Go Code Review Comments*:
  - Short, descriptive names; avoid `Get` prefixes on getters; keep receiver names consistent (1–2 letters).
  - Prefer shallow nesting and early returns over deep `if/else` chains.
  - Keep functions small and single-purpose.
- **Error Handling**:
  - Never ignore returned errors (`_ = ...` is only allowed when explicitly justified, e.g., deferred `Close()` after checking writes).
  - Wrap errors with context using `fmt.Errorf("action: %w", err)`.
  - Never use `panic` for expected runtime or I/O failures.
- **Dependencies**:
  - Prefer the Go standard library (`net/http`, `encoding/json`, `embed`, `flag`, `os/exec`, `testing`).
  - Keep external dependencies minimal and well-vetted.

## 4. Performance, Concurrency & Resource Safety

- Pass `context.Context` as the first parameter to all I/O, network, and long-running functions.
- **No Goroutine Leaks**: Every spawned goroutine must have a clear exit path tied to `ctx.Done()` or a closed channel. Handle `SIGINT`/`SIGTERM` cleanly.
- **No Timer Leaks**: Avoid `time.After` inside `for/select` loops; use `time.NewTicker` (with `defer ticker.Stop()`) or `time.NewTimer`.
- **Low Sidecar Overhead**: The CDP sidecar must idle near 0% CPU and minimal RSS when connected, waking only on target changes, disconnects, or page navigations.
- **CSS-First Frontend Performance**:
  - Use native CSS (`unicode-bidi: plaintext`, `text-align: start`, `:dir(rtl)`, and `dir="ltr"` isolation on `pre`/`code`) as the primary RTL engine.
  - Keep `MutationObserver` JS minimal, scoped to newly added nodes, and batched with `requestAnimationFrame` so token streaming never stutters.

## 5. Verification & Documentation

- Always run `gofmt -l .`, `go vet ./...`, and `go test -race ./...` after code changes.
- Write table-driven unit tests for all non-trivial logic.
- Add concise Go doc comments (`// SymbolName ...`) on all exported types, functions, and constants.
