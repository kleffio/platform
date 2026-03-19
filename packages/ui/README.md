# @kleff/ui

Domain-specific React components shared across Kleff apps. Built with Tailwind CSS — consumes `@kleff/shared-types` for prop types.

## Components

| Component | Description |
|---|---|
| `StatusBadge` | Coloured badge for `GameServerStatus` (green running, amber transitioning, red error) |
| `PlanBadge` | Tier badge for `GameServerPlanTier` (zinc free → gold pro → gradient enterprise) |
| `MetricCard` | Stat card with label, value, icon, and optional up/down delta |
| `ServerCard` | Full game server summary card with resource bars and start/stop actions |
| `RegionBadge` | Region pill with flag emoji and short or full region name |

## Usage

```tsx
import { StatusBadge, PlanBadge, MetricCard, ServerCard, RegionBadge } from "@kleff/ui";
import type { GameServer } from "@kleff/shared-types";

// Status colour-coded badge
<StatusBadge status="running" />
<StatusBadge status="provisioning" showDot />

// Plan tier badge
<PlanBadge tier="pro" />

// Metric stat card
<MetricCard
  label="Active Nodes"
  value={42}
  delta={{ value: "+3 today", direction: "up" }}
/>

// Full server card with actions
<ServerCard
  server={server}
  onStart={(id) => startServer(id)}
  onStop={(id) => stopServer(id)}
  onSelect={(server) => navigate(`/servers/${server.id}`)}
/>

// Region with flag
<RegionBadge region="eu-central-1" short />
```

## Design tokens

Components use standard Tailwind utilities (`zinc-*`, `emerald-*`, `amber-*`, `red-*`). They automatically respect the Kleff dark theme when consumed inside the dashboard.
