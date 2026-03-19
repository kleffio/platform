# @kleff/shared-types

TypeScript types shared across all Kleff apps and services.

## Module

```
@kleff/shared-types
```

## Structure

```
src/
├── common.ts       # ApiResponse<T>, PaginatedResponse<T>, PaginationMeta, UUID, ISODateString
├── user.ts         # User, UserRole, Organization, OrganizationMember, AuthSession
├── gameserver.ts   # GameServer, GameServerStatus, GameServerRegion, GameServerPlan, Deployment
└── billing.ts      # BillingPlan, Subscription, Invoice, UsageSummary
```

## Usage

```ts
import type {
  GameServer,
  GameServerStatus,
  BillingPlan,
  User,
  PaginatedResponse,
} from "@kleff/shared-types";
```

## Domain overview

### Common
- `ApiResponse<T>` — standard `{ data, message }` envelope
- `PaginatedResponse<T>` — `{ data[], pagination }` for list endpoints
- `UUID` / `ISODateString` — branded string aliases

### Users & Identity
- `User` — platform user (id, email, displayName, role)
- `UserRole` — `owner | admin | member | billing | viewer`
- `Organization` — tenant/org owning resources
- `OrganizationMember` — user ↔ org join with role

### Game Servers
- `GameServer` — full server record (name, status, region, plan, players, resources)
- `GameServerStatus` — `running | stopped | starting | stopping | restarting | provisioning | crashed | error`
- `GameServerRegion` — 8 AWS-style regions with `REGION_LABELS` display map
- `GameServerPlan` — `free | starter | pro | business | enterprise` with vCPU/RAM/storage specs
- `GameServerResources` — live CPU/memory/disk/network metrics
- `Deployment` — versioned deployment record with status

### Billing
- `BillingPlan` — pricing tier with features, limits, and support level
- `Subscription` — org's active plan with billing interval and period
- `Invoice` — itemised invoice with line items, tax, and status
- `UsageSummary` — period usage rollup (server-hours, bandwidth, storage, estimated cost)
