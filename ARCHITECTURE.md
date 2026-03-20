# Platform Architecture

## Overview

`platform` is the Go backend control plane for Kleff. It exposes a single
versioned HTTP API that accepts user and operator intent, enforces authorization,
persists desired state, and coordinates with game-server daemons running on
compute nodes.

---

## Repository layout

```
platform/
  cmd/api/                     → Binary entrypoint (main.go)
  internal/
    bootstrap/                 → App wiring: config, DI container, router, lifecycle
    core/                      → Domain modules (modular monolith)
      identity/                → Users and OIDC session management
      organizations/           → Org tenancy and membership
      deployments/             → Deploy intent and lifecycle tracking
      nodes/                   → Compute node inventory (daemon-reported)
      billing/                 → Subscriptions, invoices, Stripe integration
      usage/                   → Metering and usage records
      audit/                   → Append-only audit event log
      admin/                   → Operator-only management surfaces
    shared/                    → Cross-cutting infrastructure concerns
      ids/                     → Collision-resistant ID generation
      clock/                   → Testable time interface
      config/                  → Typed env-var helpers
      logging/                 → Structured logger factory
      middleware/              → HTTP middleware (auth, RBAC)
      database/                → PostgreSQL connectivity helpers
      events/                  → In-process domain event bus
  packages/                    → Shared Go library (go-common)
    adapters/http/             → Request helpers, middleware, response writers
    bootstrap/                 → Graceful HTTP server lifecycle
    domain/                    → Error types, pagination, response envelopes
  api/openapi/                 → OpenAPI 3.1 contract
  migrations/                  → SQL schema migrations (goose)
  infra/                       → Kubernetes manifests and Helm chart values
```

---

## Module structure (hexagonal architecture)

Each domain module in `internal/core/<module>/` follows the same layout:

```
<module>/
  domain/           → Pure Go structs, value objects, domain logic — zero external deps
  ports/            → Go interfaces: repositories, event publishers, external service clients
  application/
    commands/       → Write-side use cases  (CreateX, UpdateX, DeleteX, …)
    queries/        → Read-side use cases   (GetX, ListX, …)
    policies/       → Authorization checks per use case
  adapters/
    http/           → HTTP handler — decodes request, calls use case, encodes response
    persistence/    → SQL repository implementations (pgx)
    external/       → Third-party API adapters (Stripe, email, etc.)
```

**Dependency rule:** imports always flow inward.
`adapters` → `application` → `domain`. `domain` has no outward imports.

---

## API design principles

This API is a **control plane**, not a thin CRUD layer:

- Accepts **intent** from authenticated users and operators
- Validates and enforces **authorization policies** before mutating state
- Persists **desired state** to Postgres
- Emits **domain events** via the event bus for background workers and daemons
- Returns **read models** shaped for efficient UI consumption

The platform does **not** call container runtimes directly — daemon workers
subscribe to domain events and execute the actual server operations.

---

## Authentication & authorization

- Bearer JWTs issued by an OIDC provider (Keycloak / Auth0)
- `internal/shared/middleware.RequireAuth` validates tokens and injects `Claims`
  into the request context (`sub`, `email`, `org_id`, `roles`)
- `middleware.RequireRole(...)` enforces role-based access at the route level
- All `/api/*` routes are wrapped with `RequireAuth` in the bootstrap router

---

## Daemon communication

The game-server daemon (`gameserver-daemon`) is the execution layer:

- Daemon registers compute nodes via `POST /api/v1/nodes`
- Sends heartbeats via `PATCH /api/v1/nodes/{id}` (status, resource metrics)
- Picks up deployment jobs by subscribing to `deployment.created` domain events
- Reports outcomes back via `POST /api/v1/deployments/{id}/status`

The platform **never** initiates outbound calls to daemons.

---

## Domain status

| Module        | Status        | Notes |
|---------------|---------------|-------|
| Identity      | Scaffolded    | OIDC user provisioning, org membership |
| Organizations | Scaffolded    | Multi-tenant boundary, member RBAC |
| Deployments   | Scaffolded    | Intent recording, daemon dispatch via events |
| Nodes         | Scaffolded    | Daemon registration and heartbeat |
| Billing       | Scaffolded    | Stripe subscriptions and invoices |
| Usage         | Scaffolded    | Per-org resource metering |
| Audit         | Scaffolded    | Append-only immutable event log |
| Admin         | Scaffolded    | Staff-only operator tooling |
| Projects      | Planned       | Resource grouping within an org |
