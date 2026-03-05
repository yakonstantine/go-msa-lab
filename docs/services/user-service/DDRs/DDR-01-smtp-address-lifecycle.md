# DDR-01: SMTP Address Lifecycle and Uniqueness Enforcement

**Status**: Accepted | **Date**: 2026-03-03

## Context

User Service assigns SMTP addresses based on firstName, lastName, countryCode, and departmentCode. SMTP Service maintains a global registry including users, shared mailboxes, and distribution lists.

**Business Requirements:**
- SMTP addresses must be globally unique
- Once assigned, addresses are never reassigned—even after user deletion
- Users accumulate multiple addresses over their lifetime as primary changes
- Name or location changes may trigger new SMTP generation

**Challenge:** Prevent SMTP conflicts while maintaining reasonable performance without complex distributed locking.

## Decision

### Ownership Model

**User Service:**
- Authoritative for user-to-SMTP mapping
- Owns SMTP generation algorithm
- Enforces local uniqueness via database constraints
- Maintains historical record of all assigned SMTPs

**SMTP Service:**
- Authoritative for global SMTP existence (all types)
- Validates SMTP availability across organization
- Subscriber to User Service events (eventual consistency)
- May reject allocations asynchronously if conflicts detected

**Consistency Model:** Eventual consistency between services. Strong local guarantees, optimistic global validation.

### SMTP Generation Strategy

**Pattern:** `{firstName}.{lastName}[.suffix]@{domain}`

**Domain Derivation:**
- Determined by countryCode + departmentCode lookup
- Default: `co-group.com`
- Specific mappings: `NL/1234 → co.nl`, `BE/* → co.be`, etc.

**Uniqueness Approach:**
- Local-part (firstName.lastName) checked within target domain only
- Numeric suffix (`.1`, `.2`, `.3`...) appended if conflict exists in that domain
- Same name in different domains = no suffix needed

**Examples:**
- `john.doe@co.nl` exists → new user in NL/1234 gets `john.doe.1@co.nl`
- `john.doe@co.nl` exists → new user in BE domain gets `john.doe@co.be` (different domain, no suffix)
- `john.doe@co.nl` and `john.doe@co.be` can coexist (same person, different locations)

**Conflict Resolution:**
- Concurrent creates in same domain: both get unique suffixes via database constraint
  - Request A (NL/1234): `john.doe@co.nl` exists → assigns `john.doe.1@co.nl`
  - Request B (NL/1234, concurrent): `john.doe@co.nl` exists → assigns `john.doe.2@co.nl`
- Cross-domain: no conflict (different addresses)
  - Request C (BE, concurrent): assigns `john.doe@co.be`
- No pessimistic locking required

### Cross-Service Validation

**Strategy:** Optimistic + Silent Repair

During creation/update, User Service validates SMTP availability by calling SMTP Service HTTP API (GET /smtps/:smtp):
- 404 → SMTP available, proceed with assignment
- 200 → SMTP taken, increment suffix and retry

After local assignment, User Service emits event to SMTP Service for eventual consistency. If SMTP Service detects conflict (rare—indicates sync drift):
- User Service generates new SMTP with next suffix
- Updates local database
- Emits correction event
- Logs metric: `smtp_conflicts_repaired_total`

**Client Impact:** Original response returns initially generated address. If repair occurs, subsequent reads reflect corrected address. Correction typically completes within seconds.

**Example:**
- User created with `john.doe@co.nl`, returns 201
- SMTP Service later rejects (conflict with shared mailbox)
- User Service auto-corrects to `john.doe.1@co.nl`
- Next GET returns updated address

### Deletion Policy

**Soft Delete:** User marked as deleted, data retained, SMTP addresses tombstoned forever.

