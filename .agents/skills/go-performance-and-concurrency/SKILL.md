---
name: go-performance-and-concurrency
description: >-
  Use this skill when implementing goroutines, background sidecar loops,
  WebSocket/HTTP clients, or optimizing allocations and CPU/memory usage in Go.
---

# Go Performance & Concurrency

Keep the `antigravity-rtl` sidecar fast, deterministic, and leak-free.

## 1. Lifecycle & Goroutine Safety

1. **Context Everywhere**: Pass `ctx context.Context` into every blocking or network call and select on `<-ctx.Done()` in every loop.
2. **Signal Handling**: In `main`, wrap the root context with `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`.
3. **Timers in Loops**: Never call `time.After` inside a `for` loop. Use `ticker := time.NewTicker(...)` with `defer ticker.Stop()`.
4. **HTTP Client Reuse**: Create a single `*http.Client` with an explicit `Timeout` (e.g., `3 * time.Second`) and reuse it across polls instead of creating a new client per request.

## 2. Memory & Allocation Discipline

1. Embed static CSS/JS assets at compile time using `//go:embed` so there is zero runtime disk I/O in the hot path.
2. Preallocate slices when the capacity is known (`make([]T, 0, n)`).
3. Close HTTP response bodies (`defer resp.Body.Close()`) and WebSocket connections on every exit path.

## 3. Verification & Benchmarks

Run benchmarks and race checks when modifying hot paths or concurrent loops:

```bash
go test -race ./...
go test -bench . -benchmem ./...
```
