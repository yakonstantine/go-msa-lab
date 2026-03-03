# DDR-01: SMTP Address Lifecycle and Uniqueness Enforcement

**Status**: Accepted | **Date**: 2026-03-03

## Context

User Service assigns SMTP addresses to users based on firstName, lastName, countryCode, and departmentCode. SMTP Service maintains a global registry of all SMTP addresses across the organization (including users, shared mailboxes, distribution lists, etc.).

Business requirements:
- SMTP addresses must be globally unique across the entire organization
- Once assigned, SMTP addresses are never reassigned—even after user deletion
- Users may accumulate multiple SMTP addresses over their lifetime (when primary changes)
- Name changes may trigger SMTP regeneration

The system must prevent SMTP conflicts while maintaining reasonable performance and avoiding complex distributed locking.

## Decision

### Ownership Model

**User Service**
- Authoritative for user-to-SMTP mapping
- Owns SMTP generation algorithm
- Enforces local uniqueness via database constraints
- Stores validation snapshot in `SMTPAddresses` table

**SMTP Service**
- Authoritative for global SMTP existence (all address types: users, mailboxes, lists)
- Validates SMTP availability across organization
- Subscriber to User Service events (eventual consistency)
- May reject SMTP allocations asynchronously if conflicts detected

**`SMTPAddresses` Table**
- Validation-only snapshot of locally assigned SMTPs
- Historical record of all SMTPs ever assigned to users
- Not synchronized with SMTP Service (eventual consistency)

### Uniqueness Enforcement Strategy

**Local Uniqueness (Strong Guarantee)**
- `SMTPAddresses.Address` has unique constraint (enforced globally across all domains)
- Prevents duplicate assignment within User Service
- Concurrent create requests with same firstName/lastName in same domain → different suffixes assigned
- Same firstName/lastName in different domains → no suffix needed (e.g., `john.doe@co.nl` and `john.doe@co.be`)

**Generation Algorithm**
1. Derive domain from countryCode/departmentCode lookup:
   - Default: `co-group.com`
   - Specific mappings (examples):
     - `NL + 1234` → `co.nl`
     - `BE + *` → `co.be`
2. Compute local-part: `{firstName}.{lastName}`
3. Build candidate: `{firstName}.{lastName}@{domain}`
4. Check uniqueness: `SELECT EXISTS(SELECT 1 FROM SMTPAddresses WHERE Address = ?)`
5. If exists in target domain: Append suffix (`.1`, `.2`, `.3`...) until unique
6. Insert atomically: `INSERT INTO SMTPAddresses (Address, CorpKey, Type)`
7. If INSERT fails (unique constraint violation): Retry with next suffix
8. Store in User table: `UPDATE User SET PrimarySMTP = ?`

**Note:** Suffix iteration is scoped per domain. `john.doe@co.nl` and `john.doe@co.be` can coexist without suffixes.

**Global Validation via SMTP Service**
- After local assignment, User Service emits `UserCreated` or `UserUpdated` event
- SMTP Service consumes event and validates against global registry
- If conflict detected (rare): SMTP Service may reject asynchronously

### Race Condition Handling

**Within User Service**
- Concurrent creates with identical firstName/lastName in same domain → both iterate to unique suffixes
- Concurrent creates with identical firstName/lastName in different domains → no conflict
- Unique constraint prevents duplicates globally
- No pessimistic locking required
- Examples:
  - Request A (NL/1234): `john.doe@co.nl` exists → assigns `john.doe.1@co.nl`
  - Request B (NL/1234, concurrent): `john.doe@co.nl` exists → assigns `john.doe.2@co.nl`
  - Request C (BE/any, concurrent): assigns `john.doe@co.be` (different domain, no suffix needed)

**Cross-Service Conflicts (SMTP Service rejection)**
- Strategy: Optimistic + Silent Repair
- If SMTP Service rejects (very rare—indicates sync drift):
  1. User Service generates new SMTP with next suffix
  2. Updates local DB (`User.PrimarySMTP` and `SMTPAddresses`)
  3. Emits `UserUpdated` event with corrected SMTP
  4. Logs warning with metric: `smtp_conflicts_repaired_total`
- Original create/update response returns the initially generated SMTP address
- If a repair occurs, subsequent GETs and the `UserUpdated` event will surface the corrected `PrimarySMTP`; no additional error is returned to the caller, and the correction typically completes within seconds

### Tombstoning Policy

