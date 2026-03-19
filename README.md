# Kleff Platform

Monorepo for the Kleff platform — a full-stack application combining a React dashboard with Go microservices.

## Repository Structure

```
platform/
├── apps/
│   └── dashboard/          # React + Vite frontend
├── packages/
│   ├── ui/                 # Shared React component library
│   ├── shared-types/       # Shared TypeScript type definitions
│   └── go-common/          # Shared Go utilities for backend services
├── services/
│   ├── billing-service/    # Billing & subscription management
│   ├── gameserver-service/ # Game server lifecycle management
│   └── identity-service/   # Authentication & user identity
├── infra/
│   ├── kubernetes/         # Kubernetes manifests
│   ├── helm/               # Helm charts
│   └── argocd/             # ArgoCD application definitions
├── docker-compose.yml      # Production compose
├── docker-compose.dev.yml  # Development compose
└── .github/                # CI/CD workflows
```

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 19, Vite 7, TypeScript, Tailwind CSS 4, shadcn/ui |
| Backend | Go 1.23, Hexagonal Architecture |
| Auth | OIDC (OpenID Connect) |
| Package Manager | pnpm (Node), Go modules |
| Infrastructure | Kubernetes, Helm, ArgoCD |
| CI/CD | GitHub Actions |

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) 22+
- [pnpm](https://pnpm.io/) (enabled via `corepack enable`)
- [Go](https://go.dev/) 1.23+
- [Docker](https://www.docker.com/) & Docker Compose

### Local Development (without Docker)

```bash
# Install Node dependencies
pnpm install

# Start the dashboard
pnpm --filter dashboard dev

# Start a Go service
cd services/identity-service
go run ./cmd/api
```

### Running with Docker Compose

```bash
# Production build
docker compose up --build

# Development mode (with hot reload where supported)
docker compose -f docker-compose.dev.yml up --build
```

## Apps & Services

| Name | Directory | Port | Description |
|---|---|---|---|
| dashboard | `apps/dashboard/` | 3000 | React web application |
| identity-service | `services/identity-service/` | 8083 | Auth & user identity |
| billing-service | `services/billing-service/` | 8081 | Billing & subscriptions |
| gameserver-service | `services/gameserver-service/` | 8082 | Game server management |

## Packages

| Name | Directory | Description |
|---|---|---|
| `@kleff/ui` | `packages/ui/` | Shared React component library |
| `@kleff/shared-types` | `packages/shared-types/` | Shared TypeScript types |
| `go-common` | `packages/go-common/` | Shared Go utilities |

## Code Ownership

| Path | Owner(s) |
|---|---|
| `/apps` | @isaacwallace123 |
| Everything else | @isaacwallace123, @viktorkuts, @Reid910 |
