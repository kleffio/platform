# Dashboard

React frontend for the Kleff platform. Provides the main user interface for managing game servers, billing, and account settings.

## Tech Stack

- **Framework:** React 19 + Vite 7
- **Language:** TypeScript 5.9
- **Styling:** Tailwind CSS 4 + shadcn/ui + Radix UI
- **Routing:** React Router 7
- **Auth:** OIDC via `react-oidc-context`
- **Data Fetching:** Axios + TanStack Query 5
- **Tables:** TanStack Table 8
- **Charts:** Recharts 2

## Project Structure

```
src/
├── app/            # App-level wiring (router, guards, interceptors, providers, styles)
│   ├── guards/     # AuthGuard, GuestGuard
│   ├── interceptors/ # HTTP interceptors (auth 401/403 handling)
│   ├── providers/  # QueryClient, ThemeProvider, ToastProvider
│   ├── router/     # Route definitions and ROUTES constants
│   └── styles/     # Global CSS entry point
├── features/       # Self-contained domain slices
│   └── auth/       # OIDC auth — config, context, hooks, providers, types
├── layouts/        # Page shell components (sidebar, topbar)
├── pages/          # Route-level page components
│   ├── dashboard/  # Main dashboard page (authenticated)
│   └── components/ # Component showcase (guest)
└── shared/         # Cross-feature utilities (no domain knowledge)
    ├── api/        # Axios client & API helpers
    ├── hooks/      # Reusable React hooks
    └── utils/      # Pure utility functions
```

UI primitives live in `@kleff/ui` (the shared package), not inside this app.

## Getting Started

```bash
# From the repo root
pnpm install

# Copy environment variables
cp .env.example .env

# Start dev server
pnpm --filter dashboard dev
```

The dev server runs at `http://localhost:5173` by default.

## Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```env
VITE_OIDC_AUTHORITY=    # OIDC provider base URL (e.g. https://auth.example.com/realms/kleff)
VITE_OIDC_CLIENT_ID=    # OIDC client ID registered with the provider
VITE_API_BASE_URL=      # Backend API base URL (e.g. http://localhost:8080)
```

> **Note:** All `VITE_` variables are baked into the bundle at build time. For Docker/production builds, pass them as build arguments.

## Scripts

| Command | Description |
|---|---|
| `pnpm dev` | Start development server with HMR |
| `pnpm build` | Type-check and build for production |
| `pnpm preview` | Serve the production build locally |
| `pnpm lint` | Run ESLint |
| `pnpm format` | Format with Prettier |
| `pnpm typecheck` | Run TypeScript compiler without emitting |

## Docker

Build and run the dashboard container (from the repo root):

```bash
docker build \
  --build-arg VITE_OIDC_AUTHORITY=https://auth.example.com/realms/kleff \
  --build-arg VITE_OIDC_CLIENT_ID=kleff-dashboard \
  --build-arg VITE_API_BASE_URL=https://api.example.com \
  -f apps/dashboard/Dockerfile \
  -t kleff-dashboard .

docker run -p 3000:80 kleff-dashboard
```

The container serves the built static assets via nginx on port 80.

## Adding shadcn/ui Components

UI primitives belong in `packages/ui`, not in this app. Add new components there:

```bash
# From the repo root
cd packages/ui
pnpm dlx shadcn@latest add button
```

Then export the new component from `packages/ui/src/index.ts`.