**Why Never Reassign:**
- Email addresses have organizational memory (audit trails, external systems)
- Prevents security issues (new user receiving old user's emails)
- Simplifies uniqueness logic

**Behavior:**
- Deleted users excluded from queries
- Their SMTPs never released for reuse globally
- Next user with same name in same domain gets numeric suffix
  - Example: `john.doe@co.nl` deleted → new user in NL/1234 gets `john.doe.1@co.nl`
- Next user with same name in different domain may not need suffix
  - Example: `john.doe@co.nl` deleted → new user in BE gets `john.doe@co.be`

### Update Behavior

**Primary SMTP Regeneration Triggers:**
- firstName or lastName changes → local-part changes
  - Example: `John Doe (NL/1234)` → `John Smith (NL/1234)` regenerates to `john.smith@co.nl`
- countryCode or departmentCode changes → domain changes
  - Example: `John Doe (NL/1234)` → `John Doe (BE/any)` regenerates from `co.nl` to `co.be`
- No change if normalized pattern stays same
  - Example: `Jöhn Doe` → `John Doe` (typo fix with same normalized result)

**Same-Domain Updates:**
- If regenerated SMTP exists in user's secondaries: swap (promote from secondaries, demote current primary)
  - No new SMTP allocated, just rearrange existing addresses
- If regenerated SMTP is unique: old primary → secondaries, assign new primary
  - Old: `john.doe@co.nl` → secondaries
  - New: `john.smith@co.nl` (or `.1` if conflict)

**Cross-Domain Updates:**
- Cannot swap (different domains)
- Old primary always → secondaries
- New primary generated in target domain (with suffix if needed)
  - Example: `john.doe@co.nl` → moved to BE → becomes `john.doe@co.be` (or `.1` if exists)

**Secondary Addresses:**
- Append-only historical record
- Never removed
- Unsorted in responses
- Typically < 10 per user lifetime

### Edge Cases

**Suffix Limits:**
- No hard limit enforced (implementation may set practical limit like `.999`)
- If limit reached: return 500 error (extremely rare)

**No-Op Updates:**
- Changes that don't affect SMTP pattern → no regeneration
  - Example: Fixing diacritics that normalize to same ASCII (`Jöhn` → `John`)
  - Example: Department code change within same domain (NL/5678 → NL/9999 both map to `co.nl`)

**Monitoring:**
- Track `smtp_suffix_max` (highest suffix assigned)
- Track `smtp_service_rejections_total` (async validation failures)
- Alert if rejection rate increases (indicates sync drift)

## Consequences

**Benefits:**
- Strong local uniqueness guarantee (no duplicates)
- Simple algorithm (suffix iteration, no distributed locks)
- Full audit trail (all SMTPs preserved)
- Safe (no address reuse)

**Trade-offs:**
- Storage grows unbounded (mitigated: user scale typically < 100K)
- Secondary addresses grow per user (mitigated: typically < 10 lifetime)
- Eventual consistency with SMTP Service (mitigated: conflicts very rare)
- Silent repair invisible to clients (mitigated: observable via metrics/events)

**Operational Requirements:**
- Monitor suffix distribution and conflict repair metrics
- Alert on unusual rejection rates
- Database unique constraint enforcement on SMTP addresses
- Indexed queries for soft-deleted users

## Rejected Alternatives

1. **Pessimistic Locking (Distributed Lock)**
   - Use Redis or Postgres advisory lock during SMTP generation
   - Rejected: Adds external dependency, increases latency, complicates failure modes

2. **Synchronous SMTP Service Validation**
   - Block create/update response until SMTP Service confirms
   - Rejected: Adds 100-500ms latency, tight coupling, reduces availability

3. **SMTP Reuse After Deletion**
   - Release SMTPs back to pool after user deletion
   - Rejected: Security risk (email forwarding), audit complexity, violates business requirement

4. **Bounded Secondary List**
   - Limit secondaries to last N addresses
   - Rejected: Loses audit trail, business requirement is full history

5. **SMTP Service as Source of Truth**
   - User Service queries SMTP Service for current user SMTPs
   - Rejected: Cross-service dependency on read path, performance impact, availability coupling
