---
name: go-table-driven-testing
description: >-
  Use this skill when writing or updating unit tests in Go. Enforces
  table-driven tests, dependency injection via small interfaces, subtests with
  t.Run, and race-detector verification without external mocking frameworks.
---

# Go Table-Driven Testing

Use the standard `testing` package with table-driven subtests and lightweight fakes.

## 1. Structure of a Test

1. Define a slice of anonymous test cases (`tests := []struct { name string; ... }`).
2. Iterate with `for _, tc := range tests` and call `t.Run(tc.name, func(t *testing.T) { ... })`.
3. Use `t.Helper()` in assertion and setup helpers so failures report the caller's line number.
4. Use `t.TempDir()` for any filesystem tests (e.g., plugin installation, `.desktop` file generation) — never write to the user's real `$HOME` in unit tests.

## 2. Testing I/O & Network Boundaries

1. **HTTP / CDP**: Use `net/http/httptest.NewServer` to test target discovery (`/json`) and error handling deterministically.
2. **OS / Launcher**: Inject a small interface or temp-dir path provider rather than calling real system commands in unit tests.
3. **Errors**: Assert error conditions with `errors.Is(err, wantErr)` rather than brittle string matching whenever possible.

## 3. Verification

Run all tests with the race detector enabled:

```bash
go test -race -v ./...
```
