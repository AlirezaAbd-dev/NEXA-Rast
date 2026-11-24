---
name: go-idiomatic-clean-code
description: >-
  Use this skill when writing, refactoring, or reviewing Go packages, interfaces,
  and error handling to enforce SOLID principles, Effective Go idioms, and clean
  package boundaries.
---

# Idiomatic Clean Go & SOLID

Follow this workflow whenever authoring or refactoring Go code in this repository.

## 1. Package & Interface Design (SOLID)

1. **Consumer-Side Interfaces (ISP / DIP)**:
   - Declare interfaces in the package that *uses* them, not the package that implements them.
   - Keep interfaces to 1–3 methods (e.g., `TargetDiscoverer`, `ScriptInjector`, `LauncherManager`).
   - Accept interfaces in constructors/functions; return concrete struct pointers from `New...()`.
2. **Shallow Nesting & Early Returns (SRP)**:
   - Handle errors and edge cases first and `return` early so the happy path stays at indentation level 1.
   - Split functions exceeding ~40 lines or mixing I/O with business logic into focused helpers.
3. **Options & Configuration**:
   - Validate configuration structs at startup (`Validate() error`) before starting loops or network connections.

## 2. Error Handling & Doc Comments

1. Wrap every propagated error with what failed:
   ```go
   if err != nil {
       return fmt.Errorf("connect to cdp target %s: %w", targetID, err)
   }
   ```
2. Use sentinel errors (`var ErrNotFound = errors.New(...)`) or typed errors when callers need to branch with `errors.Is` / `errors.As`.
3. Document every exported symbol with a `// Name ...` comment explaining its contract, not just restating its signature.

## 3. Verification

Run after every meaningful edit:

```bash
gofmt -l .
go vet ./...
```
