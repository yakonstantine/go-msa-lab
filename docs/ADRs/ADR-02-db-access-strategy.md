## ADR-2: Database Access Strategy per Service

**Status**: Accepted | **Date**: 2026-03-02

### Context

The system uses PostgreSQL for persistence in both services.

The project aims to:
* Avoid ORM-driven domain models.
* Maintain explicit transaction boundaries.
* Preserve control over SQL and schema.
* Prevent EF Core mental model carryover.
* Compare layered vs minimalist architecture without introducing hidden persistence magic.

A database access strategy must be selected per service that supports these goals while reflecting realistic production practices in Go backend teams.

### Decision

The database layer will differ intentionally between services.

#### User Service (Clean Architecture)

* Use `jmoiron/sqlx`
* Use `database/sql` under the hood.

**Rationale**:
* Keeps SQL explicit.
* Avoids ORM behavior (no entity tracking, no implicit joins).
* Provides ergonomic struct scanning and named queries.
* Suitable for layered repository implementation.
* Allows clean infrastructure-layer isolation without leaking persistence details into domain.

`sqlx` will be used only in the infrastructure layer. Domain models must not include any DB or persistence tags; all struct field tagging for database mapping lives in infrastructure-facing types only.

#### SMTP Service (Minimalist Architecture)

* Use `jackc/pgx` with `pgxpool`.
* Use handwritten SQL queries.

**Rationale**:
* Native PostgreSQL driver.
* Explicit transactions and connection management.
* No reflection.
* No runtime mapping layer.
* Aligns with idiomatic Go minimalist style.
* Represents serious production backend patterns.

No ORM or reflection-based mapping layer will be introduced in this service.

### Consequences

**Positive**:
* Clear contrast between layered repository abstraction (sqlx) and direct driver usage (pgx).
* SQL remains explicit in both services.
* No ORM magic or entity tracking introduced.
* Strong production realism.
* Enhances understanding of transaction boundaries and connection lifecycles.

**Trade-offs**:
* Two different DB APIs must be maintained.
* Requires discipline to prevent sqlx leakage into domain layer.
* Slight tooling divergence in tests and setup.

**Learning Objectives**:
* Compare sqlx struct scanning vs pgx explicit mapping
* Evaluate reflection cost and debuggability
* Assess transaction ergonomics in both

This decision intentionally avoids full ORMs (e.g., GORM) to prevent hidden persistence behavior and to preserve architectural clarity.

## Rejected Alternatives

1. Use GORM

    `gorm.io/gorm`
    
    Rejected because:
    * Introduces ORM behavior (entity tracking, hooks, implicit joins).
    * Encourages EF-style mental model carryover.
    * Hides SQL and transaction boundaries.
    * Distorts architectural comparison.
    * Reduces explicitness required for event-driven systems.

2. Use pgx in both services

   Rejected because:
   * Removes contrast in persistence approach.
   * Weakens experimental dimension.
   * Makes Clean Architecture evaluation less interesting.

3. Use sqlc

   `sqlc-dev/sqlc`

   Rejected because:
   * Adds code generation complexity.
   * Introduces additional CI/tooling concerns.
   * Shifts experiment from architectural contrast to generation-vs-manual comparison.
   * Over-optimizes for compile-time safety in a learning lab where runtime visibility is educational.

4. Use ent

   `ent/ent`

   Rejected because:
   * Heavy code generation.
   * Schema-driven model design.
   * Moves system toward framework-driven architecture.
   * Reduces explicit SQL ownership.

5. Use sqlx in both services

   Rejected because:
   * Removes driver-level contrast.
   * Minimalist service should demonstrate direct driver control.
   * Weakens demonstration of pgx transaction management.

## Mapping Discipline to Preserve CA

**User Service (sqlx)**:
* Domain models live in `domain/` with **zero persistence tags**
Infrastructure DTOs in `infrastructure/persistence/` carry db struct tags (`db:"..."`)  
* Repositories map DTO <-> Domain explicitly
* If tempted to put db struct tags (`db:"..."`) on domain models to save boilerplate—stop and document the cost/benefit

**SMTP Service (pgx)**:
* Inline scanning acceptable (no domain layer to protect)
* Handlers or thin data layer can use struct tags if needed