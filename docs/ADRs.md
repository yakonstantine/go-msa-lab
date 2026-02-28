# Architectural Decision Records

## ADR-0: Go Version and Language Feature Baseline

**Status**: Accepted | **Date**: 2026-02-28

### Context

Many Go resources target older versions. Without a baseline, reviewers may flag modern idioms as wrong or miss that a better API exists.

### Decision

Target **Go 1.25** as the minimum version. It's the latest stable release when the project started and adds `sync.WaitGroup.Go()` — the last concurrency ergonomic we needed.

**Prefer these patterns in all new code:**

| Pattern | Prefer | Over |
|---|---|---|
| Generics | `any`, generic free functions | `interface{}`, type-specific duplicates |
| Collections | `slices.SortFunc`, `slices.Clone`, `maps.Clone`, `maps.Keys` | `sort.*`, manual loops |
| Builtins | `min()`, `max()` | manual `if` comparisons |
| Errors | `errors.Is`/`errors.As`, `fmt.Errorf("%w", ...)`, `errors.Join` | string matching, bare wrapping, manual concat |
| Logging | `log/slog` | `log.*` |
| Iteration | `range n`, `iter.Seq2[K, V]` | `for i := 0; i < n; i++` |
| Concurrency | `wg.Go(f)`, `sync.RWMutex` | `Add(1)` + `defer Done()` |

**Runtime changes to be aware of (no code required):**
- Go 1.22: loop variables are scoped per iteration — `x := x` workarounds are no longer needed.
- Go 1.25: nil pointer check ordering bug (introduced in 1.21) is fixed — accessing a result before checking its error will now correctly panic.

### Consequences

- Reviews must not flag any pattern in the "Prefer" column above as incorrect.
- Reviews should flag: `ioutil.*`, `sort.Slice`, `log.*` in application code, and `Add(1)`/`Done()` where `wg.Go` applies.
- Using a Go version above 1.25 requires amending this ADR.