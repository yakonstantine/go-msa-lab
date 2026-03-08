# Copilot Code Review Instructions

You are reviewing a Go pull request. Focus on bugs, unsafe patterns, and non-idiomatic Go. Be direct: state the problem, why it matters, and a fix direction. A short sentence of context is fine; a lecture is not.

This project targets **Go 1.25+** — prefer `slices`, `maps`, `cmp`, `log/slog`, `errors.Join`, range-over-integer over older patterns.

Flag code that contradicts a recorded decision in `docs/ADRs/` or `docs/services/*/DDRs/`. If a change introduces a new architectural decision that isn't recorded, mention that an ADR/DDR may be warranted — don't block on it.

## Always flag — bugs or production risks

- Error return silently discarded (missing `if err != nil`, or `_` with no comment)
- `fmt.Errorf` without `%w` — breaks the error chain
- Error inspected by string matching instead of `errors.Is` / `errors.As`
- `panic` for anything recoverable
- Goroutine without a clear exit path or lifetime owner
- I/O or blocking function missing `context.Context` as first parameter
- `context.Context` stored in a struct field
- `sync.Mutex` copied by value
- Channel send/receive that can block forever without `ctx.Done()` guard
- `http.Response.Body` not closed or not drained before close
- Resource opened without a `defer` close path
- `append` result not assigned back
- Concurrent map access without synchronization

## Flag unless there is a clear reason

- Named returns in functions longer than ~5 lines
- `init()` with side effects (I/O, global mutation, network)
- Global mutable state outside `main`
- Exported name stuttering with its package (`user.UserService`)
- `defer` inside a loop
- `time.Sleep` as synchronization
- Interface defined next to its implementation instead of at the call site
- Interface with a single method where a `func` parameter would suffice
- Test with more than two cases not table-driven
- `t.Fatal` / `t.Error` inside a goroutine spawned by the test
- Test asserting on error strings instead of sentinel errors or `errors.As`
- Type or function exported that belongs behind `internal/`

## Output

Lead each comment with the problem. Suggest a fix using only stdlib or code already in this repository. If the PR is clean, say so briefly.
