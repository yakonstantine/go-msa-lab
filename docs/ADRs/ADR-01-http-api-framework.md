## ADR-01: HTTP API Framework Selection per Service

**Status**: Accepted | **Date**: 2026-03-02

### Context

The project compares two architectural styles within the same distributed system:
* User Service - Clean Architecture
* SMTP Service - Minimalist Domain-Oriented Architecture

A decision is required regarding the HTTP layer implementation for each service. The goal is not only functional delivery but also architectural contrast, learning depth, and realistic production patterns.

Constraints:
* Go 1.25 baseline (ADR-0)
* No framework-driven architecture
* Explicit control over context propagation and error handling
* Avoid unnecessary third-party abstractions in minimalist service
* Preserve realistic industry relevance

The HTTP framework choice should reinforce the architectural style of each service rather than blur the contrast.

### Decision

The HTTP layer will differ intentionally between services:

#### User Service (Clean Architecture)

* Use `gin-gonic/gin`

**Rationale**:
* Provides structured routing and middleware support.
* Common in "enterprise-style" Go services.
* Does not enforce architectural layering; can remain confined to delivery layer.
* Keeps handlers thin and infrastructure-isolated.

Gin will be used strictly as a transport adapter.
Domain and use-case layers must not depend on Gin types.

#### SMTP Service (Minimalist Architecture)

* Use the standard library `net/http`.

**Rationale**:
* Zero external dependency.
* Maximum explicitness.
* Reinforces minimalist philosophy.
* Encourages direct control over request lifecycle and context propagation.
* Reflects idiomatic Go in lean microservices.

No router frameworks (chi, gorilla, etc.) will be introduced unless strictly required.

### Consequences

**Positive**:
* Clear architectural contrast between services.
* Demonstrates both enterprise-style and idiomatic Go approaches.
* Preserves isolation of HTTP concerns from domain logic.
* Improves learning depth across two patterns.

**Trade-offs**:
* Slight cognitive overhead maintaining two HTTP stacks.
* Handlers in SMTP service require more manual routing and middleware logic.
* Comparisons must account for framework differences when evaluating complexity.

**Learning Objectives**:
- Hands-on comparison of Gin's router/middleware/binding vs net/http ergonomics
- Evaluate developer experience in layered (CA) vs flat (minimalist) context
- Assess whether Gin's abstractions aid or obscure CA boundaries

### Rejected Alternatives
1. Use net/http in both services
  
   Rejected because:
   * Reduces architectural contrast.
   * Weakens experimental dimension of the lab.
   * Does not represent common enterprise Go stacks (where Gin is frequently used).
   * Makes the comparison purely about layering, not ecosystem trade-offs.

2. Use Gin in both services

   Rejected because:
   * Pollutes minimalist service with unnecessary abstraction.
   * Weakens demonstration of idiomatic standard library usage.
   * Makes architectural differences harder to isolate.

3. Use chi instead of Gin

   `go-chi/chi`
   
   Rejected because:
   * Adds dependency without strong differentiating value over net/http
   * Does not significantly improve Clean Architecture layering experiment.

4. Use Fiber

   `gofiber/fiber`
   
   Rejected because:
   * Less common in serious backend teams.
   * Inspired by Express.js style, less idiomatic in Go ecosystem.
   * Adds additional abstraction without architectural benefit.

5. Use Echo

   `labstack/echo`

   Rejected because:
   * Comparable to Gin in functionality.
   * No meaningful architectural difference for the lab.

## Containment Rules to Prevent Framework Bleed

**User Service (Gin)**:
* Gin types (`*gin.Context`) must not appear in use-case or domain layers
* Use-cases receive domain types, return domain errors
* Handlers translate between Gin and domain  
* If Gin's binding (`ShouldBindJSON`) becomes convenient to the point validation is skipped—document as a red flag

**SMTP Service (net/http)**:
* Handlers parse `*http.Request` directly
* No router beyond `http.ServeMux` unless routing becomes unwieldy (document if chi is added later)