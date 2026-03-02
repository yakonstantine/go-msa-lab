# Roadmap

## MS-1: Functional Baseline (MVP)

**Goal**: End-to-end working system with clean boundaries

### Scope
- User CRUD API (Create, Read, Update, Delete)
- SMTP uniqueness validation (local check + HTTP call to SMTP Service)
- Audit log storage (async channel-based processing)
- Event publishing (UserCreated, UserUpdated, UserDeleted → RabbitMQ)
- SMTP Service event consumer (subscribe + store in DB)
- Database migration tooling (schema versioning)
- Health check endpoints (both services)
- Docker Compose setup (PostgreSQL, RabbitMQ, both services)
- Basic integration tests (happy path validation)

### Acceptance Criteria
- [ ] `POST /users` creates user → SMTP validated → event published → SMTP Service receives & stores
- [ ] `PUT /users/:id` updates user → old SMTP moved to secondaries → event published
- [ ] `DELETE /users/:id` removes user → event published → SMTP Service marks as deleted
- [ ] `GET /users/:id` returns user with primary + secondary SMTPs
- [ ] Create user with duplicate SMTP → suffix appended automatically
- [ ] User changes stored in audit log asynchronously
- [ ] Health checks return 200 OK when services are healthy
- [ ] `docker compose up` starts entire system with seeded databases

### Out of Scope (MS-2)
- Retry logic, error recovery, message redelivery
- Handling SMTP Service downtime
- Concurrent uniqueness conflicts
- Event ordering guarantees
- Idempotency

### Deliverable
Working distributed system demonstrating Clean Architecture (User Service) vs. minimalist Go (SMTP Service). All services runnable via Docker Compose. Happy path works reliably.

## MS-2: Consistency & Failure Modes

**Goal**: Production-ready reliability and correctness

### Scope
- Outbox pattern (atomic event publishing with DB transaction)
- Idempotent event processing (deduplication in SMTP Service)
- Event ordering strategy (sequence numbers or acceptable out-of-order handling)
- Message retry with exponential backoff
- Dead letter queue for poison messages
- Concurrent uniqueness conflict resolution (database constraints + test)
- End-to-end testing (cover main user flows and failure scenarios)

### Acceptance Criteria
- [ ] User Service crashes after DB commit, before event publish -> event still delivered (outbox)
- [ ] Same event delivered twice -> SMTP Service processes once (idempotency)
- [ ] Publish UserUpdated before UserCreated -> handled gracefully (ordering)
- [ ] SMTP Service is down during user creation -> request succeeds or fails predictably (consistency model enforced)

### Deliverable
System handles real-world distributed failures: message loss, redelivery, race conditions, service downtime. Documented consistency guarantees (ADRs recorded during implementation).

## MS-3: Observability & Operations
**Goal**: Production-ready monitoring and debugging

### Scope
- Structured logging with `log/slog` (replace any `fmt.Println` or basic logging)
- Correlation IDs (HTTP headers + event metadata for request tracing)
- Prometheus-compatible metrics endpoints (latency, error rates, queue depth)
- DLQ log-based alerting simulation
- Schema evolution guide (backward-compatible event changes)

### Acceptance Criteria
- [ ] Create user -> correlation ID appears in User Service logs, RabbitMQ event, SMTP Service logs
- [ ] Metrics endpoint exposes: `smtp_check_duration_seconds`, `events_published_total`, `events_processed_total`
- [ ] Slow SMTP Service (inject delay) -> P99 latency visible in metrics
- [ ] Event stuck in DLQ -> correlation ID + error reason visible in logs
- [ ] Deploy User Service with new optional event field -> SMTP Service continues working (backward compatibility)
- [ ] README includes: how to run, how to debug issues

### Deliverable
Portfolio-ready system with production-grade observability. Silent failures are debuggable. Safe to demo to employers with "this is how I'd run it in production" narrative.