**User Deletion**
- Sets `User.Deleted = true` (soft delete)
- SMTPs remain in `SMTPAddresses` table indefinitely
- Deleted users' SMTPs never released for reuse (across all domains)
- Next user with same firstName/lastName in same domain gets numeric suffix
- Next user with same firstName/lastName in different domain may not need suffix

**Why Never Reassign**
- Email addresses have organizational memory (audit trails, external systems)
- Prevents security/privacy issues (new user receiving old user's emails)
- Simplifies uniqueness logic (monotonic growth)

**Query Filtering**
- All read endpoints: `WHERE User.Deleted = false`
- Tombstoned SMTPs invisible to clients
- `SMTPAddresses` table grows unbounded (acceptable for user scale)

### Secondary SMTP Semantics

**Accumulation**
- When primary SMTP changes: old primary → `UPDATE SMTPAddresses SET Type='Secondary' WHERE Address=? AND Type='Primary'`
- Append-only—never removed
- Unsorted in GET responses (insertion order or arbitrary)
- Growth unbounded per user (typically low, <10 per user lifetime)

**Read Model Query:**
```sql
SELECT Address 
FROM SMTPAddresses 
WHERE CorpKey = ? AND Type = 'Secondary'
-- No ORDER BY (unsorted)
-- No Deleted filter (SMTPAddresses has no Deleted column, linked via User table)
```

**SMTP Swap Logic (on Update):**

Primary regeneration triggered if:
- firstName or lastName changes (local-part changes)
- countryCode or departmentCode changes such that domain changes

Examples:
- `John Doe (NL/1234)` → `John Smith (NL/1234)`: local-part changes, domain stays `co.nl`
- `John Doe (NL/1234)` → `John Doe (BE/any)`: local-part stays, domain changes to `co.be`
- `Jöhn Doe` → `John Doe`: no change if normalized local-part stays same

**Same-Domain Update (firstName/lastName change only):**
- If regenerated primary already exists in user's secondaries:
  1. Promote: `UPDATE SMTPAddresses SET Type='Primary' WHERE Address=? AND CorpKey=?`
  2. Demote: `UPDATE SMTPAddresses SET Type='Secondary' WHERE Address=? AND CorpKey=?`
  3. Update User table: `UPDATE User SET PrimarySMTP = ? WHERE CorpKey=?`
  4. No new SMTP allocated—just swap
  
- If regenerated primary is unique in target domain:
  1. Old primary: `UPDATE SMTPAddresses SET Type='Secondary' WHERE Address=? AND Type='Primary'`
  2. New primary: `INSERT INTO SMTPAddresses (Address, CorpKey, Type='Primary')`
  3. Follow suffix iteration within target domain if needed

**Cross-Domain Update (countryCode/departmentCode change):**
- Cannot swap (different domain)
- Always generates new SMTP in target domain
- Old primary always moves to secondaries
- New primary follows suffix iteration in target domain
- Example: `john.doe@co.nl` → moved to country BE → generates `john.doe@co.be` (or `.1` if exists)

### Edge Cases and Limits

**Suffix Iteration Limit**
- No hard limit enforced (implementation may set practical limit like `.999`)
- If limit reached: return 500 Internal Server Error (extremely rare)
- Monitor via metric: `smtp_suffix_max` (track highest suffix assigned)

**Typo Corrections:**
- Updating firstName/lastName/countryCode/departmentCode without changing SMTP (local-part or domain): no regeneration
- Example: Fixing diacritics that normalize to same ASCII local-part
- Example: Updating departmentCode within same domain mapping (e.g., NL/5678 → NL/9999 if both map to co.nl)

**Cross-Service Sync Drift:**
- Monitor: `smtp_service_rejections_total` metric
- Alert if repair rate exceeds threshold (indicates persistent drift)
- Manual investigation required if frequent

## Consequences

**Positive**
- Strong local uniqueness guarantee (no duplicates within User Service)
- Simple algorithm (suffix iteration, no distributed locks)
- Auditable (all SMTPs preserved historically)
- Safe (never reuse deleted user addresses)

**Trade-offs**
- `SMTPAddresses` table grows unbounded (mitigated: user scale typically <100K)
- Secondary SMTPs grow per user (mitigated: typically <10 lifetime)
- Eventual consistency with SMTP Service (mitigated: conflicts very rare)
- Silent repair invisible to clients (mitigated: observability via logs/metrics)

**Database Schema Requirements**
- `SMTPAddresses.Address` UNIQUE constraint (enforced at DB level)
- `User.Deleted` indexed for query performance
- Composite index: `(CorpKey, Type)` on `SMTPAddresses` for secondary read queries

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
