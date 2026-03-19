import type { ReactNode } from "react";

interface MetricCardProps {
  label: string;
  value: string | number;
  icon?: ReactNode;
  delta?: {
    value: string;
    direction: "up" | "down" | "neutral";
  };
  className?: string;
}

const DELTA_STYLES = {
  up: "text-emerald-400",
  down: "text-red-400",
  neutral: "text-zinc-400",
};

const DELTA_ARROWS = {
  up: "↑",
  down: "↓",
  neutral: "→",
};

export function MetricCard({ label, value, icon, delta, className = "" }: MetricCardProps) {
  return (
    <div
      className={`rounded-xl border border-zinc-800 bg-zinc-900 p-4 flex flex-col gap-3 ${className}`}
    >
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-zinc-400 uppercase tracking-wider">
          {label}
        </span>
        {icon && (
          <span className="text-zinc-500">{icon}</span>
        )}
      </div>

      <div className="flex items-end justify-between gap-2">
        <span className="text-2xl font-semibold text-zinc-50 tabular-nums">
          {value}
        </span>

        {delta && (
          <span className={`text-xs font-medium ${DELTA_STYLES[delta.direction]}`}>
            {DELTA_ARROWS[delta.direction]} {delta.value}
          </span>
        )}
      </div>
    </div>
  );
}
