---
name: code-review
description: Review Go code changes for correctness and production risks in this repository. Use when asked to review Go code, PRs, or diffs. Optionally scoped to a specific service by name (user-service or smtp-service).
---

# Go Code Review

## Service Scope

If the user's message names a specific service, scope the entire review to that service only.
Supported services and their paths:

| Service | Source | DDRs |
|---|---|---|
| `user-service` | `services/user-service/` | `docs/services/user-service/DDRs/` |
| `smtp-service` | `services/smtp-service/` | *(no DDRs recorded yet)* |

If no service is named, review all changed Go files across the repository.

## Context to Load Before Reviewing

1. Read `README.md` - use the service purposes as a correctness lens.
2. Read all files under `docs/ADRs/` - flag code that contradicts any recorded decision.
3. If the scope includes `user-service`, read all files under `docs/services/user-service/DDRs/` and flag contradictions.
4. If the change implies a new architectural decision that is not yet recorded, mention that an ADR or DDR may be warranted. Do not block on it unless it is a correctness risk.

## Scope

- Prefer changed files and their direct call sites within the resolved service path.
- Prioritize Go source files and their tests.

## Always Flag - bugs and production risks

- Error return silently discarded (missing `if err != nil`, or `_` with no comment)
- `fmt.Errorf` without `%w` - breaks the error chain
- Error inspected by string matching instead of `errors.Is` / `errors.As`
- `panic` for anything recoverable (bad input, missing config, network errors)
- Goroutine launched without a clear exit path or lifetime owner
- Function doing I/O or blocking work without `context.Context` as its first parameter
- `context.Context` stored in a struct field
- `sync.Mutex` copied by value
- Channel send/receive that can block forever without a `ctx.Done()` guard
- `http.Response.Body` not closed, or closed without draining first
- File, DB, or network resource opened without a `defer` close path
- `append(s, x)` result not assigned back - the original slice is unchanged
- Concurrent map access without synchronization

## Flag Unless There Is a Clear Reason Not To

- Named returns in functions longer than ~5 lines
- `init()` with side effects (I/O, global mutation, network)
- Global mutable state outside `main`
- Exported name stuttering with its package (`user.UserService`, `smtp.SMTPClient`)
- `defer` inside a loop
- `time.Sleep` as a synchronization primitive
- Any pattern in the "Over" column of `docs/ADRs/ADR-00-go-version.md` - prefer the modern equivalent listed there
- Interface defined next to its implementation when it is only used by that implementation - prefer defining interfaces at the call site
- Interface with a single method where a `func` parameter would be simpler
- Test with more than two cases not table-driven
- `t.Fatal` / `t.Error` called inside a goroutine spawned by the test
- Test asserting on error message strings instead of sentinel errors or `errors.As`
- Type or function exported that belongs behind `internal/`

## Output Format

Group findings under these headings:

**Blocking** - bugs, production risks, safety issues that must be fixed before merge.

**Should fix** - clear violations of Go idioms or project conventions; not blocking but strongly recommended.

**Optional** - minor improvements; take or leave.

**Summary** - one sentence on overall code health.

For each finding include: file and function, what is wrong, why it matters, and a minimal fix direction.
If there are no findings in a category, omit that heading.
If the code is clean, say so in one sentence under Summary.

## Tone

Write comments as you would in a real PR review - inline, conversational, and to the point. Lead with the problem, follow with a suggested fix or direction.
If suggesting an alternative API, only reference standard library or code already present in this repository. Do not invent helpers.
If contradicting an ADR, reference it briefly and suggest alignment. Do not demand redesign unless it is a correctness issue.