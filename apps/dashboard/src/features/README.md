# src/features/

Domain feature slices. Each subfolder is a self-contained vertical slice that owns everything for one area of the product.

## Current features

| Folder | Domain |
|---|---|
| `auth/` | OIDC authentication, session management, role-based authorization |
| `servers/` | _(scaffolded)_ Game server CRUD, status polling, start/stop actions |
| `billing/` | _(scaffolded)_ Subscription management, invoice history, plan upgrades |
| `settings/` | _(scaffolded)_ User profile, organisation settings, API keys |

## Internal structure (per feature)

```
features/<name>/
├── index.ts          # Public API — only import from here, never from subfolders directly
├── config/           # Feature-specific configuration (env vars, constants)
├── context/          # React context definitions
├── hooks/            # Custom hooks
├── providers/        # React providers scoped to this feature
├── types/            # TypeScript types/interfaces
├── components/       # UI components used only within this feature
└── api/              # API call functions for this feature's endpoints
```

## Rules

- **Export through `index.ts`** — consumers import `from "@/features/auth"`, not from deep paths.
- **No cross-feature imports** — features must not import from each other. Share via `shared/` or `packages/`.
- **No page-level logic** — a feature exports building blocks; pages assemble them.
