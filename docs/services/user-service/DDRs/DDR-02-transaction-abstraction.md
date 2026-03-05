# DDR-02: Transaction Abstraction and Local Consistency

**Status**: Accepted | **Date**: 2026-03-05

## Context

User Service must ensure atomic creation of user records and their associated SMTP addresses. Both entities reside in the same PostgreSQL database but are stored in separate tables (`users` and `smtp_addresses`).

**Requirements:**
- User and SMTP address must be created atomically — no orphaned users without SMTP tracking
- SMTP uniqueness checks must happen before insertion
- Transaction management must respect Clean Architecture boundaries
- Solution must be simple enough for a learning lab without over-engineering

**Constraint:** User Service uses Clean Architecture layering, where use case layer must not depend directly on infrastructure libraries like `database/sql`.

**Challenge:** Balance Clean Architecture purity (no infrastructure dependencies in use case) against Go's idiomatic transaction patterns (explicit `*sql.Tx` usage).

## Decision

### Transaction Abstraction Pattern

**Approach:** Transaction-as-parameter with explicit lifecycle in use case layer.

**Interfaces:**
```go
type Transaction interface {
    Commit() error
    Rollback()  // No error return — intentionally silent on rollback failures
}

type TransactionFactory interface {
    BeginTx(context.Context) (Transaction, error)
}

type UserRepository interface {
    GetByCorpKey(context.Context, entity.CorpKey) (*entity.User, error)
    Create(context.Context, Transaction, *entity.User) error  // Transaction parameter
}

type SMTPRepository interface {
    GetByEmail(context.Context, entity.Email) (*entity.SMTPAddress, error)
    Create(context.Context, Transaction, *entity.SMTPAddress) error  // Transaction parameter
}
```

**Usage in Use Case:**
```go
func (uc *UseCase) createTx(ctx context.Context, u *entity.User) error {
    tx, err := uc.txf.BeginTx(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback()  // Safe no-op if already committed

    err = uc.userRepo.Create(ctx, tx, u)
    if err != nil {
        return err
    }

    err = uc.smtpRepo.Create(ctx, tx, smtp)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

**Dependencies in UseCase:**
- `TransactionFactory` — creates transactions for write operations
- `UserRepository` — used for both transactional writes and non-transactional reads
- `SMTPRepository` — used for both transactional writes and non-transactional reads

### Rationale

Transaction-as-parameter balances Clean Architecture constraints with Go idioms:
- Simple: Transaction passed to repository methods that need it
- Explicit: Transaction lifecycle visible in use case layer
- Flexible: Reads don't require transaction parameter
- Go-idiomatic: Mirrors `database/sql` patterns while abstracted
- Testable: Easy to mock Transaction interface

**Trade-offs:**
- Repository interfaces are aware of transactions (minor CA violation)
- Caller must remember to pass Transaction (not type-safe)
- `Rollback()` errors silently ignored (acceptable for most cases)

### Known Limitation: Optimistic SMTP Uniqueness Check

SMTP uniqueness check happens **before** transaction begins, creating a race condition:

```go
// SMTP generation OUTSIDE transaction
primarySMTP, err := generatePrimarySMTP(ctx, uc.smtpRepo, up)

// Transaction begins here
err = uc.createTx(ctx, u)
```

**Concurrent scenario:**
1. Request A: Check `john.doe@co.nl` → available
2. Request B: Check `john.doe@co.nl` → available (concurrent)
3. Request A: Insert → commit
4. Request B: Insert → **unique constraint violation**

**Why:** `smtpRepo.GetByEmail()` doesn't accept `Transaction` parameter, so the check can't happen within transaction isolation.

**Impact:** Concurrent creates with identical name/domain may fail with database error instead of auto-incrementing suffix.

**Mitigation:** Database unique constraint prevents data corruption. Failed requests can be retried by client.

**Future (MS-2):** Add retry logic with exponential backoff on unique constraint violations.

> **Known Limitation:** SMTP uniqueness check is optimistic (non-transactional). Concurrent creates with identical name/domain may fail with constraint violation. Retry logic will be added in MS-2.

## Consequences

**Benefits:**
- Atomic user + SMTP creation (no orphaned users)
- Clean Architecture boundaries preserved
- Simple abstraction (two interfaces)
- Explicit transaction lifecycle

**Limitations:**
- SMTP uniqueness check is non-atomic (race condition possible)
- Rollback errors silently ignored

**Operational:**
- Monitor unique constraint violations: `user_create_conflicts_total`
- Database unique constraint on `smtp_addresses.address` required

## Rejected Alternatives

### 1. Transaction-Scoped Repositories

```go
type Transaction interface {
    UserRepository() UserRepository
    SMTPRepository() SMTPRepository
    Commit() error
    Rollback()
}
```

Rejected: Adds 3+ abstraction layers, requires per-transaction repository instances, over-engineered for CRUD operations.

### 2. Direct `database/sql` Usage

```go
type UseCase struct {
    db *sql.DB  // Direct infrastructure dependency
}
```

Rejected: Violates Clean Architecture dependency rule, defeats lab's architectural comparison purpose.

### 3. Unit of Work Pattern

```go
type UnitOfWork interface {
    Users() UserRepository
    SMTPs() SMTPRepository
    Complete() error
}
```

Rejected: Not idiomatic in Go, hides transaction lifecycle, more complex than transaction-as-parameter.

### 4. CQRS with Command/Query Split

```go
type QueryService struct { /* ... */ }
type CommandService struct { /* ... */ }
```

Rejected: Over-engineered for simple CRUD, no performance justification for read/write separation.

### 5. Transactional Reads Everywhere

Wrap all queries in transactions for uniformity.

Rejected: Unnecessary transaction overhead (~1-5ms per read), increases connection pool pressure, obscures side-effect-free nature of reads.
