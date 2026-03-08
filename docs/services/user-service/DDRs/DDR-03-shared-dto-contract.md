# DDR-03: Shared DTO Contract for API and Event Payloads

**Status**: Accepted | **Date**: 2026-03-08

## Context

User Service exposes user data through two channels:
- **REST API** — synchronous pull model for on-demand queries
- **Domain Events** — asynchronous push model via message broker (planned)

Consumers integrating with User Service must choose between these models based on their needs. Some consumers may use both — pulling on startup and subscribing to events for live updates.

**Challenge:** If API responses and event payloads use different structures, consumers must maintain two mapping paths for the same logical data. This increases integration complexity and creates a surface area for subtle shape mismatches.

## Decision

### Unified DTO Contract

API responses and event payloads share the same DTO structs, defined in `internal/dto/`.

**Mapping:** Conversion from `*entity.User` to DTO lives in the adapter layer (handler or event publisher), not on the entity itself.

### Rationale

- **Consumer simplicity:** Clients choose pull or push integration with an identical data contract. No need to learn or reconcile two schemas.
- **Clean Architecture compliance:** DTOs sit in the adapter tier. The use case layer returns domain entities; each adapter maps to the shared DTO independently.
- **Single source of truth:** One struct definition prevents drift between API and event representations.

### Constraints

- **Additive-only evolution:** Since changes affect both channels simultaneously, fields may only be added, never removed or semantically changed. Removals require a deprecation period.
- **Envelope separation:** Event infrastructure will wrap DTOs in an envelope (event type, timestamp, correlation ID). The envelope is not part of the DTO — it is owned by the event publishing layer.
- **No domain leakage:** DTOs must not expose internal domain details (e.g., soft-delete flags, internal IDs) that are not relevant to consumers.

## Consequences

**Benefits:**
- Consumers integrate once regardless of delivery channel
- Fewer types to maintain and document
- Shape mismatches between API and events are structurally impossible

**Trade-offs:**
- API and event schemas are coupled — a field needed only by one channel is visible to both
- Cannot independently version API responses and event payloads
- Additive-only constraint may limit ability to clean up deprecated fields quickly

**Acceptable because:** The User Service domain is narrow and stable. The simplicity gain for consumers outweighs the flexibility cost. If the schemas need to diverge significantly in the future, this decision can be revisited by splitting the DTO package at that point.
