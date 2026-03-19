import type { GameServer } from "@kleff/shared-types";
import { REGION_LABELS } from "@kleff/shared-types";
import { StatusBadge } from "./StatusBadge.js";
import { PlanBadge } from "./PlanBadge.js";

interface ServerCardProps {
  server: GameServer;
  onStart?: (id: string) => void;
  onStop?: (id: string) => void;
  onSelect?: (server: GameServer) => void;
  className?: string;
}

function ResourceBar({ label, percent }: { label: string; percent: number }) {
  const color =
    percent >= 90 ? "bg-red-500" : percent >= 70 ? "bg-amber-400" : "bg-emerald-500";

  return (
    <div className="flex flex-col gap-1">
      <div className="flex justify-between text-xs text-zinc-500">
        <span>{label}</span>
        <span>{percent}%</span>
      </div>
      <div className="h-1 w-full rounded-full bg-zinc-800">
        <div
          className={`h-1 rounded-full transition-all ${color}`}
          style={{ width: `${Math.min(percent, 100)}%` }}
        />
      </div>
    </div>
  );
}

export function ServerCard({ server, onStart, onStop, onSelect, className = "" }: ServerCardProps) {
  const isRunning = server.status === "running";
  const isTransitioning =
    server.status === "starting" ||
    server.status === "stopping" ||
    server.status === "restarting" ||
    server.status === "provisioning";

  return (
    <div
      className={`rounded-xl border border-zinc-800 bg-zinc-900 p-4 flex flex-col gap-4 hover:border-zinc-700 transition-colors ${onSelect ? "cursor-pointer" : ""} ${className}`}
      onClick={() => onSelect?.(server)}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1 min-w-0">
          <span className="text-sm font-semibold text-zinc-50 truncate">{server.name}</span>
          <span className="text-xs text-zinc-500">{server.gameType}</span>
        </div>
        <StatusBadge status={server.status} />
      </div>

      {/* Meta */}
      <div className="flex flex-wrap gap-2 text-xs text-zinc-400">
        <span className="flex items-center gap-1">
          <span className="text-zinc-600">⬡</span>
          {REGION_LABELS[server.region] ?? server.region}
        </span>
        <span className="text-zinc-700">·</span>
        <span className="flex items-center gap-1">
          <span className="text-zinc-600">👥</span>
          {server.currentPlayers} / {server.plan.maxPlayers}
        </span>
        {server.ipAddress && (
          <>
            <span className="text-zinc-700">·</span>
            <span className="font-mono">
              {server.ipAddress}:{server.port ?? "—"}
            </span>
          </>
        )}
      </div>

      {/* Resource bars */}
      {server.resources && (
        <div className="flex flex-col gap-2">
          <ResourceBar label="CPU" percent={server.resources.cpuPercent} />
          <ResourceBar label="Memory" percent={server.resources.memoryPercent} />
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between gap-2 pt-1 border-t border-zinc-800">
        <PlanBadge tier={server.plan.tier} />

        <div className="flex items-center gap-2">
          {isRunning && onStop && (
            <button
              onClick={(e) => { e.stopPropagation(); onStop(server.id); }}
              disabled={isTransitioning}
              className="rounded-md px-2.5 py-1 text-xs font-medium text-zinc-300 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-40 transition-colors"
            >
              Stop
            </button>
          )}
          {!isRunning && onStart && (
            <button
              onClick={(e) => { e.stopPropagation(); onStart(server.id); }}
              disabled={isTransitioning}
              className="rounded-md px-2.5 py-1 text-xs font-medium text-zinc-900 bg-amber-400 hover:bg-amber-300 disabled:opacity-40 transition-colors"
            >
              Start
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
